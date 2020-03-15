package main

import "fmt"

func usage() {
	fmt.Println(`pgg - Post from the Get-Go
Pgg allows you to make http requests with ease.

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
        Prints this help message.`)
}
