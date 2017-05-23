package cmd

import (
	"context"
	"fmt"
	"os"

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

// Cmd defines a struct for defining a command.
type Cmd struct {
	*cli.App
}

// VersionAction defines the action called when seeking the Version detail.
func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap: The ultimate honeypot framework.")))
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "c,config",
					Usage: "config file",
					Value: "config.toml",
				},
				cli.BoolFlag{Name: "cpu-profile", Usage: "Enable cpu profiler"},
				cli.BoolFlag{Name: "mem-profile", Usage: "Enable memory profiler"},
				cli.BoolFlag{Name: "profiler", Usage: "Enable web profiler"},
			},
		},
		{
			Name:   "ls",
			Action: runLS,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "c,config",
					Usage: "config file",
					Value: "config.toml",
				},
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	return &Cmd{
		App: app,
	}
}

func runLS(c *cli.Context) {
	configFile := c.String("config")

	var extras []string

	for _, item := range c.Args() {
		extras = append(extras, string(item))
	}

	serverCmd := process.SyncProcess{
		Commands: []process.Command{
			{
				Name:  "honeytrap-ls",
				Level: process.SilentKill,
				Args: append([]string{
					"--config", configFile,
				}, extras...),
			},
		},
	}

	if err := serverCmd.Exec(context.Background(), os.Stdout, os.Stderr); err != nil {
		fmt.Printf("Error occured: %+q", err)
		return
	}

}

func runServer(c *cli.Context) {
	configFile := c.String("config")
	profilerEnabled := c.GlobalBool("profiler")
	cpuProfileFile := c.GlobalBool("cpu-profile")
	memProfileFile := c.GlobalBool("mem-profile")

	serverCmd := process.SyncProcess{
		Commands: []process.Command{
			{
				Name:  "honeytrap-serve",
				Level: process.SilentKill,
				Args: []string{
					"--config", configFile,
					"--profiler", fmt.Sprintf("%t", profilerEnabled),
					"--cpu-profile", fmt.Sprintf("%t", cpuProfileFile),
					"--mem-profile", fmt.Sprintf("%t", memProfileFile),
				},
			},
		},
	}

	if err := serverCmd.Exec(context.Background(), os.Stdout, os.Stderr); err != nil {
		fmt.Printf("Error occured: %+q", err)
		return
	}

}
