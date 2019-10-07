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
	var reqUrl string
	var vars []string
	var showHelp bool
	var reqMeth string // the request method
	var envName string // the environment name
	var cfgPath string // path to the config file
	var dfltPath string
	var lgr *log.Logger

	dfltPath = fmt.Sprintf("%s/.config/%s/config", os.Getenv("HOME"), PROGRAM_NAME)
	lgr = log.New(os.Stderr, "", 0)

	// parse the argument and gets the flags values.
	flag.StringVar(&reqMeth, "method", "GET", "Request method")
	flag.StringVar(&reqMeth, "m", "GET", "Request method (shorthand)")
	flag.StringVar(&envName, "env", "", "Environment to use")
	flag.StringVar(&envName, "e", "", "Environment to use")
	flag.StringVar(&cfgPath, "cfg", dfltPath, "Config file")
	flag.StringVar(&cfgPath, "c", dfltPath, "Config file")
	flag.BoolVar(&showHelp, "h", false, "Show help and usage")
	flag.Parse()

	if showHelp || len(os.Args) <= 1 {
		usage()
		return
	}

	var err error
	cfg, err = loadConfig(cfgPath)
	if err != nil {
		lgr.Fatal(Bold(BrightRed(err)))
	}

	var ok bool
	if envName == "" {
		if env, ok = cfg.Envs[cfg.DefaultEnv]; !ok {
			msg := fmt.Sprintf("error: cannot find environment %s and none specified.", cfg.DefaultEnv)
			lgr.Fatal(Bold(BrightRed(msg)))
		}
	} else {
		if env, ok = cfg.Envs[envName]; !ok {
			msg := fmt.Sprintf("error: cannot find environment %s.", envName)
			lgr.Fatal(Bold(BrightRed(msg)))
		}
	}

	vars = escapeVars(env.Vars)

	// replace variables in reqUrl with escaped ones
	r := strings.NewReplacer(vars...)
	reqUrl = r.Replace(os.Args[len(os.Args)-1])
	fmt.Println(Bold(BrightGreen(reqUrl)), "\n")

	// make the request to the reqUrl
	form := url.Values{}
	request, err := http.NewRequest(strings.ToUpper(reqMeth), reqUrl, strings.NewReader(form.Encode()))
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

	fmt.Printf("%s\n\nStatus: %s\n", string(content), Bold(BrightMagenta(response.Status)))
}
