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
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Env struct {
	Vars   []string `toml:"vars"`
	Scheme string   `toml:"scheme"`
}

type Config struct {
	DefaultEnv string                       `toml:"default_env"`
	Envs       map[string]Env               `toml:"env"`
	Forms      map[string]map[string]string `toml:"form"`
}

func configLookup() (string, error) {
	var home = "./"

	if runtime.GOOS == "windows" {
		home = os.Getenv("UserProfile")
	} else {
		home = os.Getenv("HOME")
	}

	paths := [2]string{
		fmt.Sprintf("%s/.config/pgg/config", home),
		fmt.Sprintf("%s/.pgg/config", home),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", errors.New("error: config file not found")
}

func loadConfig(path string) (Config, error) {
	var cfg Config

	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
