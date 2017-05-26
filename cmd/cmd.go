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

func run(name string) func(c *cli.Context) {
	return func(c *cli.Context) {
		cmd := process.SyncProcess{
			Commands: []process.Command{
				{
					Name:  name,
					Level: process.SilentKill,
					Args:  []string(c.Args()),
				},
			},
		}

		if err := cmd.Exec(context.Background(), os.Stdout, os.Stderr); err != nil {
			fmt.Println(color.RedString(err.Error()))
			return
		}
	}
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
			Name:            "serve",
			Action:          run("honeytrap-serve"),
			SkipFlagParsing: true,
		},
		{
			Name:   "rm",
			Action: runRM,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "c,config",
					Usage: "config file",
					Value: "config.toml",
				},
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

func runRM(c *cli.Context) {
	configFile := c.String("config")

	var extras []string

	for _, item := range c.Args() {
		extras = append(extras, string(item))
	}

	serverCmd := process.SyncProcess{
		Commands: []process.Command{
			{
				Name:  "honeytrap-rm",
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
