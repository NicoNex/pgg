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
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"

	"github.com/logrusorgru/aurora"
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

Usage: pgg [OPTIONS] URL

Options:
    -m Specify the request method.
    -e Specify the environment to use.
    -f Specify the config file.
    -h Prints this help message.`

	fmt.Println(msg)
}

func main() {
	var env Env
	var cfg Config
	var url string
	var vars []string
	var showHelp *bool
	var reqMeth *string // the request type
	var envName *string // the environment name
	var cfgPath *string // path to the config file
	var dfltPath string

	dfltPath = fmt.Sprintf("%s/.config/%s/config", os.Getenv("HOME"), PROGRAM_NAME)

	// parse the argument and gets the flags values.
	reqMeth = flag.String("m", "GET", "request method")
	envName = flag.String("e", "default", "environment")
	cfgPath = flag.String("f", dfltPath, "path to config file")
	showHelp = flag.Bool("h", false, "show help")
	flag.Parse()

	if *showHelp || len(os.Args) <= 1 {
		usage()
		return
	}

	cfg = loadConfig(*cfgPath)
	var ok bool
	if env, ok = cfg.Envs[*envName]; !ok {
		fmt.Println(aurora.BrightRed("error: no default environment set and none specified."))
		return
	}
	vars = escapeVars(env.Vars)

	// replace variables in url with escaped ones
	r := strings.NewReplacer(vars...)
	url = r.Replace(os.Args[len(os.Args)-1])
	fmt.Println(aurora.BrightYellow(url), "\n")

	// make the request to the url
	request, err := http.NewRequest(strings.ToUpper(*reqMeth), url, &bytes.Buffer{})
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}