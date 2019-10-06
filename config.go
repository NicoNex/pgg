package main

import (
	"log"
	"github.com/BurntSushi/toml"
)

type Env struct {
	Vars []string `toml:"vars"`
}

type Config struct {
	Envs map[string]Env `toml:"env"`
}

func loadConfig(path string) Config {
	var cfg Config

	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
