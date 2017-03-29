// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"

	"upspin.io/config"
)

func (s *State) rotate(args ...string) {
	const help = `
Rotate pushes an updated key to the key server.

To update an Upspin key, the sequence is:

  upspin keygen            # Create new key.
  upspin countersign       # Update file signatures to use new key.
  upspin rotate            # Save new key to key server.
  upspin share -r -fix me@example.com/  # Update keys in file metadata. 

Keygen creates a new key and saves the old one. Countersign walks
the file tree and adds signatures with the new key alongside those
for the old. Rotate pushes the new key to the KeyServer. Share walks
the file tree, re-wrapping the encryption keys that were encrypted
with the old key to use the new key.

Some of these steps could be folded together but the full sequence
makes it easier to recover if a step fails.

TODO: Rotate and countersign are terms of art, not clear to users.
`
	fs := flag.NewFlagSet("rotate", flag.ExitOnError)
	s.parseFlags(fs, args, help, "rotate")
	if fs.NArg() != 0 {
		fs.Usage()
	}

	f := s.config.Factotum() // save latest factotum
	if f == nil {
		s.exitf("no factotum available")
	}
	lastCfg := s.config
	s.config = config.SetFactotum(s.config, f.Pop()) // config now defaults to old key
	defer func() { s.config = lastCfg }()

	keyServer := s.KeyServer()
	u, err := keyServer.Lookup(s.config.UserName())
	if err != nil {
		s.exit(err)
	}
	u.PublicKey = f.PublicKey()
	err = keyServer.Put(u)
	if err != nil {
		s.exit(err)
	}
}
