package main

import (
	"fmt"

	"github.com/fatih/color"
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

var log = logging.MustGetLogger("honeytrap/cmd/lxc-cli")

var globalFlags = []cli.Flag{
/*
	cli.StringFlag{
		Name:  "c,config",
		Usage: "config file",
		Value: "config.toml",
	},
	cli.BoolFlag{
		Name:  "help, h",
		Usage: "Show help.",
	},
*/
}

// Cmd defines a struct for defining a command.
type Cmd struct {
	*cli.App
}

// VersionAction defines the action called when seeking the Version detail.
func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("LxcCLI: Providing lxc interaction commands.")))
}

func service(c *cli.Context) {

}

func main() {
	app := cli.NewApp()
	app.Name = "honeytrap-ls"
	app.Author = ""
	app.Usage = "honeytrap-ls"
	app.Flags = globalFlags
	app.Description = `The honeypot CLI tool to see containers in use.`
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = service

	app.RunAndExitOnError()
}
