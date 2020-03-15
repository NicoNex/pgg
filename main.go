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
	"os"
	"io"
	"fmt"
	"flag"
	"bytes"
	"regexp"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"mime/multipart"

	. "github.com/logrusorgru/aurora"
	term "golang.org/x/crypto/ssh/terminal"
)

const PROGRAM_NAME = "pgg"

type variable struct {
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

func parseCfgVars(rawVars []string, ch chan variable) {
	var re = regexp.MustCompile(` *[=|:] *`)

	for k, s := range rawVars {
		tokens := re.Split(s, 2)

		if len(tokens) != 2 {
			fmt.Printf("error parsing variable #%d\n", k)
			continue
		}

		ch <- variable{
			Key: fmt.Sprintf("{{%s}}", tokens[0]),
			Val: tokens[1],
		}
	}
	close(ch)
}

func replaceVars(str string, in chan variable, out chan string) {
	for v := range in {
		str = strings.ReplaceAll(str, v.Key, v.Val)
	}

	out <- str
}

func formatUrl(rawUrl string, env Env) string {
	var url string
	var varch = make(chan variable)
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

func getFileRequest(url, fpath, fieldname string) *http.Request {
	var body *bytes.Buffer

	file, err := os.Open(fpath)
	check(err)
	defer file.Close()

	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldname, filepath.Base(file.Name()))
	check(err)

	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest("POST", url, body)
	check(err)

	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request
}

func doRequest(request *http.Request) (string, string) {
	var client = &http.Client{}

	response, err := client.Do(request)
	check(err)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	check(err)

	return string(body), response.Status
}

func populateForm(form *url.Values, data map[string]string) {
	for key, value := range data {
		form.Add(key, value)
	}
}

func usage() {
	var msg = `pgg - Post from the Get-Go
Pgg is a tool that allows you to make http request.

When starting pgg looks for configuration files in the following order:
    1. ~/.config/pgg/config
    2. ~/.pgg/config

SYNOPSIS
    pgg [options] http://foobar.org

OPTIONS
    -m
        Specify the request method. (default GET)
    -e
        Specify the environment to use.
    -c
        Specify an alternative config file.
    -f
        Specify a file to upload.
    -fo
        Specify the form to use.
    -h, -help
        Prints this help message.`
	fmt.Println(msg)
}

func main() {
	var env Env
	var err error
	var cfg Config
	var fmtUrl string
	var method string // the request method
	var envName string // the environment name
	var cfgPath string // path to the config file
	var fileFlag string // path to the file to upload
	var formFlag string // form to use in the request
	var form url.Values
	var request *http.Request

	// parse the argument and gets the flags values.
	flag.StringVar(&method, "method", "GET", "Request method")
	flag.StringVar(&method, "m", "GET", "Request method")
	flag.StringVar(&envName, "e", "", "Environment to use")
	flag.StringVar(&cfgPath, "c", "", "Config file")
	flag.StringVar(&fileFlag, "f", "", "Path to the file to upload")
	flag.StringVar(&formFlag, "fo", "", "Form to use")
	// flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
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
	fmtUrl = formatUrl(flag.Arg(flag.NArg()-1), env)

	if formFlag != "" {
		frm, ok := cfg.Forms[formFlag]
		if !ok {
			die(fmt.Sprintf("error: cannot find form %s.", formFlag))
		}
		populateForm(&form, frm)
	}

	// handle the file upload
	if fileFlag != "" {
		tokens := strings.Split(fileFlag, "=")
		request = getFileRequest(fmtUrl, tokens[1], tokens[0])
	} else {
		request, err = http.NewRequest(strings.ToUpper(method), fmtUrl, strings.NewReader(form.Encode()))
		check(err)
	}

	body, status := doRequest(request)
	if isatty() {
		status := BrightMagenta(fmt.Sprintf("Status: %s", status))
		url := BrightGreen(fmtUrl)

		if body != "" {
			fmt.Printf("%s\n%s\n%s\n", body, status, url)
		} else {
			fmt.Printf("%s\n%s\n", status, url)
		}
	} else {
		fmt.Println(body)
	}
}
