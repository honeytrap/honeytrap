// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package server implements DirServer using a Tree as backing.
package server

import (
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"upspin.io/access"
	"upspin.io/cache"
	"upspin.io/dir/server/tree"
	"upspin.io/errors"
	"upspin.io/log"
	"upspin.io/metric"
	"upspin.io/path"
	"upspin.io/serverutil"
	"upspin.io/upspin"
	"upspin.io/valid"
)

// TODO(edpin): move the the special-casing about snapshots into hasRight.

// common error values.
var (
	errNotExist = errors.E(errors.NotExist)
	errPrivate  = errors.E(errors.Private)
	errReadOnly = errors.E(errors.Permission, errors.Str("tree is read only"))
)

const (
	// entryMustBeClean is used with lookup to specify whether the caller
	// needs to look at the dir entry's references and therefore whether the
	// tree must be flushed if a dirty dir entry is found.
	entryMustBeClean = true
)

// server implements upspin.DirServer.
type server struct {
	// serverConfig holds this server's Factotum, server name and store
	// endpoint where to store dir entries. It is set when the server is
	// first registered and never reset again.
	serverConfig upspin.Config

	// userName is the name of the user on behalf of whom this
	// server is serving.
	userName upspin.UserName

	// logDir is the directory path accessible through the local file system
	// where user logs are stored.
	logDir string

	// userTrees keeps track of user trees in LRU fashion, where key
	// is an upspin.UserName and value is the tree.Tree for that user name.
	// Access to userTrees must be protected by the user lock. Get the
	// user lock by calling userLock(userName) and take it prior to getting
	// a Tree from the userTree and while using the Tree. Concurrent access
	// for different users is okay as the LRU is thread-safe.
	userTrees *cache.LRU

	// access caches the parsed contents of Access files as struct
	// accessEntry, indexed by their path names.
	access *cache.LRU

	// defaultAccess caches parsed empty Access files that implicitly exists
	// at the root of every user's tree, if an explicit one is not found.
	// It's indexed by the username.
	defaultAccess *cache.LRU

	// remoteGroups caches groupEntry objects that store remote Group files
	// that must be periodically forgotten so they're reloaded fresh again
	// when needed.
	remoteGroups *cache.LRU

	// userLocks is a pool of user locks. To find the correct lock for a
	// user, a string hash of a username selects the index into the slice to
	// use. This fixed pool ensures we don't have a growing number of locks
	// and that we also don't have a race creating new locks when we first
	// touch a user.
	userLocks []sync.Mutex

	// snapshotControl is a channel for passing control messages to the
	// snapshot loop. Possible control messages are: the username to
	// snapshot or close the channel to stop the snapshot loop.
	snapshotControl chan upspin.UserName

	// now returns the time now. It's usually just upspin.Now but is
	// overridden for tests.
	now func() upspin.Time
}

var _ upspin.DirServer = (*server)(nil)

// options are optional parameters to almost every inner method of directory
// for doing optional, non-correctness-related work.
type options struct {
	span *metric.Span
	// Add other things below (for example, some health monitoring stats).
}

// New creates a new instance of DirServer with the given options
func New(cfg upspin.Config, options ...string) (upspin.DirServer, error) {
	const op = "dir/server.New"
	if cfg == nil {
		return nil, errors.E(op, errors.Invalid, errors.Str("nil config"))
	}
	if cfg.DirEndpoint().Transport == upspin.Unassigned {
		return nil, errors.E(op, errors.Invalid, errors.Str("directory endpoint cannot be unassigned"))
	}
	if cfg.KeyEndpoint().Transport == upspin.Unassigned {
		return nil, errors.E(op, errors.Invalid, errors.Str("key endpoint cannot be unassigned"))
	}
	if cfg.StoreEndpoint().Transport == upspin.Unassigned {
		return nil, errors.E(op, errors.Invalid, errors.Str("store endpoint cannot be unassigned"))
	}
	if cfg.UserName() == "" {
		return nil, errors.E(op, errors.Invalid, errors.Str("empty user name"))
	}
	if cfg.Factotum() == nil {
		return nil, errors.E(op, errors.Invalid, errors.Str("nil factotum"))
	}
	// Check which options are present and pick suitable defaults.
	userCacheSize := 1000
	accessCacheSize := 1000
	groupCacheSize := 100
	logDir := ""
	for _, opt := range options {
		o := strings.Split(opt, "=")
		if len(o) != 2 {
			return nil, errors.E(op, errors.Invalid, errors.Errorf("invalid option format: %q", opt))
		}
		k, v := o[0], o[1]
		switch k {
		case "userCacheSize", "accessCacheSize", "groupCacheSize":
			cacheSize, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return nil, errors.E(op, errors.Invalid, errors.Errorf("invalid cache size %q: %s", v, err))
			}
			if cacheSize < 1 {
				return nil, errors.E(op, errors.Invalid, errors.Errorf("%s: cache size too small: %d", k, cacheSize))
			}
			switch opt {
			case "userCacheSize":
				userCacheSize = int(cacheSize)
			case "accessCacheSize":
				accessCacheSize = int(cacheSize)
			case "groupCacheSize":
				groupCacheSize = int(cacheSize)
			}
		case "logDir":
			logDir = v
		default:
			return nil, errors.E(op, errors.Invalid, errors.Errorf("unknown option %q", k))
		}
	}
	if logDir == "" {
		dir, err := ioutil.TempDir("", "DirServer")
		if err != nil {
			return nil, errors.E(op, errors.IO, err)
		}
		log.Error.Printf("%s: warning: writing important logs to a temporary directory (%q). A server restart will lose data.", op, dir)
		logDir = dir
	}

	s := &server{
		serverConfig:  cfg,
		userName:      cfg.UserName(),
		logDir:        logDir,
		userTrees:     cache.NewLRU(userCacheSize),
		access:        cache.NewLRU(accessCacheSize),
		defaultAccess: cache.NewLRU(accessCacheSize),
		remoteGroups:  cache.NewLRU(groupCacheSize),
		userLocks:     make([]sync.Mutex, numUserLocks),
		now:           upspin.Now,
	}
	serverutil.RegisterShutdown(s.shutdown)
	// Start background services.
	s.startSnapshotLoop()
	go s.groupRefreshLoop()
	return s, nil
}

// Lookup implements upspin.DirServer.
func (s *server) Lookup(name upspin.PathName) (*upspin.DirEntry, error) {
	const op = "dir/server.Lookup"
	o, m := newOptMetric(op)
	defer m.Done()
	return s.lookupWithPermissions(op, name, o)
}

func (s *server) lookupWithPermissions(op string, name upspin.PathName, opts ...options) (*upspin.DirEntry, error) {
	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}

	if isSnapshotUser(p.User()) {
		if isSnapshotOwner(s.userName, p.User()) {
			return s.lookup(op, p, entryMustBeClean, opts...)
		}
		// Non-owners cannot see other people's snapshots.
		return nil, errors.E(op, name, errPrivate)
	}

	entry, err := s.lookup(op, p, entryMustBeClean, opts...)

	// Check if the user can know about the file at all. If not, to prevent
	// leaking its existence, return NotExist.
	if err == upspin.ErrFollowLink {
		return s.errLink(op, entry, opts...)
	}
	if err != nil {
		if errors.Match(errNotExist, err) {
			if canAny, _, err := s.hasRight(access.AnyRight, p, opts...); err != nil {
				return nil, err
			} else if !canAny {
				return nil, errors.E(op, name, errors.Private)
			}
		}
		return nil, err // s.lookup wraps err already.
	}

	// Check for Read access permission.
	canRead, _, err := s.hasRight(access.Read, p, opts...)
	if err == upspin.ErrFollowLink {
		return nil, errors.E(op, errors.Internal, p.Path(), errors.Str("can't be link at this point"))
	}
	if err != nil {
		return nil, errors.E(op, err)
	}
	if !canRead {
		canAny, _, err := s.hasRight(access.AnyRight, p, opts...)
		if err != nil {
			return nil, errors.E(op, err)
		}
		if !canAny {
			return nil, s.errPerm(op, p, opts...)
		}
		entry.MarkIncomplete()
	}
	return entry, nil
}

// lookup implements Lookup for a parsed path. It is used by Lookup as well as
// by put. If entryMustBeClean is true, the returned entry is guaranteed to have
// valid references in its DirBlocks.
func (s *server) lookup(op string, p path.Parsed, entryMustBeClean bool, opts ...options) (*upspin.DirEntry, error) {
	o, ss := subspan("lookup", opts)
	defer ss.End()

	tree, err := s.loadTreeFor(p.User(), o)
	if err != nil {
		return nil, errors.E(op, err)
	}
	entry, dirty, err := tree.Lookup(p)
	if err != nil {
		// This could be ErrFollowLink so return the entry as well.
		return entry, err
	}
	if dirty && entryMustBeClean {
		// Flush and repeat.
		err = tree.Flush()
		if err != nil {
			return nil, errors.E(op, err)
		}
		entry, dirty, err = tree.Lookup(p)
		if err != nil {
			return nil, errors.E(op, err)
		}
		if dirty {
			return nil, errors.E(op, errors.Internal, errors.Str("flush didn't clean entry"))
		}
	}
	if entry.IsLink() {
		return entry, upspin.ErrFollowLink
	}
	return entry, nil
}

// Put implements upspin.DirServer.
func (s *server) Put(entry *upspin.DirEntry) (*upspin.DirEntry, error) {
	const op = "dir/server.Put"
	o, m := newOptMetric(op)
	defer m.Done()

	err := valid.DirEntry(entry)
	if err != nil {
		return nil, errors.E(op, err)
	}
	p, err := path.Parse(entry.Name)
	if err != nil {
		return nil, errors.E(op, entry.Name, err)
	}
	ownerCreatingSnapshotRoot := false
	// TODO: isSnapshotUser and isSnapshotOwner are calling parse
	// repeatedly. Clean up.
	if isSnapshotUser(p.User()) {
		if !isSnapshotOwner(s.userName, p.User()) {
			// Non-owners can't even see the snapshot.
			return nil, errors.E(op, entry.Name, errPrivate)
		}
		if isSnapshotControlFile(p) {
			err = isValidSnapshotControlEntry(entry)
			if err != nil {
				return nil, errors.E(op, err)
			}
			// Start a snapshot for this user.
			s.snapshotControl <- p.User()
			return entry, nil // Confirm snapshot has been started.
		}
		if !p.IsRoot() {
			// Not root: owner can't mutate anything else.
			return nil, errors.E(op, entry.Name, errReadOnly)
		}
		// Else: isOwner && putting the root -> OK, if root does not
		// exist yet.
		ownerCreatingSnapshotRoot = true
	}

	isAccess := access.IsAccessFile(p.Path())
	isGroup := access.IsGroupFile(p.Path())
	isLink := entry.IsLink()

	// Links can't be named Access or Group.
	if isLink {
		if isAccess || isGroup {
			return nil, errors.E(op, p.Path(), errors.Invalid, errors.Str("link cannot be named Access or Group"))
		}
	}
	// Directories cannot have reserved names.
	if isAccess && entry.IsDir() {
		return nil, errors.E(op, errors.Invalid, entry.Name, errors.Str("cannot make directory named Access"))
	}

	// Special files must use integrity pack (plain text + signature).
	isGroupFile := isGroup && !entry.IsDir()
	if (isGroupFile || isAccess) && entry.Packing != upspin.EEIntegrityPack {
		return nil, errors.E(op, p.Path(), errors.Invalid, errors.Str("must use integrity pack"))
	}

	if isAccess {
		// Validate access files at Put time to detect bad ones early.
		_, err := s.loadAccess(entry, o)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}
	if isGroupFile {
		// Validate group files at Put time to detect bad ones early.
		err = s.loadGroup(p, entry)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// Check for links along the path.
	existingEntry, err := s.lookup(op, p, !entryMustBeClean, o)
	if err == upspin.ErrFollowLink {
		return s.errLink(op, existingEntry, o)
	}

	if errors.Match(errNotExist, err) {
		// OK; entry not found as expected. Can we create it?
		if ownerCreatingSnapshotRoot {
			// OK, can create.
		} else {
			canCreate, _, err := s.hasRight(access.Create, p, o)
			if err == upspin.ErrFollowLink {
				return nil, errors.E(op, p.Path(), errors.Internal, errors.Str("unexpected ErrFollowLink"))
			}
			if err != nil {
				return nil, errors.E(op, err)
			}
			if !canCreate {
				return nil, s.errPerm(op, p, o)
			}
		}
		// New file should have a valid sequence number, if user didn't pick one already.
		if entry.Sequence == upspin.SeqNotExist || entry.Sequence == upspin.SeqIgnore && !entry.IsDir() {
			entry.Sequence = upspin.NewSequence()
		}
	} else if err != nil {
		// Some unexpected error happened looking up path. Abort.
		return nil, errors.E(op, err)
	} else {
		// Error is nil therefore path exists.
		// Check if it's root.
		if p.IsRoot() {
			return nil, errors.E(op, p.Path(), errors.Exist)
		}
		// Check if we can overwrite.
		if existingEntry.IsDir() {
			return nil, errors.E(op, p.Path(), errors.Exist, errors.Str("can't overwrite directory"))
		}
		if entry.IsDir() {
			return nil, errors.E(op, p.Path(), errors.Exist, errors.Str("can't overwrite file with directory"))
		}
		// To overwrite a file, we need Write permission.
		canWrite, _, err := s.hasRight(access.Write, p, o)
		if err == upspin.ErrFollowLink {
			return nil, errors.E(op, p.Path(), errors.Internal, errors.Str("unexpected ErrFollowLink"))
		}
		if err != nil {
			return nil, errors.E(op, err)
		}
		if !canWrite {
			return nil, s.errPerm(op, p, o)
		}
		// If the file is expected not to be there, it's an error.
		if entry.Sequence == upspin.SeqNotExist {
			return nil, errors.E(op, entry.Name, errors.Exist)
		}
		// We also must have the correct sequence number or SeqIgnore.
		if entry.Sequence != upspin.SeqIgnore {
			if entry.Sequence != existingEntry.Sequence {
				return nil, errors.E(op, entry.Name, errors.Invalid, errors.Str("sequence number"))
			}
		}
		// Note: sequence number updates for directories is maintained
		// by the Tree since directory entries are never Put by the
		// user explicitly. Here we adjust the dir entries that the user
		// sent us (those representing files only).
		entry.Sequence = upspin.SeqNext(existingEntry.Sequence)

		// If we're updating an Access file delete it from the cache and
		// let it be re-loaded lazily when needed again.
		if access.IsAccessFile(entry.Name) {
			s.access.Remove(entry.Name)
		}
		// If we're updating a Group file, remove the old one from the
		// access group cache. Let the new one be loaded lazily.
		if access.IsGroupFile(entry.Name) {
			err = access.RemoveGroup(entry.Name)
			if err != nil {
				// Nothing to do but log.
				log.Error.Printf("%s: Error removing group file %s: %s", op, entry.Name, err)
			}
		}
	}

	return s.put(op, p, entry, o)
}

// put performs Put on the user's tree.
func (s *server) put(op string, p path.Parsed, entry *upspin.DirEntry, opts ...options) (*upspin.DirEntry, error) {
	o, ss := subspan("put", opts)
	defer ss.End()

	tree, err := s.loadTreeFor(p.User(), o)
	if err != nil {
		return nil, errors.E(op, err)
	}

	entry, err = tree.Put(p, entry)
	if err == upspin.ErrFollowLink {
		return entry, err
	}
	if err != nil {
		return nil, errors.E(op, p.Path(), err)
	}
	return entry, nil
}

// Glob implements upspin.DirServer.
func (s *server) Glob(pattern string) ([]*upspin.DirEntry, error) {
	const op = "dir/server.Glob"
	o, m := newOptMetric(op)
	defer m.Done()

	lookup := func(name upspin.PathName) (*upspin.DirEntry, error) {
		const op = "dir/server.Lookup"
		o, ss := subspan(op, []options{o})
		defer ss.End()
		return s.lookupWithPermissions(op, name, o)
	}
	listDir := func(dirName upspin.PathName) ([]*upspin.DirEntry, error) {
		const op = "dir/server.listDir"
		o, ss := subspan(op, []options{o})
		defer ss.End()
		return s.listDir(op, dirName, o)
	}

	entries, err := serverutil.Glob(pattern, lookup, listDir)
	if err != nil && err != upspin.ErrFollowLink {
		err = errors.E(op, err)
	}
	return entries, err
}

func (s *server) globWithoutPermissions(pattern string) ([]*upspin.DirEntry, error) {
	const op = "dir/server.globWithoutPermissions"
	o, m := newOptMetric(op)
	defer m.Done()

	lookup := func(name upspin.PathName) (*upspin.DirEntry, error) {
		const op = "dir/server.Lookup"
		o, ss := subspan(op, []options{o})
		defer ss.End()
		p, err := path.Parse(name)
		if err != nil {
			return nil, errors.E(op, name, err)
		}
		return s.lookup(op, p, !entryMustBeClean, o)
	}
	listDir := func(dirName upspin.PathName) ([]*upspin.DirEntry, error) {
		const op = "dir/server.listDir"
		o, ss := subspan(op, []options{o})
		defer ss.End()
		p, err := path.Parse(dirName)
		if err != nil {
			return nil, errors.E(op, dirName, err)
		}
		tree, err := s.loadTreeFor(p.User(), o)
		if err != nil {
			return nil, errors.E(op, err)
		}
		entries, _, err := tree.List(p)
		if err != nil {
			return nil, errors.E(op, err)
		}
		return entries, nil
	}

	entries, err := serverutil.Glob(pattern, lookup, listDir)
	if err != nil && err != upspin.ErrFollowLink {
		err = errors.E(op, err)
	}
	return entries, err
}

// listDir implements serverutil.ListFunc, with an additional options variadic.
// dirName should always be a directory.
func (s *server) listDir(op string, dirName upspin.PathName, opts ...options) ([]*upspin.DirEntry, error) {
	parsed, err := path.Parse(dirName)
	if err != nil {
		return nil, errors.E(op, err)
	}

	tree, err := s.loadTreeFor(parsed.User(), opts...)
	if err != nil {
		return nil, errors.E(op, err)
	}

	canList, canRead := false, false
	if isSnapshotUser(parsed.User()) {
		if !isSnapshotOwner(s.userName, parsed.User()) {
			// Non-owners can't see snapshots.
			return nil, errors.E(op, dirName, errNotExist)
		}
		// Owner can always see everything, regardless of access files.
		canList, canRead = true, true
	} else {
		// Check that we have list rights for any file in the directory.
		canList, _, err = s.hasRight(access.List, parsed, opts...)
		if err != nil {
			// TODO(adg): this error needs sanitizing
			return nil, errors.E(op, dirName, err)
		}
		if !canList {
			return nil, errors.E(op, dirName, errors.Private)
		}
		canRead, _, _ = s.hasRight(access.Read, parsed, opts...)
	}

	if canRead {
		// User wants DirEntries with valid blocks, so we must flush
		// the Tree (we could check if !dirty first, but flush when
		// nothing is dirty is cheap and doing everything again if it
		// was dirty is expensive, so flush now).
		err = tree.Flush()
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// Fetch the directory's contents.
	entries, _, err := tree.List(parsed)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if !canRead {
		for _, e := range entries {
			e.MarkIncomplete()
		}
	}
	return entries, nil
}

// Delete implements upspin.DirServer.
func (s *server) Delete(name upspin.PathName) (*upspin.DirEntry, error) {
	const op = "dir/server.Delete"
	o, m := newOptMetric(op)
	defer m.Done()

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}
	if isSnapshotUser(p.User()) {
		if isSnapshotOwner(s.userName, p.User()) {
			// Owner can't mutate.
			return nil, errors.E(op, name, errReadOnly)
		}
		// Everyone else can't even see it.
		return nil, errors.E(op, name, errPrivate)
	}

	canDelete, link, err := s.hasRight(access.Delete, p, o)
	if err == upspin.ErrFollowLink {
		return s.errLink(op, link, o)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}
	if !canDelete {
		return nil, errors.E(op, name, access.ErrPermissionDenied)
	}

	// Load the tree for this user.
	t, err := s.loadTreeFor(p.User(), o)
	if err != nil {
		return nil, errors.E(op, err)
	}
	entry, err := t.Delete(p)
	if err != nil {
		return entry, err // could be ErrFollowLink.
	}
	// If we just deleted an Access file, remove it from the access cache
	// too.
	if access.IsAccessFile(p.Path()) {
		s.access.Remove(p.Path())
	}
	// If we just deleted a Group file, remove it from the Group cache too.
	if access.IsGroupFile(p.Path()) {
		err = access.RemoveGroup(p.Path())
		if err != nil {
			// Nothing to do but log (it may not have been loaded
			// yet, so it's not an error).
			log.Printf("%s: Error removing group file: %s", op, err)
		}
	}
	// If we just deleted the root, close the tree, remove it from the cache
	// and delete all logs associated with the tree owner.
	if p.IsRoot() {
		if err := s.closeTree(p.User()); err != nil {
			return nil, errors.E(op, name, err)
		}
		if err := tree.DeleteLogs(p.User(), s.logDir); err != nil {
			return nil, errors.E(op, name, err)
		}
	}

	return entry, nil
}

// WhichAccess implements upspin.DirServer.
func (s *server) WhichAccess(name upspin.PathName) (*upspin.DirEntry, error) {
	const op = "dir/server.WhichAccess"
	o, m := newOptMetric(op)
	defer m.Done()

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}

	// Check whether the user has Any right on p.
	hasAny, link, err := s.hasRight(access.AnyRight, p, o)
	if err == upspin.ErrFollowLink {
		return s.errLink(op, link, o)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}
	if !hasAny {
		return nil, errors.E(op, errors.Private, name)
	}

	return s.whichAccess(p, o)
}

// Watch implements upspin.DirServer.Watch.
func (s *server) Watch(name upspin.PathName, order int64, done <-chan struct{}) (<-chan upspin.Event, error) {
	const op = "dir/server.Watch"
	o, m := newOptMetric(op)
	defer m.Done()

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}

	tree, err := s.loadTreeFor(p.User(), o)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Establish a channel with the tree and start a goroutine that filters
	// out requests not visible by the caller.
	treeEvents, err := tree.Watch(p, order, done)
	if err != nil {
		return nil, errors.E(op, err)
	}
	events := make(chan upspin.Event, 1)

	go s.watch(op, treeEvents, events)

	return events, nil
}

// watcher runs in a goroutine reading events from the tree and passing them
// along to the original caller, but first verifying whether the user has rights
// to know about the event.
func (s *server) watch(op string, treeEvents <-chan *upspin.Event, outEvents chan<- upspin.Event) {
	const sendTimeout = time.Minute

	t := time.NewTimer(sendTimeout)
	defer close(outEvents)
	defer t.Stop()

	sendEvent := func(e *upspin.Event) bool {
		// Send e on outEvents, with a timeout.
		if !t.Stop() {
			<-t.C
		}
		t.Reset(sendTimeout)
		select {
		case outEvents <- *e:
			// OK, sent.
			return true
		case <-t.C:
			// Timed out.
			log.Printf("%s: timeout sending event for %s", op, s.userName)
			return false
		}
	}

	for {
		e, ok := <-treeEvents
		if !ok {
			// Tree closed channel. Close outgoing event as well.
			return
		}
		if e.Entry == nil {
			// It's likely an error. Pass it along. We're sure to
			// have treeEvents closed in the next loop.
			outEvents <- *e
			continue
		}

		// Check permissions on e.Entry.
		p, err := path.Parse(e.Entry.Name)
		if err != nil {
			sendEvent(&upspin.Event{Error: errors.E(op, err)})
			return
		}
		if isSnapshotUser(p.User()) && isSnapshotOwner(s.userName, p.User()) {
			// Okay to watch.
			if !sendEvent(e) {
				return
			}
			continue
		}
		hasAny, _, err := s.hasRight(access.AnyRight, p)
		if err != nil {
			sendEvent(&upspin.Event{Error: errors.E(op, err)})
			return
		}
		if !hasAny {
			continue
		}
		hasRead, _, err := s.hasRight(access.Read, p)
		if err != nil {
			sendEvent(&upspin.Event{Error: errors.E(op, err)})
			return
		}
		if !hasRead {
			e.Entry.MarkIncomplete()
		}
		if !sendEvent(e) {
			return
		}
	}
}

// Dial implements upspin.Dialer.
func (s *server) Dial(ctx upspin.Config, e upspin.Endpoint) (upspin.Service, error) {
	const op = "dir/server.Dial"
	if e.Transport == upspin.Unassigned {
		return nil, errors.E(op, errors.Invalid, errors.Str("transport must not be unassigned"))
	}
	if err := valid.UserName(ctx.UserName()); err != nil {
		return nil, errors.E(op, errors.Invalid, err)
	}

	cp := *s // copy of the generator instance.
	// Override userName (rest is "global").
	cp.userName = ctx.UserName()
	// create a default Access file for this user and cache it.
	defaultAccess, err := access.New(upspin.PathName(cp.userName + "/"))
	if err != nil {
		return nil, errors.E(op, err)
	}
	cp.defaultAccess.Add(cp.userName, defaultAccess)
	return &cp, nil
}

// Endpoint implements upspin.Service.
func (s *server) Endpoint() upspin.Endpoint {
	// TODO: to be removed.
	return s.serverConfig.DirEndpoint()
}

// Ping implements upspin.Service.
func (s *server) Ping() bool {
	return true
}

// Close implements upspin.Service.
func (s *server) Close() {
	const op = "dir/server.Close"

	// Remove this user's tree from the cache. This allows it to be
	// garbage-collected even if other servers have pointers into the
	// cache (which at least one will have, the one created with New).
	if err := s.closeTree(s.userName); err != nil {
		// TODO: return an error when Close expects it.
		log.Error.Printf("%s: Error closing user tree %q: %q", op, s.userName, err)
	}
}

func (s *server) closeTree(user upspin.UserName) error {
	mu := s.userLock(s.userName)
	mu.Lock()
	defer mu.Unlock()

	if t, ok := s.userTrees.Remove(user).(*tree.Tree); ok {
		// Close will flush and release all resources.
		if err := t.Close(); err != nil {
			return err
		}
	}
	return nil
}

// loadTreeFor loads the user's tree, if it exists.
func (s *server) loadTreeFor(user upspin.UserName, opts ...options) (*tree.Tree, error) {
	defer span(opts).StartSpan("loadTreeFor").End()

	if err := valid.UserName(user); err != nil {
		return nil, errors.E(errors.Invalid, err)
	}

	mu := s.userLock(user)
	mu.Lock()
	defer mu.Unlock()

	// Do we have a cached tree for this user already?
	if val, found := s.userTrees.Get(user); found {
		if tree, ok := val.(*tree.Tree); ok {
			return tree, nil
		}
		// This should never happen because we only store type tree.Tree in the userTree.
		return nil, errors.E(user, errors.Internal,
			errors.Errorf("userTrees contained value of unexpected type %T", val))
	}
	// User is not in the cache. Load a tree from the logs, if they exist.
	hasLog, err := tree.HasLog(user, s.logDir)
	if err != nil {
		return nil, err
	}
	if !hasLog && !s.canCreateRoot(user) {
		// Tree for user does not exist and the logged-in user is not
		// allowed to create it.
		return nil, errNotExist
	}
	log, logIndex, err := tree.NewLogs(user, s.logDir)
	if err != nil {
		return nil, err
	}
	// If user has root, we can load the tree from it.
	if _, err := logIndex.Root(); err != nil {
		// Likely the user has no root yet.
		if !errors.Match(errNotExist, err) {
			// No it's some other error. Abort.
			return nil, err
		}
		// Ok, let it proceed. The  user will still need to make the
		// root, but we allow setting up a new tree for now.
		err = logIndex.SaveOffset(0)
		if err != nil {
			return nil, err
		}
		// Fall through and load a new tree.
	}
	// Create a new tree for the user.
	tree, err := tree.New(s.serverConfig, log, logIndex)
	if err != nil {
		return nil, err
	}
	// Add to the cache and return
	s.userTrees.Add(user, tree)
	return tree, nil
}

// canCreateRoot reports whether the current user can create a root for the
// named user.
func (s *server) canCreateRoot(user upspin.UserName) bool {
	if s.userName == user {
		return true
	}
	if isSnapshotUser(user) && isSnapshotOwner(s.userName, user) {
		return true
	}
	return false
}

// errPerm checks whether the user has any right to the given path, and if so
// returns a Permission error. Otherwise it returns a NotExist error.
// This is used to prevent probing of the name space.
func (s *server) errPerm(op string, p path.Parsed, opts ...options) error {
	// Before returning, check that the user has the right to know,
	// to prevent leaking the name space.
	if hasAny, _, err := s.hasRight(access.AnyRight, p, opts...); err != nil {
		// Some error other than ErrFollowLink.
		return errors.E(op, err)
	} else if !hasAny {
		// User does not have Any right. Return a 'Private' error.
		return errors.E(op, p.Path(), errors.Private)
	}
	return errors.E(op, p.Path(), access.ErrPermissionDenied)
}

// errLink checks whether the user has any right to the given entry, and if so
// returns the entry and ErrFollowLink. If the use has no rights, it returns a
// NotExist error. This is used to prevent probing of the name space using
// links.
func (s *server) errLink(op string, link *upspin.DirEntry, opts ...options) (*upspin.DirEntry, error) {
	p, err := path.Parse(link.Name)
	if err != nil {
		return nil, errors.E(op, errors.Internal, link.Name, err)
	}
	if hasAny, _, err := s.hasRight(access.AnyRight, p, opts...); err != nil {
		// Some error other than ErrFollowLink.
		return nil, errors.E(op, err)
	} else if hasAny {
		// User has Any right on the link. Let them follow it.
		return link, upspin.ErrFollowLink
	}
	// Denied. User has no right on link. Return a 'Private' error.
	return nil, errors.E(op, p.Path(), errors.Private)
}

// shutdown is called when the server is being forcefully shut down.
func (s *server) shutdown() {
	it := s.userTrees.NewIterator()
	for {
		k, v, next := it.GetAndAdvance()
		if !next {
			break
		}
		user := k.(upspin.UserName)
		tree := v.(*tree.Tree)
		err := tree.Close()
		if err != nil {
			log.Error.Printf("dir/server.shutdown: Error closing tree for user %s: %s", user, err)
		}
	}
}

// newOptMetric creates a new options populated with a metric for operation op.
func newOptMetric(op string) (options, *metric.Metric) {
	m, sp := metric.NewSpan(op)
	opts := options{
		span: sp,
	}
	return opts, m
}

// span returns the first span found in opts or a new one if not found.
func span(opts []options) *metric.Span {
	for _, o := range opts {
		if o.span != nil {
			return o.span
		}
	}
	// This is probably an error. Metrics should be created at the entry
	// points only.
	return metric.New("FIXME").StartSpan("FIXME")
}

// subspan creates a span for an operation op in the given option. It returns
// a new option with the new span, for passing along subfunctions.
func subspan(op string, opts []options) (options, *metric.Span) {
	s := span(opts).StartSpan(op)
	return options{span: s}, s
}
