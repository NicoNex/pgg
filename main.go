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
	"log"
	"flag"
	"errors"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"

	. "github.com/logrusorgru/aurora"
)

const PROGRAM_NAME = "pgg"

func escapeVars(rawVars []string) []string {
	var esc []string

	for _, s := range rawVars {
		tok := strings.Split(s, "=")
		esc = append(esc, fmt.Sprintf("{{%s}}", tok[0]), tok[1])
	}

	return esc
}

// Replace the variables in the url with their values defined in the config.
func formatUrl(rawUrl string, rawVars []string) string {
	var vars []string

	vars = escapeVars(rawVars)
	r := strings.NewReplacer(vars...)
	return r.Replace(rawUrl)
}

func configLookup() (string, error) {
	home := os.Getenv("HOME")
	paths := [2]string{
		fmt.Sprintf("%s/.config/%s/config", home, PROGRAM_NAME),
		fmt.Sprintf("%s/.%s/config", home, PROGRAM_NAME),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); os.IsExist(err) {
			return p, nil
		}
	}

	return "", errors.New("error: config file not found")
}

func usage() {
	var msg = `pgg - Post from the Get-Go
Pgg is a tool that allows you to make http request.

Pgg looks for the config file in the default location:
    $HOME/.config/pgg/config

SYNOPSIS
    pgg [options] http://foobar.org

OPTIONS
    -m, -method
        Specify the request method. (default GET)

    -e, -env
        Specify the environment to use.

    -c, -cfg
        Specify an alternative config file.

    -h
        Prints this help message.

    --help
        Prints options details.
`

	fmt.Println(msg)
}

func main() {
	var env Env
	var cfg Config
	var fmtUrl string
	var showHelp bool
	var reqMeth string // the request method
	var envName string // the environment name
	var cfgPath string // path to the config file
	var lgr *log.Logger

	lgr = log.New(os.Stderr, "", 0)

	// parse the argument and gets the flags values.
	flag.StringVar(&reqMeth, "method", "GET", "Request method")
	flag.StringVar(&reqMeth, "m", "GET", "Request method (shorthand)")
	flag.StringVar(&envName, "env", "", "Environment to use")
	flag.StringVar(&envName, "e", "", "Environment to use")
	flag.StringVar(&cfgPath, "cfg", "", "Config file")
	flag.StringVar(&cfgPath, "c", "", "Config file")
	flag.BoolVar(&showHelp, "h", false, "Show help and usage")
	flag.Parse()

	if showHelp || len(os.Args) < 2 {
		usage()
		return
	}

	// If no cfg file specified in argument look in the default paths.
	if cfgPath == "" {
		p, err := configLookup()
		if err != nil {
			lgr.Fatal(Bold(BrightRed(err)))
		}
		cfgPath = p
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		lgr.Fatal(Bold(BrightRed(err)))
	}

	// The flag value overrides the default.
	var ok bool
	if envName == "" {
		if env, ok = cfg.Envs[cfg.DefaultEnv]; !ok {
			msg := fmt.Sprintf("error: cannot find environment %s and none specified", envName)
			lgr.Fatal(Bold(BrightRed(msg)))
		}
	} else {
		if env, ok = cfg.Envs[envName]; !ok {
			msg := fmt.Sprintf("error: cannot find environment %s.", envName)
			lgr.Fatal(Bold(BrightRed(msg)))
		}
	}

	rawUrl := os.Args[len(os.Args)-1]
	fmtUrl = formatUrl(rawUrl, env.Vars)
	fmt.Println(Bold(BrightGreen(fmtUrl)))

	// make the request to the reqUrl
	form := url.Values{}
	request, err := http.NewRequest(strings.ToUpper(reqMeth), fmtUrl, strings.NewReader(form.Encode()))
	if err != nil {
		lgr.Fatal(Bold(BrightRed(err)))
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		lgr.Fatal(Bold(BrightRed(err)))
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		lgr.Fatal(Bold(BrightRed(err)))
	}

	body := string(content)
	if body != "" {
		fmt.Printf("%s\n\nStatus: %s\n", body, Bold(BrightMagenta(response.Status)))
	} else {
		fmt.Printf("Status: %s\n", Bold(BrightMagenta(response.Status)))
	}
}
