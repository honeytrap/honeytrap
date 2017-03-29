// Copyright © 2013, 2014, The Go-LXC Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"log"

	"gopkg.in/lxc/go-lxc.v2"
)

var (
	lxcpath string
	name    string
	clear   bool
	x86     bool
	regular bool
)

func init() {
	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&name, "name", "rubik", "Name of the original container")
	flag.BoolVar(&clear, "clear", false, "Attach with clear environment")
	flag.BoolVar(&x86, "x86", false, "Attach using x86 personality")
	flag.BoolVar(&regular, "regular", false, "Attach using a regular user")
	flag.Parse()
}

func main() {
	c, err := lxc.NewContainer(name, lxcpath)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	options := lxc.DefaultAttachOptions
	options.ClearEnv = false
	if clear {
		options.ClearEnv = true
	}
	if x86 {
		options.Arch = lxc.X86
	}
	if regular {
		options.UID = 1000
		options.GID = 1000
	}
	log.Printf("AttachShell\n")
	err = c.AttachShell(options)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("RunCommand\n")
	_, err = c.RunCommand([]string{"id"}, options)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
}
