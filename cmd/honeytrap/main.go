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
package honeytrap

import (
	"context"
	"fmt"

	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/server"
	"github.com/honeytrap/honeytrap/services"
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
	cli.StringFlag{
		Name:  "data, d",
		Value: "~/.honeytrap",
		Usage: "Store data in `DIR`",
	},
	cli.BoolFlag{Name: "cpu-profile", Usage: "Enable cpu profiler"},
	cli.BoolFlag{Name: "mem-profile", Usage: "Enable memory profiler"},

	cli.BoolFlag{Name: "list-services", Usage: "List the available services"},
	cli.BoolFlag{Name: "list-channels", Usage: "List the available channels"},
	cli.BoolFlag{Name: "list-listeners", Usage: "List the available listeners"},
}

// Cmd defines a struct for defining a command.
type Cmd struct {
	*cli.App
}

func tryConfig(path string) (server.OptionFn, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "":
		fallthrough
	case "file":
		fn, err := server.WithConfig(u.Path)
		if err != nil {
			return nil, err
		}
		return fn, nil
	case "http":
		fallthrough
	case "https":
		fn, err := server.WithRemoteConfig(path)
		if err != nil {
			return nil, err
		}
		return fn, nil
	default:
		return nil, fmt.Errorf("Unknown path scheme %s", u.Scheme)
	}
}

func serve(c *cli.Context) error {
	var options []server.OptionFn

	// Honeytrap will search for a config file in these files, in descending priority
	configCandidates := []string{
		c.String("config"),
		"/etc/honeytrap/config.toml",
		"/etc/honeytrap.toml",
	}

	successful := false
	for _, candidate := range configCandidates {
		fn, err := tryConfig(candidate)
		if err != nil {
			log.Error("Failed to read config file %s: %s", candidate, err.Error())
			continue
		}
		log.Debug("Using config file %s\n", candidate)
		options = append(options, fn)
		successful = true
		break
	}
	if !successful {
		return cli.NewExitError("No configuration file found! Check your config (-c).", 1)
	}

	if d := c.String("data"); d == "" {
	} else if fn, err := server.WithDataDir(d); err != nil {
		ec := cli.NewExitError(err.Error(), 1)
		return ec
	} else {
		options = append(options, fn)
	}

	options = append(options, server.WithToken())

	if c.GlobalBool("cpu-profile") {
		options = append(options, server.WithCPUProfiler())
	}

	if c.GlobalBool("mem-profile") {
		options = append(options, server.WithMemoryProfiler())
	}

	srvr, err := server.New(
		options...,
	)

	if err != nil {
		ec := cli.NewExitError(err.Error(), 1)
		return ec
	}

	// enumerate the available services
	if c.GlobalBool("list-services") {
		fmt.Println("services")
		fmt.Println("=======")
		services.Range(func(name string) {
			fmt.Printf("* %s\n", name)
		})
		return nil
	}

	// enumerate the available channels
	if c.GlobalBool("list-channels") {
		fmt.Println("channels")
		fmt.Println("=======")
		pushers.Range(func(name string) {
			fmt.Printf("* %s\n", name)
		})
		return nil
	}

	// enumerate the available listeners
	if c.GlobalBool("list-listeners") {
		fmt.Println("listeners")
		fmt.Println("=======")
		listener.Range(func(name string) {
			fmt.Printf("* %s\n", name)
		})
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt)
		signal.Notify(s, syscall.SIGTERM)

		select {
		case <-s:
			cancel()
		}
	}()

	srvr.Run(ctx)
	srvr.Stop()
	return nil
}

func New() *cli.App {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer,
			`Version: %s
Release-Tag: %s
Commit-ID: %s
`, color.YellowString(cmd.Version), color.YellowString(cmd.ReleaseTag), color.YellowString(cmd.CommitID))
	}

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

	return app
}
