package main

import "github.com/honeytrap/honeytrap/cmd"

func main() {
	app := cmd.New()
	app.RunAndExitOnError()
}
