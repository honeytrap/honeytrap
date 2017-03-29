// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tree implements a tree whose nodes are DirEntry entries.
package tree

// TODO: fine-grained locking; metrics; performance tuning.

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"upspin.io/client/clientutil"
	"upspin.io/errors"
	"upspin.io/log"
	"upspin.io/pack"
	"upspin.io/path"
	"upspin.io/upspin"
	"upspin.io/valid"
)

// node is an internal representation of a node in the tree.
// All node accesses must be protected the tree's mutex.
type node struct {
	// entry is the DirEntry this node represents.
	entry upspin.DirEntry

	// watchers is the set of watchers for this node.
	watchers []*watcher

	// kids maps a path element of a path name to the dir entries that represent them.
	// It is empty if this node's dirEntry represents a file or an empty directory;
	// if it represents a directory, kids holds the memory-loaded subdir nodes
	// (not all subdir nodes may be in-memory at a given time).
	kids map[string]*node

	// dirty indicates whether this node's DirEntry has been modified
	// since it was last written to the store.
	dirty bool
}

// Tree is a representation of a directory tree for a single Upspin user.
// The tree reads and writes from/to its backing Store server, which is
// configured when instantiating the Tree. It uses a Log to log changes not
// yet committed to the Store.
type Tree struct {
	// mu protects all accesses to the tree and its nodes and must
	// be held when calling all unexported methods.
	mu sync.Mutex

	user     upspin.UserName
	config   upspin.Config
	packer   upspin.Packer
	log      *Log
	logIndex *LogIndex
	root     *node
	// dirtyNodes is the set of dirty nodes, grouped by path length.
	// The index of the slice is the path length of the nodes therein.
	// The value of the map is ignored.
	dirtyNodes []map[*node]bool
}

// String implements fmt.Stringer.
// t.mu must be held.
func (n *node) String() string {
	return fmt.Sprintf("node: %q, dirty: %v, kids: %d", n.entry.Name, n.dirty, len(n.kids))
}

// New creates an empty Tree using the server's config, a Log and a
// LogIndex for a particular user's tree. Config is used for contacting
// StoreServer, defining the default packing and setting the server name.
// All fields of the config must be defined. Log manipulates the log on behalf
// of the tree for a user. LogIndex is used by Tree to track the most recent
// changes stored in the log for the user. The user name in Log and LogIndex
// must be for the exact same user. If there are unprocessed log entries in
// the Log, the Tree's state is recovered from it.
// TODO: Maybe new is doing too much work. Figure out how to break in two without
// returning an inconsistent new tree if log is unprocessed.
func New(config upspin.Config, log *Log, logIndex *LogIndex) (*Tree, error) {
	const op = "dir/server/tree.New"
	if config == nil {
		return nil, errors.E(op, errors.Invalid, errors.Str("config is nil"))
	}
	if log == nil {
		return nil, errors.E(op, errors.Invalid, errors.Str("log is nil"))
	}
	if logIndex == nil {
		return nil, errors.E(op, errors.Invalid, errors.Str("logIndex is nil"))
	}
	if config.StoreEndpoint().Transport == upspin.Unassigned {
		return nil, errors.E(op, errors.Invalid, errors.Str("unassigned store endpoint"))
	}
	if config.KeyEndpoint().Transport == upspin.Unassigned {
		return nil, errors.E(op, errors.Invalid, errors.Str("unassigned key endpoint"))
	}
	if config.Factotum() == nil {
		return nil, errors.E(op, errors.Invalid, errors.Str("factotum is nil"))
	}
	if config.UserName() == "" {
		return nil, errors.E(op, errors.Invalid, errors.Str("username in tree config is empty"))
	}
	if log.User() == "" {
		return nil, errors.E(op, errors.Invalid, errors.Str("username in log is empty"))
	}
	if log.User() != logIndex.User() {
		return nil, errors.E(op, errors.Invalid, errors.Str("username in log and logIndex mismatch"))
	}
	if err := valid.UserName(log.User()); err != nil {
		return nil, errors.E(op, errors.Invalid, err)
	}
	packer := pack.Lookup(config.Packing())
	if packer == nil {
		return nil, errors.E(op, errors.Invalid, errors.Errorf("no packing %s registered", config.Packing()))
	}
	t := &Tree{
		user:     log.User(),
		config:   config,
		packer:   packer,
		log:      log,
		logIndex: logIndex,
	}
	// Do we have entries in the log to process, to recover from a crash?
	err := t.recoverFromLog()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return t, nil
}

// Lookup returns an entry that represents the path. The returned
// DirEntry may or may not have valid references inside. If dirty is
// true, the references are not up-to-date. Calling Flush in a critical
// section prior to Lookup will ensure the entry is not dirty.
//
// If the returned error is ErrFollowLink, the caller should retry the
// operation as outlined in the description for upspin.ErrFollowLink.
// Otherwise in the case of error the returned DirEntry will be nil.
func (t *Tree) Lookup(p path.Parsed) (de *upspin.DirEntry, dirty bool, err error) {
	const op = "dir/server/tree.Lookup"
	t.mu.Lock()
	defer t.mu.Unlock()

	node, _, err := t.loadPath(p)
	if err == upspin.ErrFollowLink {
		return &node.entry, node.dirty, err
	}
	if err != nil {
		return nil, false, errors.E(op, err)
	}
	return node.entry.Copy(), node.dirty, nil
}

// Put puts an entry at path p into the Tree. If the entry exists, it will be
// overwritten.
//
// If the returned error is ErrFollowLink, the caller should retry the
// operation as outlined in the description for upspin.ErrFollowLink
// (with the added step of updating the Name field of the argument
// DirEntry). Otherwise, the returned DirEntry will be the one put.
func (t *Tree) Put(p path.Parsed, de *upspin.DirEntry) (*upspin.DirEntry, error) {
	const op = "dir/server/tree.Put"
	t.mu.Lock()
	defer t.mu.Unlock()

	if p.IsRoot() {
		return de, t.createRoot(p, de)
	}
	node, watchers, err := t.put(p, de)
	if err == upspin.ErrFollowLink {
		return node.entry.Copy(), err
	}
	if err != nil {
		return nil, err
	}
	// Generate log entry.
	logEntry := &LogEntry{
		Op:    Put,
		Entry: *de,
	}
	err = t.log.Append(logEntry)
	if err != nil {
		return nil, errors.E(op, err)
	}
	notifyWatchers(watchers)
	return de.Copy(), nil
}

// put implements the bulk of Tree.Put, but does not append to the log so it
// can be used to recover the Tree's state from the log.
// t.mu must be held.
func (t *Tree) put(p path.Parsed, de *upspin.DirEntry) (*node, []*watcher, error) {
	// If putting a/b/c/d, ensure a/b/c is loaded.
	parentPath := p.Drop(1)
	parent, watchers, err := t.loadPath(parentPath)
	if err == upspin.ErrFollowLink { // encountered a link along the path.
		return parent, watchers, err
	}
	if err != nil {
		return nil, watchers, err
	}
	if parent.entry.IsLink() {
		return parent, watchers, upspin.ErrFollowLink
	}
	// Now add this dirEntry as a new node
	node := &node{
		entry: *de,
	}
	// If any parent watchers were watching this node, move them to this
	// node.
	moveDownWatchers(node, parent)
	err = t.addKid(node, p, parent, parentPath)
	if err != nil {
		return nil, watchers, err
	}
	return node, watchers, nil
}

// PutDir puts a DirEntry representing an existing directory (with existing
// DirBlocks) into the tree at the point represented by dstDir. The last
// element of dstDir must not yet exist. dstDir must not cross a link nor be
// the root directory. It returns the newly put entry.
func (t *Tree) PutDir(dstDir path.Parsed, de *upspin.DirEntry) (*upspin.DirEntry, error) {
	const op = "dir/server/tree.PutDir"
	t.mu.Lock()
	defer t.mu.Unlock()

	if dstDir.IsRoot() {
		// TODO: handle this later. It might come in handy for reinstating an old root.
		return nil, errors.E(op, errors.Invalid, errors.Str("can't PutDir at the root"))
	}

	// Create a synthetic node and load its kids.
	existingEntryNode := &node{
		entry: *de,
	}
	existingEntryNode.entry.Name = dstDir.Path()
	err := t.loadKids(existingEntryNode)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Put the synthetic node into the tree at dst.
	n, watchers, err := t.put(dstDir, &existingEntryNode.entry)
	if err == upspin.ErrFollowLink {
		return nil, errors.E(op, errors.Invalid, dstDir.Path(), errors.Str("path cannot contain a link"))
	}
	if err != nil {
		return nil, errors.E(op, err)
	}
	notifyWatchers(watchers)
	// Flush now to create a new version of the root.
	err = t.flush() // TODO: avoid this. Create a log operation PutDir.
	if err != nil {
		return nil, errors.E(op, err)
	}
	return n.entry.Copy(), nil
}

// addKid adds a node n with path nodePath as the kid of parent, whose path is parentPath.
// t.mu must be held.
func (t *Tree) addKid(n *node, nodePath path.Parsed, parent *node, parentPath path.Parsed) error {
	if !parent.entry.IsDir() {
		return errors.E(errors.NotDir, errors.Errorf("path: %q", parent.entry.Name))
	}
	if parent.kids == nil {
		// This is a directory with no kids. If it's dirty, it's new.
		// If it's not dirty, load kids from Store.
		if parent.dirty {
			parent.kids = make(map[string]*node)
		} else {
			err := t.loadKids(parent)
			if err != nil {
				return err
			}
		}
	}
	nElem := parentPath.NElem()
	if nodePath.Drop(1).Path() != parentPath.Path() {
		err := errors.E(nodePath.Path(), errors.Internal, errors.Str("parent path does match parent of dir path"))
		log.Error.Print(err)
		return err
	}
	// No need to check if it exists. Simply overwrite. DirServer checks these things.
	parent.kids[nodePath.Elem(nElem)] = n
	// Mark entire path as dirty, from the point that needs to be re-packed
	// and up to the root.
	if n.entry.IsDir() && len(n.entry.Blocks) == 0 {
		return t.markDirty(nodePath)
	}
	return t.markDirty(parentPath)
}

// markDirty marks the entire path from root to p as dirty.
// t.mu must be held.
func (t *Tree) markDirty(p path.Parsed) error {
	// Do we have room to track the max path depth in p?
	if n := p.NElem() + 1; len(t.dirtyNodes) < n { // +1 for the root.
		newDirtyNodes := make([]map[*node]bool, n)
		copy(newDirtyNodes, t.dirtyNodes)
		t.dirtyNodes = newDirtyNodes
	}

	// Start with the root.
	n := t.root
	t.setNodeDirtyAt(0, n)

	// Navigate through every element of p.
	var err error
	for i := 0; i < p.NElem(); i++ {
		elem := p.Elem(i)
		n, err = t.loadNode(n, elem)
		if err != nil {
			return err
		}
		// Non-directory entries are never marked dirty by the Tree,
		// only their parents (directories), which have their kids'
		// names and references packed in them.
		if !n.entry.IsDir() {
			err := errors.E(errors.Internal, n.entry.Name, errors.Str("marking non-dir dirty"))
			log.Error.Printf("%s", err)
			return err
		}
		t.setNodeDirtyAt(i+1, n)
	}
	return nil
}

// setNodeDirtyAt sets the node as dirty and adds it to the dirtyNodes list at a given level.
// The dirtyNodes list is expected to be large enough to accommodate level entries.
// t.mu must be held.
func (t *Tree) setNodeDirtyAt(level int, n *node) {
	n.dirty = true
	if t.dirtyNodes[level] == nil {
		t.dirtyNodes[level] = make(map[*node]bool)
	}
	t.dirtyNodes[level][n] = true // repetitions don't matter.
	n.entry.Sequence++
}

// loadPath ensures the tree contains all nodes up to p and returns p's node.
// If any node is not already in memory, it is loaded from the store server.
// If while loading the path a link is discovered, the link is returned and if
// it's not the last element of the path, ErrFollowLink is returned. If the node
// does not exist, loadPath returns a NotExist error and the closest existing
// ancestor of p, if any. Along with node, loadPath also returns all watchers on
// p's node and in all of p's node's ancestors.
// t.mu must be held.
func (t *Tree) loadPath(p path.Parsed) (*node, []*watcher, error) {
	err := t.loadRoot()
	if err != nil {
		return nil, nil, err
	}
	node := t.root
	// Keep track of all of p's ancestors watchers.
	removeDeadWatchers(node)
	watchers := append([]*watcher(nil), node.watchers...)
	for i := 0; i < p.NElem(); i++ {
		child, err := t.loadNode(node, p.Elem(i))
		if errors.Match(errors.E(errors.NotExist), err) {
			return node, watchers, err
		}
		node = child
		if err != nil {
			return node, watchers, err // err could be upspin.ErrFollowLink.
		}
		removeDeadWatchers(node)
		watchers = append(watchers, node.watchers...)
	}
	if node.entry.Name != p.Path() {
		return node, watchers, errors.E(errors.NotExist, p.Path())
	}
	return node, watchers, nil
}

// loadDir loads the contents of a directory's node if it's not already loaded.
// The node must be known to be a directory and cannot be a link.
// t.mu must be held.
func (t *Tree) loadDir(dir *node) error {
	// Must load from store if kids are not loaded.
	if dir.kids == nil && len(dir.entry.Blocks) > 0 {
		err := t.loadKids(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

// loadNode loads a child node of parent with the given path-wise element name,
// loading it from storage if is not already loaded. If the parent node is a
// link, ErrFollowLink is returned, along with the parent node itself.
// t.mu must be held.
func (t *Tree) loadNode(parent *node, elem string) (*node, error) {
	if parent.entry.IsLink() {
		return parent, upspin.ErrFollowLink
	}
	if !parent.entry.IsDir() {
		return nil, errors.E(errors.NotExist, path.Join(parent.entry.Name, elem))
	}
	err := t.loadDir(parent)
	if err != nil {
		return nil, err
	}
	for dirName, node := range parent.kids {
		if elem == dirName {
			return node, nil
		}
	}
	return nil, errors.E(errors.NotExist, path.Join(parent.entry.Name, elem))
}

// loadKids loads all kids of a parent node from the Store.
// t.mu must be held.
func (t *Tree) loadKids(parent *node) error {
	data, err := clientutil.ReadAll(t.config, &parent.entry)
	if err != nil {
		return err
	}
	err = t.loadKidsFromBlock(parent, data)
	if err != nil {
		return err
	}
	return nil
}

// loadRoot loads the root into memory if it is not already loaded.
// t.mu must be held.
func (t *Tree) loadRoot() error {
	if t.root != nil {
		return nil
	}
	rootDirEntry, err := t.logIndex.Root()
	if err != nil {
		return err
	}
	if rootDirEntry == nil {
		return errors.E(errors.NotExist, t.user)
	}
	t.root = &node{
		entry: *rootDirEntry,
	}
	return nil
}

// createRoot creates the root at p using the given dir entry. A root must not already exist.
// t.mu must be held.
func (t *Tree) createRoot(p path.Parsed, de *upspin.DirEntry) error {
	const op = "dir/server/tree.createRoot"
	// Do we have a root already?
	if t.root != nil {
		// Root already exists.
		return errors.E(op, errors.Exist, errors.Str("root already created"))
	}
	// Check that we're trying to create a root for the owner of the Tree only.
	if p.User() != t.user {
		return errors.E(p.User(), p.Path(), errors.Invalid, errors.Str("can't create root for another user"))
	}
	_, err := t.logIndex.Root()
	if err != nil && !errors.Match(errors.E(errors.NotExist), err) {
		// Error reading the root.
		return err
	}
	// To be sure, the log must be empty too (or t.root wouldn't be empty).
	if t.log.LastOffset() != 0 {
		err := errors.E(errors.Internal, errors.Str("index not empty, but root not found"))
		log.Error.Print(err)
		return err
	}
	// Finally let's create it.
	node := &node{
		entry: *de,
	}
	t.root = node
	err = t.markDirty(p)
	if err != nil {
		return err
	}
	// The root of the tree must be flushed immediately or its recovery
	// becomes cumbersome. Nothing else exists prior to a root existing,
	// so only the root will be flushed.
	log.Printf("Created root: %s", p)
	return t.flush()
}

// List lists the contents of a prefix. If prefix names a directory, all
// entries of the directory are returned. If prefix names a file, that
// file's entry is returned. List does not interpret wildcards. Dirty
// reports whether any DirEntry returned is dirty (and thus may contain
// outdated references).
//
// If the returned error is upspin.ErrFollowLink, the caller should
// retry the operation as outlined in the description for
// upspin.ErrFollowLink. (And in that case, only one DirEntry will be
// returned, that of the link itself.)
func (t *Tree) List(prefix path.Parsed) ([]*upspin.DirEntry, bool, error) {
	const op = "dir/server/tree.List"
	t.mu.Lock()
	defer t.mu.Unlock()

	node, _, err := t.loadPath(prefix)
	if err == upspin.ErrFollowLink {
		return []*upspin.DirEntry{node.entry.Copy()}, node.dirty, err
	}
	if err != nil {
		return nil, false, errors.E(op, err)
	}
	if !node.entry.IsDir() {
		return []*upspin.DirEntry{node.entry.Copy()}, node.dirty, nil
	}
	err = t.loadDir(node)
	if err != nil {
		return nil, false, errors.E(op, err)
	}
	dirty := node.dirty
	var entries []*upspin.DirEntry
	for _, n := range node.kids {
		entries = append(entries, n.entry.Copy())
	}
	return entries, dirty, nil
}

// Delete deletes the entry associated with the path. If the path identifies
// a link, Delete will delete the link itself, not its target.
//
// If the returned error is upspin.ErrFollowLink, the caller should
// retry the operation as outlined in the description for
// upspin.ErrFollowLink. (And in that case, the DirEntry will never
// represent the full path name of the argument.) Otherwise, the
// returned DirEntry will be nil whether the operation succeeded
// or not.
func (t *Tree) Delete(p path.Parsed) (*upspin.DirEntry, error) {
	const op = "dir/server/tree.Delete"
	t.mu.Lock()
	defer t.mu.Unlock()

	if p.IsRoot() {
		return nil, t.deleteRoot()
	}

	node, watchers, err := t.delete(p)
	if err == upspin.ErrFollowLink {
		return node.entry.Copy(), err
	}
	if err != nil {
		return nil, errors.E(op, err)
	}
	// Generate log entry.
	logEntry := &LogEntry{
		Op:    Delete,
		Entry: node.entry,
	}
	err = t.log.Append(logEntry)
	if err != nil {
		return nil, errors.E(op, err)
	}
	notifyWatchers(watchers)
	return node.entry.Copy(), err
}

// delete implements the bulk of Tree.Delete, but does not append to the log
// so it can be used to recover from the Tree's state from the log.
// t.mu must be held.
func (t *Tree) delete(p path.Parsed) (*node, []*watcher, error) {
	parentPath := p.Drop(1)
	parent, watchers, err := t.loadPath(parentPath)
	if err == upspin.ErrFollowLink {
		return parent, watchers, err
	}
	if err != nil {
		return nil, watchers, err
	}
	// Load the node of interest, which is the NElem-th element in its
	// parent's path.
	elem := p.Elem(parentPath.NElem())
	node, err := t.loadNode(parent, elem)
	if err == upspin.ErrFollowLink {
		return node, watchers, err
	}
	if err != nil {
		// Can't load parent.
		return nil, watchers, err
	}
	if len(node.kids) > 0 {
		// Node is a non-empty directory.
		return nil, watchers, errors.E(errors.NotEmpty, p.Path())
	}
	// Remove this elem from the parent's kids map.
	// No need to check if it was there -- it wouldn't have loaded if it weren't.
	delete(parent.kids, elem)

	// If node was dirty, there's no need to flush it to Store ever.
	t.removeFromDirtyList(p, node)

	// If there were watchers on this node, move them to the parent.
	parent.watchers = append(parent.watchers, node.watchers...)

	// Update parent: mark it dirty and log its new version.
	err = t.markDirty(parentPath)
	if err != nil {
		// In practice this can't happen, since the entire path is
		// already loaded.
		return nil, watchers, err
	}
	return node, watchers, nil
}

// deleteRoot deletes the root, if it's empty.
// t.mu must be held.
func (t *Tree) deleteRoot() error {
	if t.root == nil {
		return errors.E(errors.NotExist, errors.Str("root does not exist"))
	}
	log.Printf("Deleting root %q", t.root.entry.Name)
	if len(t.root.kids) > 0 {
		// Root is not empty.
		return errors.E(errors.NotEmpty, t.root.entry.Name)
	}
	// Make sure all log entries are saved, because we're about to lose the
	// last reference to them (and they could be backed up by another tree,
	// so they may still be needed -- we can't simply throw all away).
	err := t.flush()
	if err != nil {
		return err
	}
	// We're all caught up now. Hopefully, some other entry somewhere has a
	// link to this root, because we're about to lose it forever.
	err = t.logIndex.DeleteRoot()
	if err != nil {
		return err
	}
	err = t.log.Truncate(0)
	if err != nil {
		return err
	}
	t.root = nil
	return nil
}

// removeFromDirtyList removes a node n at path p from the list of dirty
// nodes, if n was there.
// t.mu must be held.
func (t *Tree) removeFromDirtyList(p path.Parsed, n *node) {
	nElem := p.NElem()
	if nElem >= len(t.dirtyNodes) {
		// Dirty list does not even go this far. Nothing to do.
		return
	}
	m := t.dirtyNodes[nElem]
	delete(m, n)
}

// Flush flushes all dirty dir entries to the Tree's Store.
func (t *Tree) Flush() error {
	const op = "dir/server/tree.Flush"

	t.mu.Lock()
	defer t.mu.Unlock()

	err := t.loadRoot()
	if err != nil {
		return errors.E(op, err)
	}

	return t.flush()
}

// flush flushes all dirty entries.
// t.mu must be held.
func (t *Tree) flush() error {
	const op = "dir/server/tree.Flush"

	if t.root == nil {
		// Nothing to do.
		return nil
	}

	// Flush from highest path depth up to root.
	for i := len(t.dirtyNodes) - 1; i >= 0; i-- {
		m := t.dirtyNodes[i]
		// For each node at level i, flush it.
		for n := range m {
			err := t.store(n)
			if err != nil {
				return errors.E(op, err)
			}
			n.dirty = false
		}
	}
	// Throw away the entire slice of maps.
	t.dirtyNodes = nil

	// TODO: Verify the log had at least the same number of dirty entries
	// (it could have more because of deletes).

	// Save the last index we operated on.
	err := t.logIndex.SaveOffset(t.log.LastOffset())
	if err != nil {
		return errors.E(op, err)
	}

	// Save new root to the log index.
	return t.logIndex.SaveRoot(&t.root.entry)
}

// Close flushes all dirty blocks to Store and releases all resources
// used by the tree. Further uses of the tree will have unpredictable
// results.
func (t *Tree) Close() error {
	const op = "dir/server/tree.Close"
	t.mu.Lock()
	defer t.mu.Unlock()

	err := t.flush()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

// recoverFromLog inspects the LogIndex and the Log and replays the missing
// operations. It can only be called from New.
func (t *Tree) recoverFromLog() error {
	const (
		op        = "dir/server/tree.recoverFromLog"
		batchSize = 10 // max number of entries to recover at a time.
	)
	lastOffset := t.log.LastOffset()
	lastProcessed, err := t.logIndex.ReadOffset()
	if err != nil {
		return err
	}
	if lastOffset == lastProcessed {
		// All caught up.
		log.Debug.Printf("%s: Tree is all caught up for user %s", op, t.user)
		return nil
	}
	err = t.loadRoot()
	if err != nil {
		return errors.E(op, err)
	}

	// Tree is not current. Replay all entries from the log. Read in chunks
	// of batchSizes entries at a time (a balance between efficiency and
	// how long we want to process the log without checkpointing our state).
	recovered := 0
	next := lastProcessed
	hadError := false
	for {
		log.Debug.Printf("%s: Recovering from log...", op)
		var replay []LogEntry
		replay, next, err = t.log.ReadAt(batchSize, next)
		if err != nil {
			log.Error.Printf("%s: Error in log recovery, possible data loss at offset %d: %s", op, next, err)
			err = t.logIndex.SaveOffset(lastProcessed)
			if err != nil {
				return errors.E(op, err)
			}
			err = t.log.Truncate(next)
			if err != nil {
				return errors.E(op, err)
			}
			hadError = true
		}
		for _, logEntry := range replay {
			de := logEntry.Entry

			p, err := path.Parse(de.Name)
			if err != nil {
				// We don't expect this to fail because
				// de.Name was in the log already and thus
				// has been validated.
				return errors.E(op, err)
			}

			switch logEntry.Op {
			case Put:
				log.Debug.Printf("%s: Putting dirEntry: %q", op, de.Name)
				_, _, err = t.put(p, &de)
			case Delete:
				log.Debug.Printf("%s: Deleting path: %q", op, p.Path())
				_, _, err = t.delete(p)
			default:
				return errors.E(op, errors.Internal, errors.Errorf("no such log operation: %v", logEntry.Op))
			}
			if err != nil {
				// Now we're in serious trouble. We can't recover.
				return errors.E(op, t.user, errors.Errorf("can't recover log: %v", err))
			}
		}
		recovered += len(replay)
		if len(replay) < batchSize {
			break
		}
		if hadError {
			break
		}
	}
	log.Debug.Printf("%s: %d entries recovered. Tree is current.", op, recovered)
	log.Debug.Printf("%s: Tree:\n%s\n", op, t)
	return nil
}

// OnEviction implements cache.EvictionNotifier.
func (t *Tree) OnEviction(key interface{}) {
	const op = "dir/server/tree.OnEviction"
	log.Debug.Printf("%s: tree being evicted: %s", op, t.log.User())
	err := t.Flush()
	if err != nil {
		log.Error.Printf("%s: flush: %v", op, err)
	}
}

// String implements fmt.Stringer.
func (t *Tree) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var buf bytes.Buffer
	t.loadRoot()
	fn := func(n *node, level int) error {
		for i := 0; i < level; i++ {
			buf.WriteString("\t")
		}
		buf.WriteString(string(n.entry.Name))
		buf.WriteString("\n")
		return nil
	}
	err := t.traverse(t.root, 0, fn)
	if err != nil {
		fmt.Fprintf(&buf, "error: %s", err)
	}
	return buf.String()
}

// traverse traverses the tree depth-first calling fn on each node. If fn
// returns an error, traverse aborts with that error.
// t.mu must be held.
func (t *Tree) traverse(n *node, level int, fn func(n *node, level int) error) error {
	err := fn(n, level)
	if err != nil {
		return err
	}
	if n.entry.IsDir() && n.kids == nil {
		err := t.loadKids(n)
		if err != nil {
			return err
		}
	}
	if len(n.kids) == 0 {
		return nil
	}
	sortedKids := make([]*node, 0, len(n.kids))
	for _, kid := range n.kids {
		sortedKids = append(sortedKids, kid)
	}
	sort.Sort(nodeSlice(sortedKids))
	for _, kid := range sortedKids {
		err = t.traverse(kid, level+1, fn)
		if err != nil {
			return err
		}
	}
	return nil
}

// For sorting a slice of nodes based on each entry's SignedName.
type nodeSlice []*node

func (p nodeSlice) Len() int           { return len(p) }
func (p nodeSlice) Less(i, j int) bool { return p[i].entry.SignedName < p[j].entry.SignedName }
func (p nodeSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
