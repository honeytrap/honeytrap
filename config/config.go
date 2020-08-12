// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//Package config is the honeytrap configuration, it is set by the server.
package config

import (
	"regexp"

	"io"
	"os"

	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:config")

var elapso = regexp.MustCompile(`(\d+)(\w+)`)

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{module} ▶ %{level:.4s} %{id:03x} %{message}%{color:reset}",
)

// Config defines the central type where all configuration is umarhsalled to.
type Config struct {
	toml.MetaData

	Listener toml.Primitive `toml:"listener"`

	Web toml.Primitive `toml:"web"`

	Services  map[string]toml.Primitive `toml:"service"`
	Ports     []toml.Primitive          `toml:"port"`
	Directors map[string]toml.Primitive `toml:"director"`
	Channels  map[string]toml.Primitive `toml:"channel"`

	Filters []toml.Primitive `toml:"filter"`

	Logging []struct {
		Output string `toml:"output"`
		Level  string `toml:"level"`
	} `toml:"logging"`
}

// Default Config defines the default Config to be used to set default values.
var Default = Config{}

// Load attempts to load the giving toml configuration file.
func (c *Config) Load(r io.Reader) error {
	md, err := toml.DecodeReader(r, c)
	if err != nil {
		return err
	}
	c.MetaData = md

	if len(c.Logging) == 0 {
		fmt.Println("Warning: no logging backends configured. Add one to view log messages.")
	}
	var logBackends []logging.Backend
	for _, log := range c.Logging {
		var err error

		var output io.Writer

		switch log.Output {
		case "stdout":
			output = os.Stdout
		case "stderr":
			output = os.Stderr
		default:
			output, err = os.OpenFile(os.ExpandEnv(log.Output), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
		}

		if err != nil {
			panic(err)
		}

		backend := logging.NewLogBackend(output, "", 0)
		backendFormatter := logging.NewBackendFormatter(backend, format)
		backendLeveled := logging.AddModuleLevel(backendFormatter)

		level, err := logging.LogLevel(log.Level)
		if err != nil {
			panic(err)
		}

		backendLeveled.SetLevel(level, "")

		logBackends = append(logBackends, backendLeveled)
	}

	logging.SetBackend(logBackends...)

	return nil
}
