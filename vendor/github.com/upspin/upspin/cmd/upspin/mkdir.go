// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "flag"

func (s *State) mkdir(args ...string) {
	const help = `
Mkdir creates Upspin directories.
`
	fs := flag.NewFlagSet("mkdir", flag.ExitOnError)
	s.parseFlags(fs, args, help, "mkdir directory...")
	if fs.NArg() == 0 {
		fs.Usage()
	}
	for _, name := range s.globAllUpspinPath(fs.Args()) {
		_, err := s.client.MakeDirectory(name)
		if err != nil {
			s.exit(err)
		}
	}
}
