package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/config"
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

var (
	log = logging.MustGetLogger("honeytrap/cmd/honeytrap-ls")

	globalFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "c,config",
			Usage: "config file",
			Value: "config.toml",
		},
	}
)

// Cmd defines a struct for defining a command.
type Cmd struct {
	*cli.App
}

// VersionAction defines the action called when seeking the Version detail.
func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("honeytrap-ls: Providing container listing.")))
}

func main() {
	app := cli.NewApp()
	app.Name = "honeytrap-ls"
	app.Author = ""
	app.Usage = "honeytrap-ls"
	app.Flags = globalFlags
	app.Description = `List all current active containers with the running server.`
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
		},
		{
			Name:   "containers",
			Action: serviceContainers,
		},
		{
			Name:   "attackers",
			Action: serviceUsers,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.RunAndExitOnError()
}

// serviceUsers requests all containers users from the running honeytrap instance by
// using the address generated from the config provided.
func serviceUsers(c *cli.Context) {
	conf, err := config.New()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	configFile := c.GlobalString("config")
	if err := conf.Load(configFile); err != nil {
		fmt.Printf("Configuration Error: %q - %q", configFile, err.Error())
		return
	}

	ip, port, _ := net.SplitHostPort(conf.Web.Port)
	if ip == "" {
		ip, _, _ = net.SplitHostPort(getAddr(""))
	}

	webIP := net.JoinHostPort(ip, port)

	var addr string

	if conf.Web.Path != "" {
		addr = fmt.Sprintf("%s/%s", webIP, conf.Web.Path)
	} else {
		addr = webIP
	}

	fmt.Printf("Honeytrap-ls: Attackers\n")
	fmt.Printf("Honeytrap Server: Token: %q\n", conf.Token)
	fmt.Printf("Honeytrap Server: API Addr: %q\n", addr)

	targetAddr := fmt.Sprintf("http://%s/metrics/attackers", addr)

	fmt.Printf("Honeytrap Server: Request Addr: %q\n", targetAddr)

	req, err := http.NewRequest("GET", targetAddr, nil)
	if err != nil {
		fmt.Printf("HTTP Request Error: %q - %q", addr, err.Error())
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP Request Error: %q - %q", addr, err.Error())
		return
	}

	fmt.Printf("Honeytrap Server: API Response Status: %d - %q\n", res.StatusCode, res.Status)

	defer res.Body.Close()

	var body bytes.Buffer
	io.Copy(&body, res.Body)

	fmt.Printf("\n%+s\n", body.String())
}

// serviceContainers requests all containers from the running honeytrap instance by
// using the address generated from the config provided.
func serviceContainers(c *cli.Context) {
	conf, err := config.New()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	configFile := c.GlobalString("config")
	if err := conf.Load(configFile); err != nil {
		fmt.Printf("Configuration Error: %q - %q", configFile, err.Error())
		return
	}

	ip, port, _ := net.SplitHostPort(conf.Web.Port)
	if ip == "" {
		ip, _, _ = net.SplitHostPort(getAddr(""))
	}

	webIP := net.JoinHostPort(ip, port)

	var addr string

	if conf.Web.Path != "" {
		addr = fmt.Sprintf("%s/%s", webIP, conf.Web.Path)
	} else {
		addr = webIP
	}

	fmt.Printf("Honeytrap-ls: Containers\n")
	fmt.Printf("Honeytrap Server: Token: %q\n", conf.Token)
	fmt.Printf("Honeytrap Server: API Addr: %q\n", addr)

	targetAddr := fmt.Sprintf("http://%s/metrics/containers", addr)

	fmt.Printf("Honeytrap Server: Request Addr: %q\n", targetAddr)

	req, err := http.NewRequest("GET", targetAddr, nil)
	if err != nil {
		fmt.Printf("HTTP Request Error: %q - %q", addr, err.Error())
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP Request Error: %q - %q", addr, err.Error())
		return
	}

	fmt.Printf("Honeytrap Server: API Response Status: %d - %q\n", res.StatusCode, res.Status)

	defer res.Body.Close()

	var body bytes.Buffer
	io.Copy(&body, res.Body)

	fmt.Printf("\n%+s\n", body.String())
}

// getAddr takes the giving address string and if it has no ip or use the
// zeroth ip format, then modifies the ip with the current systems ip.
func getAddr(addr string) string {
	if addr == "" {
		if real, err := getMainIP(); err == nil {
			return real + ":0"
		}
	}

	ip, port, err := net.SplitHostPort(addr)
	if err == nil && ip == "" || ip == "0.0.0.0" {
		if realIP, err := getMainIP(); err == nil {
			return net.JoinHostPort(realIP, port)
		}
	}

	return addr
}

// getMainIP returns the giving system IP by attempting to connect to a imaginary
// ip and returns the giving system ip.
func getMainIP() (string, error) {
	udp, err := net.DialTimeout("udp", "8.8.8.8:80", 1*time.Millisecond)
	if udp == nil {
		return "", err
	}

	defer udp.Close()

	localAddr := udp.LocalAddr().String()
	ip, _, _ := net.SplitHostPort(localAddr)

	return ip, nil
}
