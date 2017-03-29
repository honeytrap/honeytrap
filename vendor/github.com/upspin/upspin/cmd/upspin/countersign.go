// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "upspin.io/cmd/upspin"

// This file has the implementation of the countersign command.  Invoke before publishing the new keys.

import (
	"flag"
	"fmt"
	"os"

	"upspin.io/config"
	"upspin.io/pack/ee"
	"upspin.io/upspin"
)

func (s *State) countersign(args ...string) {
	const help = `
Countersign updates the signatures and encrypted data for all items
owned by the user. It is intended to be run after a user has changed
keys.

See the description for rotate for information about updating keys.
`
	fs := flag.NewFlagSet("countersign", flag.ExitOnError)
	s.parseFlags(fs, args, help, "countersign")
	if fs.NArg() != 0 {
		fs.Usage()
	}
	s.countersignCommand(fs)
}

// Countersigner holds the state for the countersign calculation.
type Countersigner struct {
	state  *State
	oldKey upspin.PublicKey
}

// countersignCommand is the main function for the countersign subcommand.
func (s *State) countersignCommand(fs *flag.FlagSet) {
	u, err := s.KeyServer().Lookup(s.config.UserName())
	if err != nil || len(u.PublicKey) == 0 {
		s.exitf("can't find old key for %q: %s\n", s.config.UserName(), err)
	}
	c := &Countersigner{
		state:  s,
		oldKey: u.PublicKey,
	}
	newF := s.config.Factotum()
	if newF == nil {
		s.exitf("no factotum available")
	}

	lastCfg := s.config
	s.config = config.SetFactotum(s.config, s.config.Factotum().Pop())
	defer func() { s.config = lastCfg }()

	root := upspin.PathName(string(s.config.UserName()) + "/")
	entries := c.entriesFromDirectory(root)
	for _, e := range entries {
		c.countersign(e, newF)
	}
}

// countersign adds a second signature using factotum.
func (c *Countersigner) countersign(entry *upspin.DirEntry, newF upspin.Factotum) {
	// TODO(ehg)  Handle PlainPack and EEIntegrityPack as well.
	err := ee.Countersign(c.oldKey, newF, entry)
	if err != nil {
		c.state.exit(err)
	}
	_, err = c.state.DirServer(entry.Name).Put(entry)
	if err != nil {
		// If we get ErrFollowLink, the item changed underfoot, so reporting
		// an error in that case is OK.
		fmt.Fprintf(os.Stderr, "error putting entry back for %q: %s\n", entry.Name, err)
		c.state.exitCode = 1
	}
}

// entriesFromDirectory returns the list of relevant entries in the directory, recursively.
func (c *Countersigner) entriesFromDirectory(dir upspin.PathName) []*upspin.DirEntry {
	// Get list of files for this directory.
	thisDir, err := c.state.DirServer(dir).Glob(upspin.AllFilesGlob(dir)) // Do not want to follow links.
	if err != nil {
		c.state.exitf("globbing %q: %s", dir, err)
	}
	entries := make([]*upspin.DirEntry, 0, len(thisDir))
	// Add plain files that have signatures by self.
	for _, e := range thisDir {
		if !e.IsDir() && !e.IsLink() &&
			e.Packing == upspin.EEPack &&
			string(e.Writer) == string(c.state.config.UserName()) {
			entries = append(entries, e)
		}
	}
	// Recur into subdirectories.
	for _, e := range thisDir {
		if e.IsDir() {
			entries = append(entries, c.entriesFromDirectory(e.Name)...)
		}
	}
	return entries
}
