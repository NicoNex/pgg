package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"

	"github.com/logrusorgru/aurora"
)

const PROGRAM_NAME = "rq"

func escapeVars(rawVars []string) []string {
	var esc []string

	for _, s := range rawVars {
		tok := strings.Split(s, "=")
		esc = append(esc, fmt.Sprintf("{{%s}}", tok[0]), tok[1])
	}

	return esc
}

func main() {
	var env Env
	var cfg Config
	var url string
	var vars []string
	var reqMeth *string // the request type
	var envName *string // the environment name
	var cfgPath *string // path to the config file
	var dfltPath string

	dfltPath = fmt.Sprintf("%s/.config/%s/config", os.Getenv("HOME"), PROGRAM_NAME)

	// parse the argument and gets the flags values.
	reqMeth = flag.String("m", "GET", "request method")
	envName = flag.String("e", "default", "environment")
	cfgPath = flag.String("f", dfltPath, "path to config file")
	flag.Parse()

	cfg = loadConfig(*cfgPath)
	env = cfg.Envs[*envName]
	vars = escapeVars(env.Vars)

	// replace variables in url with escaped ones
	r := strings.NewReplacer(vars...)
	url = r.Replace(os.Args[len(os.Args)-1])
	fmt.Println(aurora.BrightYellow(url), "\n")

	// make the request to the url
	request, err := http.NewRequest(strings.ToUpper(*reqMeth), url, &bytes.Buffer{})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(content))
}