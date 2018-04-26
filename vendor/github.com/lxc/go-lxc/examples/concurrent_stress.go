// Copyright © 2013, 2014, The Go-LXC Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"sync"

	"gopkg.in/lxc/go-lxc.v2"
)

var (
	lxcpath       string
	iteration     int
	threads       int
	template      string
	quiet         bool
	startstop     bool
	createdestroy bool
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&template, "template", "busybox", "Template to use")
	flag.IntVar(&threads, "threads", 10, "Number of operations to run concurrently")
	flag.IntVar(&iteration, "iteration", 1, "Number times to run the test")
	flag.BoolVar(&quiet, "quiet", false, "Don't produce any output")
	flag.BoolVar(&startstop, "startstop", false, "Flag to execute Start and Stop")
	flag.BoolVar(&createdestroy, "createdestroy", false, "Flag to execute Create and Destroy")
	flag.Parse()
}

func main() {
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("Using %d GOMAXPROCS\n", runtime.NumCPU())

	var wg sync.WaitGroup

	options := lxc.BusyboxTemplateOptions
	options.Template = template

	for i := 0; i < iteration; i++ {
		log.Printf("-- ITERATION %d --\n", i+1)
		for _, mode := range []string{"CREATE", "START", "STOP", "DESTROY"} {
			log.Printf("\t-- %s --\n", mode)
			for j := 0; j < threads; j++ {
				wg.Add(1)
				go func(i int, mode string) {
					c, err := lxc.NewContainer(strconv.Itoa(i), lxcpath)
					if err != nil {
						log.Fatalf("ERROR: %s\n", err.Error())
					}

					if mode == "CREATE" && startstop == false {
						log.Printf("\t\tCreating the container (%d)...\n", i)
						if err := c.Create(options); err != nil {
							log.Fatalf("\t\t\tERROR: %s\n", err.Error())
						}
					} else if mode == "START" && createdestroy == false {
						log.Printf("\t\tStarting the container (%d)...\n", i)
						if err := c.Start(); err != nil {
							log.Fatalf("\t\t\tERROR: %s\n", err.Error())
						}
					} else if mode == "STOP" && createdestroy == false {
						log.Printf("\t\tStoping the container (%d)...\n", i)
						if err := c.Stop(); err != nil {
							log.Fatalf("\t\t\tERROR: %s\n", err.Error())
						}
					} else if mode == "DESTROY" && startstop == false {
						log.Printf("\t\tDestroying the container (%d)...\n", i)
						if err := c.Destroy(); err != nil {
							log.Fatalf("\t\t\tERROR: %s\n", err.Error())
						}
					}
					wg.Done()
				}(j, mode)
			}
			wg.Wait()
		}
	}
}
