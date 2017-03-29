// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"log"

	"upspin.io/client"
	"upspin.io/config"
	"upspin.io/pack"
	"upspin.io/upspin"
)

func (s *State) repack(args ...string) {
	const help = `
Repack rewrites the data referred to by each path , storing it again using the
packing specificied by its -pack option, ee by default. If the data is already
packed with the specified packing, the data is untouched unless the -f (force)
flag is specified, which can be helpful if the data is to be repacked using a
fresh key.

Repack does not delete the old storage. See the deletestorage command
for more information.
`
	fs := flag.NewFlagSet("repack", flag.ExitOnError)
	fs.Bool("f", false, "force repack even if the file is already packed as requested")
	fs.String("pack", "ee", "packing to use when rewriting")
	fs.Bool("r", false, "recur into subdirectories")
	fs.Bool("v", false, "verbose: log progress")
	s.parseFlags(fs, args, help, "repack [-pack ee] [flags] path...")
	if fs.NArg() == 0 {
		fs.Usage()
	}

	s.repackCommand(fs)
}

// repackCommand implements the repack command. It builds a temporary client
// with the new packing and iterates over the files.
func (s *State) repackCommand(fs *flag.FlagSet) {
	packer := pack.LookupByName(stringFlag(fs, "pack"))
	if packer == nil {
		s.exitf("no such packing %q", stringFlag(fs, "pack"))
	}

	prevClient := s.client
	s.client = client.New(config.SetPacking(s.config, packer.Packing()))
	defer func() { s.client = prevClient }()

	for _, entry := range s.globAllUpspin(fs.Args()) {
		s.repackFileOrDir(entry, packer, boolFlag(fs, "f"), boolFlag(fs, "r"), boolFlag(fs, "v"))
	}
}

// repackFileOrDir repacks its argument. If it is a directory and the -r flag is set, it descends.
// The implementation makes a copy and then does some renaming to avoid wiping the
// original if something goes wrong, but it is not foolproof.
func (s *State) repackFileOrDir(entry *upspin.DirEntry, packer upspin.Packer, force, recur, verbose bool) {
	name := entry.Name
	if verbose {
		log.Printf("repack %s", name)
	}
	if entry.IsDir() {
		if !recur {
			s.exitf("%q is a directory", name)
		}
		entries, err := s.client.Glob(upspin.AllFilesGlob(name))
		if err != nil {
			s.exit(err)
		}
		for _, entry := range entries {
			s.repackFileOrDir(entry, packer, force, true, verbose)
		}
		return
	}
	if entry.Packing == packer.Packing() && !force {
		if verbose {
			log.Printf("%s already packed with %s", name, packer)
		}
		return
	}
	// The implementation copies the old to the new and then
	// renames, so if there is an error we don't lose the original.
	// This requires create permission but does not require the
	// whole file be in memory. TODO rewrite in place?
	old, err := s.client.Open(entry.Name)
	if err != nil {
		s.exit(err)
	}
	new, err := s.client.Create(entry.Name + "._rename")
	if err != nil {
		old.Close()
		s.exit(err)
	}
	// Will close by hand - no defer - so renames happens with no I/O open.
	_, err = io.Copy(new, old)
	old.Close()
	if err != nil {
		new.Close()
		s.exit(err)
	}
	err = new.Close()
	if err != nil {
		s.exit(err)
	}
	// New file exists. Delete the old one.
	err = s.client.Delete(old.Name())
	if err != nil {
		// Failure. The old file exists, so delete the new one if we can.
		s.client.Delete(new.Name())
		s.exit(err)
	}
	// Scary moment!
	err = s.client.Rename(new.Name(), old.Name())
	if err != nil {
		log.Printf("rename failed, but repacked contents are now in %q", new.Name())
		s.exit(err)
	}
}
