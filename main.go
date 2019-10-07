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

Usage: pgg [OPTIONS] reqUrl

Options:
    -m     Specify the request method.
    -e     Specify the environment to use.
    -c     Specify an alternative config file.

    -h     Prints this help message.
    --help Prints options details.

Example:
    pgg [-m POST] [-e foo] [-c bar.conf] http://foobar.org
`

	fmt.Println(msg)
}

func main() {
	var env Env
	var cfg Config
	var reqUrl string
	var vars []string
	var showHelp *bool
	var reqMeth *string // the request type
	var envName *string // the environment name
	var cfgPath *string // path to the config file
	var dfltPath string
	var lgr *log.Logger

	dfltPath = fmt.Sprintf("%s/.config/%s/config", os.Getenv("HOME"), PROGRAM_NAME)
	lgr = log.New(os.Stderr, "", 0)

	// parse the argument and gets the flags values.
	reqMeth = flag.String("m", "GET", "Request method")
	envName = flag.String("e", "default", "Environment to use")
	cfgPath = flag.String("c", dfltPath, "Config file")
	showHelp = flag.Bool("h", false, "Show help and usage")
	flag.Parse()

	if *showHelp || len(os.Args) <= 1 {
		usage()
		return
	}

	var err error
	cfg, err = loadConfig(*cfgPath)
	if err != nil {
		lgr.Fatal(Bold(BrightRed(err)))
	}

	var ok bool
	if env, ok = cfg.Envs[*envName]; !ok {
		lgr.Fatal(Bold(BrightRed("error: no default environment set and none specified.")))
	}
	vars = escapeVars(env.Vars)

	// replace variables in reqUrl with escaped ones
	r := strings.NewReplacer(vars...)
	reqUrl = r.Replace(os.Args[len(os.Args)-1])
	fmt.Println(Bold(BrightBlue(reqUrl)), "\n")

	// make the request to the reqUrl
	form := url.Values{}
	request, err := http.NewRequest(strings.ToUpper(*reqMeth), reqUrl, strings.NewReader(form.Encode()))
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

	fmt.Printf("%s\nStatus: %s\n", string(content), Bold(BrightMagenta(response.Status)))
}
