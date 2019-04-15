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
	flag.StringVar(&name, "name", "rubik", "Name of the original container")
	flag.Parse()
}

func main() {
	c, err := lxc.NewContainer(name, lxcpath)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("IPAddress(\"lo\")\n")
	if addresses, err := c.IPAddress("lo"); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		for i, v := range addresses {
			log.Printf("%d) %s\n", i, v)
		}
	}

	log.Printf("IPAddresses()\n")
	if addresses, err := c.IPAddresses(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		for i, v := range addresses {
			log.Printf("%d) %s\n", i, v)
		}
	}

	log.Printf("IPv4Addresses()\n")
	if addresses, err := c.IPv4Addresses(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		for i, v := range addresses {
			log.Printf("%d) %s\n", i, v)
		}
	}

	log.Printf("IPv6Addresses()\n")
	if addresses, err := c.IPv6Addresses(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		for i, v := range addresses {
			log.Printf("%d) %s\n", i, v)
		}
	}
}
