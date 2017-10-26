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
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/server"
	cli "gopkg.in/urfave/cli.v1"

	logging "github.com/op/go-logging"
)

var helpTemplate = `NAME:
{{.Name}} - {{.Usage}}

DESCRIPTION:
{{.Description}}

USAGE:
{{.Name}} {{if .Flags}}[flags] {{end}}command{{if .Flags}}{{end}} [arguments...]

COMMANDS:
	{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
	{{end}}{{if .Flags}}
FLAGS:
	{{range .Flags}}{{.}}
	{{end}}{{end}}
VERSION:
` + cmd.Version +
	`{{ "\n"}}`

var log = logging.MustGetLogger("honeytrap/cmd/honeytrap-serve")

var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "config, c",
		Value: "config.toml",
		Usage: "Load configuration from `FILE`",
	},
	cli.BoolFlag{Name: "cpu-profile", Usage: "Enable cpu profiler"},
	cli.BoolFlag{Name: "mem-profile", Usage: "Enable memory profiler"},
	cli.BoolFlag{Name: "profiler", Usage: "Enable web profiler"},
}

// Cmd defines a struct for defining a command.
type Cmd struct {
	*cli.App
}

func serve(c *cli.Context) {
	options := []server.OptionFn{
		server.WithToken(),
	}

	if v := c.String("config"); v == "" {
	} else if fn, err := server.WithConfig(v); err != nil {
		fmt.Println(color.RedString("Error opening config file: %s", err.Error()))
	} else {
		options = append(options, fn)
	}

	if c.GlobalBool("cpu-profile") {
		options = append(options, server.WithCPUProfiler())
	}

	if c.GlobalBool("mem-profile") {
		options = append(options, server.WithMemoryProfiler())
	}

	var server = server.New(
		options...,
	)

	server.Run(context.Background())

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)

	<-s

	server.Stop()
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	app := cli.NewApp()
	app.Name = "honeytrap"
	app.Author = ""
	app.Usage = "honeytrap"
	app.Flags = globalFlags
	app.Description = `honeytrap: The honeypot server.`
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = serve

	app.RunAndExitOnError()
}
