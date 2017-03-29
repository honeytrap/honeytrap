// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package serverutil provides helper functions for Upspin servers.
package serverutil // import "upspin.io/serverutil"

import (
	goPath "path"
	"strings"

	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

var (
	errNotExist   = errors.E(errors.NotExist)
	errPrivate    = errors.E(errors.Private)
	errPermission = errors.E(errors.Permission)
)

// ListFunc lists the entries in the directory specified by path.
// It should handle access control internally, returning a Private or
// Permission error if the caller does not have access.
// It should return an ErrFollowLink error iff the given path name is a link.
// In that one case, it should also return only the DirEntry for that path.
type ListFunc func(upspin.PathName) ([]*upspin.DirEntry, error)

// LookupFunc is a DirServer.Lookup implementation.
type LookupFunc func(upspin.PathName) (*upspin.DirEntry, error)

// Glob executes a DirServer.Glob operation for the specified pattern
// using the provided LookupFunc and ListFunc to retrieve data.
func Glob(pattern string, lookup LookupFunc, ls ListFunc) ([]*upspin.DirEntry, error) {
	p, err := path.Parse(upspin.PathName(pattern))
	if err != nil {
		return nil, err
	}

	// If there are no glob meta-characters in the pattern, just do a lookup.
	if !hasMeta(p.FilePath()) {
		de, err := lookup(p.Path())
		if de == nil {
			return nil, err
		}
		return []*upspin.DirEntry{de}, err
	}

	// Look for the longest path prefix that does not contain a
	// metacharacter, so we know which level we need to start listing.
	firstMeta := 0
	i := 0
	for ; i < p.NElem(); i++ {
		firstMeta = i
		if hasMeta(p.Elem(i)) {
			break
		}
	}

	basePath := p.First(firstMeta)                 // Path without the meta component.
	basePattern := p.First(firstMeta + 1).String() // Pattern including first meta component.
	patternTail := strings.TrimPrefix(p.String(), basePattern)

	// The return values of this function.
	var result []*upspin.DirEntry
	var errLink error

	var toGlob []string // Additional patterns to glob.

	entries, err := ls(basePath.Path())
	if err != nil {
		if err == upspin.ErrFollowLink {
			return entries, err
		}
		return nil, errors.E(basePath.Path(), err)
	}
	for _, e := range entries {
		// Match the entire entry name against our base pattern as we
		// are listing the directory before the pattern meta component.
		match, err := goPath.Match(basePattern, string(e.Name))
		if err != nil {
			return nil, errors.E(errors.Invalid, err)
		}
		if !match {
			continue
		}

		if patternTail != "" {
			// If we haven't reached the end of the pattern...
			if e.IsDir() {
				// ...and this is a directory, then append the
				// pattern tail to this name and add it to the
				// list of globs yet to try.
				toGlob = append(toGlob, string(path.Join(e.Name, patternTail)))
				continue
			}
			if !e.IsLink() {
				// ...and this is not a directory or link,
				// then it's only a partial match of the full
				// pattern, so we skip it.
				continue
			}
			// ...and this is a link, we want to emit it as a
			// result but also return a 'must follow link' error.
			errLink = upspin.ErrFollowLink
		}
		result = append(result, e)
	}

	// Perform any additional glob operations recursively.
	for _, pattern := range toGlob {
		entries, err := Glob(pattern, lookup, ls)
		if errors.Match(errPrivate, err) ||
			errors.Match(errPermission, err) ||
			errors.Match(errNotExist, err) {
			// Ignore paths when access is restricted.
			continue
		}
		if err == upspin.ErrFollowLink {
			errLink = err
		} else if err != nil {
			return nil, err
		}
		result = append(result, entries...)
	}

	upspin.SortDirEntries(result, false)
	return result, errLink
}

func hasMeta(elem string) bool {
	return strings.ContainsAny(elem, "*?[^")
}
