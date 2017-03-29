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
)

func init() {
	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.Parse()
}

func main() {
	log.Printf("Defined containers:\n")
	c := lxc.DefinedContainers(lxcpath)
	for i := range c {
		log.Printf("%s (%s)\n", c[i].Name(), c[i].State())
	}

	log.Println()

	log.Printf("Active containers:\n")
	c = lxc.ActiveContainers(lxcpath)
	for i := range c {
		log.Printf("%s (%s)\n", c[i].Name(), c[i].State())
	}

	log.Println()

	log.Printf("Active and Defined containers:\n")
	c = lxc.ActiveContainers(lxcpath)
	for i := range c {
		log.Printf("%s (%s)\n", c[i].Name(), c[i].State())
	}
}
