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
	"fmt"
	"flag"
	"regexp"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"

	. "github.com/logrusorgru/aurora"
	term "golang.org/x/crypto/ssh/terminal"
)

const PROGRAM_NAME = "pgg"

// TODO: find a way to do this in pure go
func isatty() bool {
	fd := os.Stdout.Fd()
	return term.IsTerminal(int(fd))
}

func escapeVars(rawVars []string) []string {
	var esc []string

	for _, s := range rawVars {
		tok := strings.Split(s, "=")
		esc = append(esc, fmt.Sprintf("{{%s}}", tok[0]), tok[1])
	}

	return esc
}

// Replace the variables in the url with their values defined in the config.
func formatUrl(rawUrl string, env Env) string {
	var vars []string

	if ok, _ := regexp.MatchString(`[a-z]+:\/\/`, rawUrl); !ok {
		rawUrl = fmt.Sprintf("%s%s", env.Scheme, rawUrl)
	}

	vars = escapeVars(env.Vars)
	r := strings.NewReplacer(vars...)
	return r.Replace(rawUrl)
}

func die(msg interface{}) {
	fmt.Println(BrightRed(msg))
	os.Exit(1)
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
    -m, -method
        Specify the request method. (default GET)
    -e, -env
        Specify the environment to use.
    -c, -cfg
        Specify an alternative config file.
	-f, -file
		Specify a file to upload.
    -h
        Prints this help message.
    --help
        Prints options details.`
	fmt.Println(msg)
}


func main() {
	var env Env
	var cfg Config
	var showHelp bool
	var fmtUrl string
	var method string // the request method
	var envName string // the environment name
	var cfgPath string // path to the config file
	var filepath string // path to the file to upload

	// parse the argument and gets the flags values.
	flag.StringVar(&method, "method", "GET", "Request method")
	flag.StringVar(&method, "m", "GET", "Request method (shorthand)")
	flag.StringVar(&envName, "env", "", "Environment to use")
	flag.StringVar(&envName, "e", "", "Environment to use")
	flag.StringVar(&cfgPath, "cfg", "", "Config file")
	flag.StringVar(&cfgPath, "c", "", "Config file")
	flag.StringVar(&filepath, "file", "", "Path to the file to upload")
	flag.StringVar(&filepath, "f", "", "Path to the file to upload")
	flag.BoolVar(&showHelp, "h", false, "Show help and usage")
	flag.Parse()

	if showHelp || len(os.Args) < 2 {
		usage()
		return
	}

	// If no cfg file specified in argument look in the default paths.
	if cfgPath == "" {
		var err error
		cfgPath, err = configLookup()
		if err != nil {
			die(err)
		}
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		die(err)
	}

	// The flag value overrides the default.
	var ok bool
	if envName == "" {
		if env, ok = cfg.Envs[cfg.DefaultEnv]; !ok {
			msg := fmt.Sprintf("error: cannot find environment %s and none specified", envName)
			die(msg)
		}
	} else {
		if env, ok = cfg.Envs[envName]; !ok {
			msg := fmt.Sprintf("error: cannot find environment %s.", envName)
			die(msg)
		}
	}

	rawUrl := os.Args[len(os.Args)-1]
	fmtUrl = formatUrl(rawUrl, env)

	// make the request to the reqUrl
	form := url.Values{}
	request, err := http.NewRequest(strings.ToUpper(method), fmtUrl, strings.NewReader(form.Encode()))
	if err != nil {
		die(err)
	}

	// handle the file upload
	if filename != "" {
		method = "POST"
		file, handler, err := request.FormFile(filepath)
		if err != nil {
			die(err)
		}
		defer file.Close()

		f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			die(err)
		}
		defer f.Close()
		io.Copy(f, file)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		die(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		die(err)
	}

	body := string(content)
	if isatty() {
		status := BrightMagenta(fmt.Sprintf("Status: %s", response.Status))
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
