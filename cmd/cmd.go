package cmd

import (
	"bytes"
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/process"
	"github.com/minio/cli"
	"github.com/op/go-logging"
)

// Version defines the version number for the cli.
var Version = "0.1"

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
` + Version +
	`{{ "\n"}}`

var log = logging.MustGetLogger("honeytrap/cmd")

var serveFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "c,config",
		Usage: "config file",
		Value: "config.toml",
	},
	/*
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Show help.",
		},
	*/
	cli.BoolFlag{Name: "cpu-profile", Usage: "Enable cpu profiler"},
	cli.BoolFlag{Name: "mem-profile", Usage: "Enable memory profiler"},
	cli.BoolFlag{Name: "profiler", Usage: "Enable web profiler"},
}

// Cmd defines a struct for defining a command.
type Cmd struct {
	*cli.App
}

// VersionAction defines the action called when seeking the Version detail.
func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap: The ultimate honeypot framework.")))
}

func runServer(c *cli.Context) {
	configFile := c.String("config")
	profilerEnabled := c.GlobalBool("profiler")
	cpuProfileFile := c.GlobalBool("cpu-profile")
	memProfileFile := c.GlobalBool("mem-profile")

	configArg := fmt.Sprintf("--config %s", configFile)
	profilerArg := fmt.Sprintf("--profiler %t", profilerEnabled)
	cpuProfileArg := fmt.Sprintf("--cpu-profile %t", cpuProfileFile)
	memProfileArg := fmt.Sprintf("--mem-profile %t", memProfileFile)

	serverCmd := process.AsyncProcess{
		Commands: []process.Command{
			{
				Name:  "honeytrap-serve",
				Level: process.RedAlert,
				Args:  []string{configArg, profilerArg, cpuProfileArg, memProfileArg},
			},
		},
	}

	var pout, perr bytes.Buffer

	if err := serverCmd.AsyncExec(context.Background(), &pout, &perr); err != nil {
		fmt.Println(perr.String())
		return
	}

	fmt.Println(pout.String())
}

// New returns a new instance of the Cmd struct.
func New() *Cmd {
	app := cli.NewApp()
	app.Name = "honeytrap"
	app.Author = ""
	app.Usage = "honeytrap"
	app.Description = `The ultimate honeypot framework.`
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
		},
		{
			Name:   "serve",
			Action: runServer,
			Flags:  serveFlags,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	return &Cmd{
		App: app,
	}
}
