// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

// TODOs:
// - The watcher "tails" the log, starting from a given order number. It is
//   done in a goroutine because the order can be very far from current state
//   and we don't want to block the caller until all such state is sent on the
//   Event channel. However, once the watcher has caught up with the current
//   state of the Tree, there's no longer a need for a goroutine or for reading
//   the log directly (and thus spend time in disk I/O, unmarshalling, etc). We
//   can simply note that the end of file was reached, quit the goroutine and
//   send events as they come in. This requires some extra synchronization code
//   and ensuring that sending does not block the Tree (we can keep the
//   goroutine if we don't want to impose a short timeout on the channel).

import (
	"sync/atomic"
	"time"

	"upspin.io/errors"
	"upspin.io/log"
	"upspin.io/path"
	"upspin.io/upspin"
)

const (
	// watcherTimeout is the timeout to send notifications on a watcher
	// channel. It happens in a goroutine, so it's safe to hang for a while.
	watcherTimeout = 1 * time.Minute
)

var (
	errTimeout  = errors.E(errors.IO, errors.Str("channel operation timed out"))
	errClosed   = errors.E(errors.IO, errors.Str("channel closed"))
	errNotExist = errors.E(errors.NotExist)
)

// watcher holds together the done channel and the event channel for a given
// watch point.
type watcher struct {
	// The path name this watcher watches.
	path path.Parsed

	// events is the Event channel with the client. It is write-only.
	events chan *upspin.Event

	// done is the client's done channel. When it's closed, this watcher
	// dies.
	done <-chan struct{}

	// hasWork is an internal channel that the Tree uses to tell the watcher
	// goroutine to look for work at the end of the log.
	hasWork chan bool

	// log is a read-only cloned instance of the Tree's log that keeps track
	// of this watcher's progress.
	log *Log

	// closed indicates whether the watcher is closed (1) or open (0).
	// It must be loaded and stored atomically.
	closed int32
}

// Watch implements upspin.DirServer.Watch.
func (t *Tree) Watch(p path.Parsed, order int64, done <-chan struct{}) (<-chan *upspin.Event, error) {
	const op = "dir/server/tree.Watch"

	t.mu.Lock()
	defer t.mu.Unlock()

	// Clone the logs so we can keep reading it while the current tree
	// continues to be updated (we're about to unlock this tree).
	cLog, err := t.log.Clone()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Create a watcher, but do not attach it to any node yet.
	// TODO: limit number of watchers on any given node/tree?
	ch := make(chan *upspin.Event)
	w := &watcher{
		path:    p,
		events:  ch,
		done:    done,
		hasWork: make(chan bool, 1),
		log:     cLog,
		closed:  0,
	}

	if order == -1 {
		// Send the current state first. We must flush the tree so we
		// know our logs are current (or we need to recover the tree
		// from the logs).
		err := t.flush()
		if err != nil {
			return nil, errors.E(op, err)
		}

		// Make a copy of the tree so we have an immutable tree in
		// memory, at a fixed log position.
		cIndex, err := t.logIndex.Clone()
		if err != nil {
			return nil, errors.E(op, err)
		}
		offset := t.log.LastOffset()
		clone := &Tree{
			user:     t.user,
			config:   t.config,
			packer:   t.packer,
			log:      cLog,
			logIndex: cIndex,
		}
		// Start sending the current state of the cloned tree and setup
		// the watcher for this tree once the current state is sent.
		go w.sendCurrentAndWatch(clone, t, p, offset)
	} else {
		// Set up the notification hook.
		err = t.addWatcher(p, w)
		if err != nil {
			return nil, errors.E(op, err)
		}

		// Start the watcher.
		go w.watch(order)
	}

	return w.events, nil
}

// addWatcher adds a watcher to the node at a given path location.
// t.mu must be held.
func (t *Tree) addWatcher(p path.Parsed, w *watcher) error {
	n, _, err := t.loadPath(p)
	if err != nil && !errors.Match(errNotExist, err) {
		return err
	}
	if err != nil && n == nil {
		return err
	}
	// Add watcher to node, or to an ancestor.
	n.watchers = append(n.watchers, w)
	return nil
}

// sendCurrentAndWatch takes an original tree and its clone and sends the state
// of the clone starting from the subtree rooted at p. The offset refers to the
// last log offset saved by the original tree.
// It must run in a goroutine. Errors are logged.
func (w *watcher) sendCurrentAndWatch(clone, orig *Tree, p path.Parsed, offset int64) {
	const op = "dir/server/tree.sendCurrentAndWatch"

	n, _, err := clone.loadPath(p)
	if err != nil && !errors.Match(errNotExist, err) {
		log.Error.Printf("%s: error loading path: %s", op, err)
		w.sendError(err)
		w.close()
		return
	}
	// If p exists, traverse the sub-tree and send its current state on the
	// events channel.
	if err == nil {
		fn := func(n *node, level int) error {
			logEntry := &LogEntry{
				Op:    Put,
				Entry: n.entry,
			}
			err = w.sendEvent(logEntry, offset)
			if err != nil {
				return err
			}
			return nil
		}
		err = clone.traverse(n, 0, fn)
		if err != nil {
			log.Error.Printf("%s: error traversing tree: %s", op, err)
			w.sendError(err)
			w.close()
			return
		}
	}
	// Set up the notification hook on the original tree. We must lock it.
	orig.mu.Lock()
	err = orig.addWatcher(p, w)
	orig.mu.Unlock()
	if err != nil {
		log.Error.Printf("%s: error adding watcher: %s", op, err)
		w.sendError(err)
		w.close()
		return
	}
	// Start the watcher (in this goroutine -- don't start a new one here).
	w.watch(offset)
}

// sendEvent sends a single logEntry read from the log at offset position
// to the event channel. If the channel blocks for longer than watcherTimeout,
// the operation fails and the watcher is invalidated (marked for deletion).
func (w *watcher) sendEvent(logEntry *LogEntry, offset int64) error {
	var event *upspin.Event
	// Strip block information for directories. We avoid an extra copy
	// if it's not a directory.
	if logEntry.Entry.IsDir() {
		entry := logEntry.Entry
		entry.MarkIncomplete()
		event = &upspin.Event{
			Order:  offset,
			Delete: logEntry.Op == Delete,
			Entry:  &entry, // already a copy.
		}
	} else {
		event = &upspin.Event{
			Order:  offset,
			Delete: logEntry.Op == Delete,
			Entry:  &logEntry.Entry, // already a copy.
		}
	}
	select {
	case w.events <- event:
		// Event was sent.
		return nil
	case <-w.done:
		// Client is done receiving events.
		return errClosed
	case <-time.After(watcherTimeout):
		// TODO: time.After leaks. Use NewTimer.
		// Oops. Client didn't read fast enough.
		return errTimeout
	}
}

func (w *watcher) sendError(err error) {
	e := &upspin.Event{
		Error: err,
	}
	select {
	case w.events <- e:
		// Error event was sent.
	case <-time.After(3 * watcherTimeout):
		// Can't send another error since we timed out again. Log an
		// error and close the watcher.
		log.Error.Printf("dir/server/tree.sendError: %s", errTimeout)
	}
}

// sendEventFromLog sends notifications to the given watcher for all
// descendant entries of a target path, reading from the given log starting at a
// given offset until it reaches the end of the log. It returns the next offset
// to read.
func (w *watcher) sendEventFromLog(offset int64) (int64, error) {
	const op = "dir/server/tree.sendEventFromLog"
	curr := offset
	for {
		// Is the receiver still interested in reading events?
		select {
		case <-w.done:
			return 0, errClosed
		default:
		}

		logs, next, err := w.log.ReadAt(1, curr)
		if err != nil {
			return next, errors.E(op, errors.Invalid, errors.Errorf("cannot read log at order %d: %v", curr, err))
		}
		if len(logs) != 1 {
			// End of log.
			return next, nil
		}
		curr = next
		logEntry := logs[0]
		path := logEntry.Entry.SignedName
		if !isPrefixPath(path, w.path) {
			// Not a log of interest.
			continue
		}
		err = w.sendEvent(&logEntry, curr)
		if err != nil {
			return 0, err
		}
	}
}

// watch, which runs in a goroutine, reads from the log starting at a given
// offset and sends notifications on the event channel until the end of the log
// is reached. It waits to be notified of more work or until the client's
// done channel is closed, in which case it terminates.
func (w *watcher) watch(offset int64) {
	defer w.close()
	for {
		var err error
		offset, err = w.sendEventFromLog(offset)
		if err != nil {
			if err != errTimeout && err != errClosed {
				log.Error.Printf("watch: sending error to client: %s", err)
				w.sendError(err)
			}
			return
		}
		select {
		case <-w.done:
			// Done channel was closed. Close watcher and quit this
			// goroutine.
			return
		case <-w.hasWork:
			// Wake up and work from where we left off.
		}
	}
}

// isClosed reports whether this watcher has been closed.
func (w *watcher) isClosed() bool {
	return atomic.LoadInt32(&w.closed) == 1
}

// close closes the watcher. Must only be called internally by the watcher's
// goroutine.
func (w *watcher) close() {
	atomic.StoreInt32(&w.closed, 1)
	close(w.events)
}

// removeDeadWatchers removes all watchers on a node that have closed their done
// or Event channels.
func removeDeadWatchers(n *node) {
	curr := 0
	for i := 0; i < len(n.watchers); i++ {
		doneCh := n.watchers[i].done

		closed := n.watchers[i].isClosed()
		if !closed {
			// If the done channel is ready, it's been closed.
			select {
			case <-doneCh:
				closed = true
			default:
			}
		}
		if closed {
			// Remove this entry. If there are more, simply copy the
			// next one over the ith entry. Otherwise just shrink
			// the slice.
			if i > curr {
				n.watchers[curr] = n.watchers[i]
			}
			continue
		}
		curr++
	}
	n.watchers = n.watchers[:curr]
}

// moveDownWatchers moves watchers from the parent to the node if and only if
// the parent watcher is watching node.
func moveDownWatchers(node, parent *node) {
	curr := 0
	p, _ := path.Parse(node.entry.Name) // err can't happen.
	for i := 0; i < len(parent.watchers); i++ {
		w := parent.watchers[i]
		if w.path.NElem() < p.NElem() || w.path.First(p.NElem()).Path() == p.Path() {
			curr++
			continue
		}
		// Remove this watcher from the parent and add it to the node.
		// If there are more watchers beyond curr, simply copy the
		// next ith watcher to the curr watcher. Otherwise just shrink
		// the slice.
		// Note: The node is newly-put, so it does not have watchers yet
		// and hence there's not need to look for duplicate watchers
		// here.
		node.watchers = append(node.watchers, w)
		if i > curr {
			parent.watchers[curr] = parent.watchers[i]
		}
	}
	parent.watchers = parent.watchers[:curr]
}

// notifyWatchers tells all watchers there are new entries in the log to be
// processed.
func notifyWatchers(watchers []*watcher) {
	for _, w := range watchers {
		select {
		case w.hasWork <- true:
			// OK, sent.
		default:
			// Watcher is busy. It will get to it eventually.
		}
	}
}

// newIsPrefixPath reports whether the path has a pathwise prefix.
func isPrefixPath(name upspin.PathName, prefix path.Parsed) bool {
	parsed, err := path.Parse(name)
	if err != nil {
		log.Error.Print("dir/server/tree.isPrefixPath: error parsing path", name)
		return false
	}
	return parsed.HasPrefix(prefix)
}
