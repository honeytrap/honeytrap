package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/process"
	"github.com/minio/cli"
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
		cmd := process.Command{
			Name:  name,
			Level: process.RedAlert,
			Args:  []string(c.Args()),
		}

		if err := cmd.Run(context.Background(), os.Stdout, os.Stderr); err != nil {
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
			Name:            "containers",
			Action:          run("honeytrap-containers"),
			SkipFlagParsing: true,
		},
		{
			Name:            "users",
			Action:          run("honeytrap-users"),
			SkipFlagParsing: true,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	return &Cmd{
		App: app,
	}
}
