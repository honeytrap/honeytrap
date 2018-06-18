/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package config

import (
	"regexp"

	"io"
	"os"

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

// DefaultConfig defines the default Config to be used to set default values.
var Default = Config{}

// Load attempts to load the giving toml configuration file.
func (c *Config) Load(r io.Reader) error {
	md, err := toml.DecodeReader(r, c)
	if err != nil {
		return err
	}
	c.MetaData = md

	logBackends := []logging.Backend{}
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
