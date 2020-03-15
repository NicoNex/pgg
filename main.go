/*
 * Pgg
 * Copyright (C) 2019  Nicolò Santamaria
 *
 * Pgg is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pgg is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/logrusorgru/aurora"
	term "golang.org/x/crypto/ssh/terminal"
)

const PROGRAM_NAME = "pgg"

type pair struct {
	Key string
	Val string
}

func isatty() bool {
	fd := os.Stdout.Fd()
	return term.IsTerminal(int(fd))
}

func die(msg interface{}) {
	fmt.Println(BrightRed(msg))
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		die(err)
	}
}

func parseCfgVars(rawVars []string, ch chan pair) {
	var re = regexp.MustCompile(` *[=|:] *`)

	for k, s := range rawVars {
		tokens := re.Split(s, 2)

		if len(tokens) != 2 {
			fmt.Printf("error parsing pair #%d\n", k)
			continue
		}

		ch <- pair{
			Key: fmt.Sprintf("{{%s}}", tokens[0]),
			Val: tokens[1],
		}
	}
	close(ch)
}

func replaceVars(str string, varch chan pair, out chan string) {
	for v := range varch {
		str = strings.ReplaceAll(str, v.Key, v.Val)
	}

	out <- str
}

func formatUrl(rawUrl string, env Env) string {
	var url string
	var varch = make(chan pair)
	var urlch = make(chan string, 1)

	go replaceVars(rawUrl, varch, urlch)
	go parseCfgVars(env.Vars, varch)

	url = <-urlch
	close(urlch)
	if ok, _ := regexp.MatchString(`[a-z]+:\/\/`, url); !ok {
		url = fmt.Sprintf("%s%s", env.Scheme, url)
	}

	return url
}

func sendRequest(request *http.Request) (string, string) {
	var client = &http.Client{}

	response, err := client.Do(request)
	check(err)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	check(err)

	return string(body), response.Status
}

func parseForm(raw string, outch chan pair) {
	re := regexp.MustCompile(`\s+`)
	for _, entry := range re.Split(raw, -1) {
		toks := strings.Split(entry, "=")
		if len(toks) != 2 {
			continue
		}
		outch <- pair{
			Key: toks[0],
			Val: toks[1],
		}
	}
	close(outch)
}

func populateForm(r *http.Request, rawform string) url.Values {
	var form url.Values
	var data = make(chan pair)

	go parseForm(rawform, data)
	for p := range data {
		form.Add(p.Key, p.Val)
	}

	return form
}

func getFileContent(w io.Writer, fname string) {
	file, err := os.Open(fname)
	check(err)
	defer file.Close()

	_, err = io.Copy(w, file)
	check(err)
}

func getBody(form string) bytes.Buffer {
	var body bytes.Buffer
	var datach = make(chan pair)

	go parseForm(form, datach)
	writer := multipart.NewWriter(&body)

	for p := range datach {
		switch p.Val[0] {
		case '@':
			// TODO: fix index out of range here
			fname := filepath.Base(p.Val[1:])
			fwriter, err := writer.CreateFormFile(p.Key, fname)
			check(err)
			getFileContent(fwriter, fname)

		default:
			err := writer.WriteField(p.Key, p.Val)
			check(err)
		}
	}
	writer.Close()
	return body
}

func getBodyCfgForm(data map[string]string) bytes.Buffer {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	for key, value := range data {
		switch value[0] {
		case '@':
			fname := value[1:]
			fwriter, err := writer.CreateFormFile(key, fname)
			check(err)
			getFileContent(fwriter, fname)

		default:
			err := writer.WriteField(key, value)
			check(err)
		}
	}
	writer.Close()
	return body
}

func parseHeader(rawHeader string) pair {
	re := regexp.MustCompile(` *[=|:] *`)
	data := re.Split(rawHeader, 2)
	if len(data) != 2 {
		die("error: invaild header")
	}

	return pair{
		Key: data[0],
		Val: data[1],
	}
}

func main() {
	var env Env
	var err error
	var cfg Config
	var fmtUrl string
	var method string   // the request method
	var envName string  // the environment name
	var cfgPath string  // path to the config file
	var formFlag string // the form to use in the request
	var cfgForm string  // form to use in the request
	var headerFlag string
	var dataFlag string
	var body bytes.Buffer
	var request *http.Request

	// parse the argument and gets the flags values.
	flag.StringVar(&method, "m", "GET", "Request method.")
	flag.StringVar(&envName, "e", "", "Environment to use.")
	flag.StringVar(&cfgPath, "c", "", "Path to the config file.")
	flag.StringVar(&formFlag, "f", "", "Form key-value pairs.")
	flag.StringVar(&cfgForm, "fn", "", "Form name from the config file.")
	flag.StringVar(&headerFlag, "H", "", "Header key-value pair to send.")
	flag.StringVar(&dataFlag, "d", "", "Raw data to send.")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	// If no cfg file specified in argument look in the default paths.
	if cfgPath == "" {
		cfgPath, err = configLookup()
		check(err)
	}

	cfg, err = loadConfig(cfgPath)
	check(err)

	// The flag value overrides the default.
	if envName == "" {
		envName = cfg.DefaultEnv
	}

	var ok bool
	if env, ok = cfg.Envs[envName]; !ok {
		die(fmt.Sprintf("error: cannot find environment %s.", envName))
	}
	fmtUrl = formatUrl(flag.Arg(0), env)

	// Handle the form.
	if dataFlag != "" {
		body = *bytes.NewBuffer([]byte(dataFlag))
		method = "POST"
	} else if formFlag != "" {
		body = getBody(formFlag)
		method = "POST"
	} else if cfgForm != "" {
		data, ok := cfg.Forms[cfgForm]
		if !ok {
			die(fmt.Sprintf("error: cannot find form %s.", cfgForm))
		}
		body = getBodyCfgForm(data)
		method = "POST"
	}

	request, err = http.NewRequest(strings.ToUpper(method), fmtUrl, &body)
	check(err)

	// Handle the header flag.
	// TODO: fix the header that doesn't get added.
	if headerFlag != "" {
		h := parseHeader(headerFlag)
		request.Header.Add(h.Key, h.Val)
	}

	response, status := sendRequest(request)
	if isatty() {
		status := BrightMagenta(fmt.Sprintf("Status: %s", status))
		url := BrightGreen(fmtUrl)

		if response != "" {
			fmt.Printf("%s\n%s\n%s\n", response, status, url)
		} else {
			fmt.Printf("%s\n%s\n", status, url)
		}
	} else {
		fmt.Println(response)
	}
}
