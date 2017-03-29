// Copyright © 2013, 2014, The Go-LXC Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"log"
	"runtime"
	"strconv"
	"sync"

	"gopkg.in/lxc/go-lxc.v2"
)

var (
	lxcpath string
	count   int
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.IntVar(&count, "count", 10, "Number of containers")
	flag.Parse()
}

func main() {
	var wg sync.WaitGroup

	options := lxc.BusyboxTemplateOptions
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			c, err := lxc.NewContainer(strconv.Itoa(i), lxcpath)
			if err != nil {
				log.Fatalf("ERROR: %s\n", err.Error())
			}

			log.Printf("Creating the container (%d)...\n", i)
			if err := c.Create(options); err != nil {
				log.Fatalf("ERROR: %s\n", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
