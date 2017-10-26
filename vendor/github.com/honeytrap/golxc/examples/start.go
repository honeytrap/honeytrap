// Copyright © 2013, 2014, The Go-LXC Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"log"
	"time"

	"gopkg.in/lxc/go-lxc.v2"
)

var (
	lxcpath string
	name    string
)

func init() {
	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&name, "name", "rubik", "Name of the container")
	flag.Parse()
}

func main() {
	c, err := lxc.NewContainer(name, lxcpath)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	c.SetLogFile("/tmp/" + name + ".log")
	c.SetLogLevel(lxc.TRACE)

	log.Printf("Starting the container...\n")
	if err := c.Start(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("Waiting container to startup networking...\n")
	if _, err := c.WaitIPAddresses(5 * time.Second); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
}
