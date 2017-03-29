// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
)

func (s *State) get(args ...string) {
	const help = `
Get writes to standard output the contents identified by the Upspin path.
`
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	outFile := fs.String("out", "", "output file (default standard output)")
	s.parseFlags(fs, args, help, "get [-out=outputfile] path")

	names := s.globAllUpspinPath(fs.Args())
	if len(names) != 1 {
		fs.Usage()
	}

	data, err := s.client.Get(names[0])
	if err != nil {
		s.exit(err)
	}
	// Write to outfile or to stdout if none set
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
