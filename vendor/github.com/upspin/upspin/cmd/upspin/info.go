// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"
	"time"

	"upspin.io/access"
	"upspin.io/pack"
	"upspin.io/path"
	"upspin.io/upspin"
)

func (s *State) info(args ...string) {
	const help = `
Info prints to standard output a thorough description of all the
information about named paths, including information provided by
ls but also storage references, sizes, and other metadata.

If the path names an Access or Group file, it is also checked for
validity. If it is a link, the command attempts to access the target
of the link.
`
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	s.parseFlags(fs, args, help, "info path...")

	if fs.NArg() == 0 {
		fs.Usage()
	}
	for _, name := range fs.Args() {
		entries, err := s.DirServer(upspin.PathName(name)).Glob(name)
		// ErrFollowLink is OK; we still get the relevant entry.
		if err != nil && err != upspin.ErrFollowLink {
			s.exit(err)
		}
		for _, entry := range entries {
			s.printInfo(entry)
			switch {
			case access.IsAccessFile(entry.Name):
				s.checkAccessFile(entry.Name)
			case access.IsGroupFile(entry.Name):
				s.checkGroupFile(entry.Name)
			}
		}
	}
}

// infoDirEntry wraps a DirEntry to allow new methods for easy formatting.
// It also has fields that hold relevant information as we acquire it.
type infoDirEntry struct {
	*upspin.DirEntry
	state *State
	// The following fields are computed as we run.
	access    *access.Access
	lastUsers string
}

func (d *infoDirEntry) TimeString() string {
	return d.Time.Go().In(time.Local).Format("Mon Jan 2 15:04:05 MST 2006")
}

func (d *infoDirEntry) AttrString() string {
	return attrFormat(d.Attr)
}

func (d *infoDirEntry) Rights() []access.Right {
	return []access.Right{access.Read, access.Write, access.List, access.Create, access.Delete}
}

func (d *infoDirEntry) Readers() string {
	d.state.sharer.addAccess(d.DirEntry)
	d.lastUsers = "<nobody>"
	if d.IsDir() {
		return "is a directory"
	}
	_, users, _, err := d.state.sharer.readers(d.DirEntry)
	if err != nil {
		return err.Error()
	}
	d.lastUsers = users
	return users
}

func (d *infoDirEntry) Sequence() int64 {
	return upspin.SeqVersion(d.DirEntry.Sequence)
}

func (d *infoDirEntry) Hashes() string {
	h := ""
	if d.IsDir() || d.Packing != upspin.EEPack {
		return h
	}
	packer := pack.Lookup(d.Packing)
	hashes, err := packer.ReaderHashes(d.Packdata)
	if err != nil {
		return h
	}
	for _, r := range hashes {
		if h == "" {
			h += " "
		}
		h += fmt.Sprintf("%x...", r[:4])
	}
	return h
}

func (d *infoDirEntry) Users(right access.Right) string {
	users := userListToString(d.state.usersWithAccess(d.state.client, d.access, right))
	if users == d.lastUsers {
		return "(same)"
	}
	d.lastUsers = users
	return users
}

func (d *infoDirEntry) WhichAccess() string {
	var acc *access.Access
	accEntry, err := d.state.whichAccessFollowLinks(d.Name)
	if err != nil {
		return err.Error()
	}
	accFile := "owner only"
	if accEntry == nil {
		// No access file applies.
		acc, err = access.New(d.Name)
		if err != nil {
			// Can't happen, since the name must be valid.
			d.state.exitf("%q: %s", d.Name, err)
		}
	} else {
		accFile = string(accEntry.Name)
		data, err := read(d.state.client, accEntry.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open access file %q: %s\n", accFile, err)
		}
		acc, err = access.Parse(accEntry.Name, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse access file %q: %s\n", accFile, err)
		}
	}
	d.access = acc
	return accFile
}

// printInfo prints, in human-readable form, most of the information about
// the entry, including the users that have permission to access it.
// TODO: Present this more neatly.
// TODO: Present group information.
func (s *State) printInfo(entry *upspin.DirEntry) {
	infoDir := &infoDirEntry{
		state:    s,
		DirEntry: entry,
	}
	writer := tabwriter.NewWriter(os.Stdout, 4, 4, 1, ' ', 0)
	err := infoTmpl.Execute(writer, infoDir)
	if err != nil {
		s.exitf("executing info template: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		s.exitf("flushing template output: %v", err)
	}
	if !entry.IsLink() {
		return
	}
	// Check and print information about the link target.
	target, err := s.client.Lookup(entry.Link, true)
	if err != nil {
		// Print the whole error indented, starting on the next line. This helps it stand out.
		s.exitf("Error: link %s has invalid target %s:\n\t%v", entry.Name, entry.Link, err)
	}
	fmt.Printf("Target of link %s:\n", entry.Name)
	s.printInfo(target)
}

func attrFormat(attr upspin.Attribute) string {
	a := attr
	tail := ""
	if a&upspin.AttrIncomplete > 0 {
		tail = " (incomplete)"
		a ^= upspin.AttrIncomplete
	}
	switch a {
	case upspin.AttrNone:
		return "none (plain file)" + tail
	case upspin.AttrDirectory:
		return "directory" + tail
	case upspin.AttrLink:
		return "link" + tail
	}
	return fmt.Sprintf("attribute(%#x)", attr)
}

var infoTmpl = template.Must(template.New("info").Parse(infoText))

const infoText = `{{.Name}}
	packing:	{{.Packing}}
	size:	{{.Size}}
	time:	{{.TimeString}}
	writer:	{{.Writer}}
	attributes:	{{.AttrString}}
	sequence:	{{.Sequence}}
	access file:	{{.WhichAccess}}
	key holders: 	{{.Readers}}
	key hashes:     {{.Hashes}}
	{{range $right := .Rights -}}
	can {{$right}}:	{{$.Users $right}}
	{{end -}}
	Block#	Offset	Size	Location
	{{range $index, $block := .Blocks -}}
	{{$index}}	{{.Offset}}	{{.Size}}	{{.Location}}
	{{end}}`

// checkGroupFile diagnoses likely problems with the contents and rights
// of the Group file.
// TODO: We could check that packing is Plain but that should never be a problem.
func (s *State) checkGroupFile(name upspin.PathName) {
	parsed, err := path.Parse(name)
	if err != nil {
		s.exit(err) // Should never happen.
	}
	groupSeen := make(map[upspin.PathName]bool)
	userSeen := make(map[upspin.UserName]bool)
	s.doCheckGroupFile(parsed, groupSeen, userSeen)
}

// doCheckGroupFile is the inner, recursive implementation of checkGroupFile.
func (s *State) doCheckGroupFile(parsed path.Parsed, groupSeen map[upspin.PathName]bool, userSeen map[upspin.UserName]bool) {
	group := parsed.Path()
	if groupSeen[group] {
		return
	}
	groupSeen[group] = true
	data, err := s.client.Get(group)
	if err != nil {
		s.exitf("cannot read Group file: %v", err)
	}

	// Get the Access file, if any, that applies.
	// TODO: We've already got it in earlier code, so could save it.
	whichAccess, err := s.DirServer(group).WhichAccess(group)
	if err != nil {
		s.exitf("unexpected error finding Access file for Group file %s: %v", group, err)
	}
	var accessFile *access.Access
	if whichAccess == nil {
		accessFile, err = access.New(group)
		if err != nil {
			s.exitf("cannot create default Access file: %v", err)
		}
	} else {
		data, err := s.client.Get(whichAccess.Name)
		if err != nil {
			s.exitf("cannot get Access file: %v", err)
		}
		accessFile, err = access.Parse(whichAccess.Name, data)
		if err != nil {
			s.exitf("cannot parse Access file: %v", err)
		}
	}

	// Each member should be either a plain user or a group and be able to access the Group file.
	members, err := access.ParseGroup(parsed, data)
	if err != nil {
		s.exitf("error parsing Group file %s: %v", group, err)
	}
	for _, member := range members {
		if member.IsRoot() {
			// Normal user.
			user := member.User()
			if !s.userExists(user, userSeen) {
				s.failf("user %s in Group file %s not found in key server", user, group)
				continue
			}
			// Member must be able to read the Group file.
			canRead, err := accessFile.Can(user, access.Read, group, s.client.Get)
			if err != nil {
				s.exitf("error checking permissions in Group file %s for user %s: %v", group, user, err)
				continue
			}
			if !canRead {
				s.failf("user %s is missing read access for group %s", user, group)
			}
			continue
		}
		if !access.IsGroupFile(member.Path()) {
			s.failf("do not understand member %s of Group file %s", member, parsed) // Should never happen.
			continue
		}
		// Member is a group. Recur using Group file.
		s.doCheckGroupFile(member, groupSeen, userSeen)
	}
}

func (s *State) checkAccessFile(name upspin.PathName) {
	data, err := s.client.Get(name)
	if err != nil {
		s.exitf("cannot get Access file: %v", err)
	}
	accessFile, err := access.Parse(name, data)
	if err != nil {
		s.exitf("cannot parse Access file: %v", err)
	}
	users := accessFile.List(access.AnyRight)

	groupSeen := make(map[upspin.PathName]bool)
	userSeen := make(map[upspin.UserName]bool)
	for _, user := range users {
		if user.IsRoot() {
			// Normal user.
			if !s.userExists(user.User(), userSeen) {
				s.failf("user %s in Access file %s not found in key server", user.User(), name)
			}
			continue
		}
		// Member is a group.
		s.doCheckGroupFile(user, groupSeen, userSeen)
	}
}

func (s *State) userExists(user upspin.UserName, userSeen map[upspin.UserName]bool) bool {
	if userSeen[user] || user == access.AllUsers { // all@upspin.io is baked in.
		return true // Previous answer will do.
	}
	// Ignore wildcards.
	if isWildcardUser(user) {
		return true
	}
	userSeen[user] = true
	_, err := s.KeyServer().Lookup(user)
	return err == nil
}
