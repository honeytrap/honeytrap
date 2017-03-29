// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"

	"upspin.io/bind"
	"upspin.io/upspin"
)

func (s *State) getref(args ...string) {
	const help = `
Getref writes to standard output the contents identified by the reference from
the user's default store server. It does not resolve redirections.
`
	fs := flag.NewFlagSet("getref", flag.ExitOnError)
	outFile := fs.String("out", "", "output file (default standard output)")
	s.parseFlags(fs, args, help, "getref [-out=outputfile] ref")

	if fs.NArg() != 1 {
		fs.Usage()
	}
	ref := fs.Arg(0)

	store, err := bind.StoreServer(s.config, s.config.StoreEndpoint())
	if err != nil {
		s.exit(err)
	}
	fmt.Fprintf(os.Stderr, "Using store server at %s\n", s.config.StoreEndpoint())

	data, _, locs, err := store.Get(upspin.Reference(ref))
	if err != nil {
		s.exit(err)
	}
	if len(locs) > 0 {
		fmt.Fprintf(os.Stderr, "Redirection detected:\n")
		for _, loc := range locs {
			fmt.Fprintf(os.Stderr, "%+v\n", loc)
		}
		return
	}

	// Write to outfile or to stdout if none set.
	var output *os.File
	if *outFile == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(*outFile)
		if err != nil {
			s.exit(err)
		}
		defer output.Close()
	}
	_, err = output.Write(data)
	if err != nil {
		s.exitf("Copying to output failed: %v", err)
	}
}
