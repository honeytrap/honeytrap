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

	memLimit, err := c.MemoryLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	memorySwapLimit, err := c.MemorySwapLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	if err := c.SetMemoryLimit(memLimit / 4); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	if err := c.SetMemorySwapLimit(memorySwapLimit / 4); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
}
