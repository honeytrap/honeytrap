// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dircache

// This file defines structures that keep track of individual target directories.
// It particular it keeps a count of entries from the directory still in the LRU
// and handles refreshing of directory entries.

import (
	"strings"
	"sync"
	"time"

	"upspin.io/bind"
	"upspin.io/errors"
	"upspin.io/log"
	"upspin.io/path"
	"upspin.io/upspin"
)

// proxiedDir contains information about a proxied user directoies.
type proxiedDir struct {
	l     *clog
	ep    *upspin.Endpoint // endpoint for directory server
	order int64            // last order seen in watch
	atime time.Time        // time of last access
	user  upspin.UserName

	die   chan bool // channel used to tell watcher to die
	dying chan bool // channel used to confirm watcher is dying
}

// proxiedDirs is used to translate between a user name and the relevant cached directory.
type proxiedDirs struct {
	sync.Mutex

	closing bool // when this is true do not allocate any new watchers
	l       *clog
	m       map[upspin.UserName]*proxiedDir
}

func newProxiedDirs(l *clog) *proxiedDirs {
	return &proxiedDirs{m: make(map[upspin.UserName]*proxiedDir), l: l}
}

// close terminates all watchers.
func (p *proxiedDirs) close() {
	p.Lock()
	defer p.Unlock()
	if p.closing {
		return
	}
	p.closing = true
	for _, d := range p.m {
		d.close()
	}
}

// proxyFor saves the endpoint and makes sure it is being watched.
func (p *proxiedDirs) proxyFor(name upspin.PathName, ep *upspin.Endpoint) {
	p.Lock()
	defer p.Unlock()
	if p.closing {
		return
	}

	parsed, err := path.Parse(name)
	if err != nil {
		log.Info.Printf("parse error on a cleaned name: %s", name)
		return
	}
	u := parsed.User()
	d := p.m[u]
	if d == nil {
		d = &proxiedDir{l: p.l, ep: ep, user: u}
		p.m[u] = d
	}

	// Remember when we last accessed this proxied directory.
	// TODO: Use this time to stop listening to directories we
	// haven't looked at in a long time. We will also have to
	// forget about cached information for them if we stop
	// watching.
	d.atime = time.Now()

	// If the endpoint changed, kill off the current watcher.
	if d.ep != nil && *d.ep != *ep {
		d.close()
	}

	// Start a watcher if none is running.
	d.ep = ep
	if d.die == nil {
		d.die = make(chan bool)
		d.dying = make(chan bool)
		go d.watcher()
	}
}

// setOrder remembers an order read from the logfile.
func (p *proxiedDirs) setOrder(name upspin.PathName, order int64) {
	p.Lock()
	defer p.Unlock()
	if p.closing {
		return
	}

	parsed, err := path.Parse(name)
	if err != nil {
		log.Info.Printf("parse error on a cleaned name: %s", name)
		return
	}
	u := parsed.User()
	d := p.m[u]
	if d == nil {
		d = &proxiedDir{l: p.l, user: u}
		p.m[u] = d
	}
	d.order = order
}

// close terminates the goroutines associated with a proxied dir.
func (d *proxiedDir) close() {
	if d.die != nil {
		close(d.die)
		<-d.dying
		d.die = nil
	}
}

// watcher watches a directory and caches any changes to something already in the LRU.
func (d *proxiedDir) watcher() {
	log.Debug.Printf("dircache.Watcher %s %s", d.user, d.ep)
	defer close(d.dying)
	nextLogTime := time.Now()
	// If we don't no better, always read in the whole state. It
	// is shorter than the the history of all operations.
	if d.order == 0 {
		d.order = -1
	}
	lastErr := ""
	seen := 0
	for {
		err := d.watch()
		if err == nil {
			log.Debug.Printf("dircache.Watcher %s %s exiting", d.user, d.ep)
			// watch() only returns if the watcher has been told to die
			// or if there is an error requiring a new Watch.
			return
		}
		if err == upspin.ErrNotSupported {
			// Can't survive this.
			log.Debug.Printf("rpc/dircache.watcher: %s: %s", d.user, err)
			return
		}
		if strings.Contains(err.Error(), "cannot read log at order") {
			// Reread current state.
			d.order = -1
		}
		// Rate limit repeat messages. Otherwise the log will get pretty
		// full when disconnected.
		newErr := err.Error()
		if lastErr == newErr {
			seen++
			if seen > 10 && !time.Now().After(nextLogTime) {
				continue
			}
		} else {
			seen = 0
		}
		log.Info.Printf("rpc/dircache.watcher: %s: %s", d.user, err)
		nextLogTime = time.Now().Add(time.Minute)
		lastErr = newErr
		time.Sleep(time.Second)
	}
}

// watch loops receiving watch events. It returns nil if told to die.
// Otherwise it returns whatever error was encountered.
func (d *proxiedDir) watch() error {
	dir, err := bind.DirServer(d.l.cfg, *d.ep)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	event, err := dir.Watch(upspin.PathName(string(d.user)+"/"), d.order, done)
	if err != nil {
		return err
	}

	// Loop receiving events until we are told to stop or the event stream is closed.
	for {
		select {
		case <-d.die:
			return nil
		case e, ok := <-event:
			if !ok {
				return errors.E("Watch event stream closed")
			}
			if err := d.handleEvent(&e); err != nil {
				return err
			}
		}
	}
}

func (d *proxiedDir) handleEvent(e *upspin.Event) error {
	// Something odd happened?
	if e.Error != nil {
		log.Info.Printf("rpc/dircache.handleEvent: %s", e.Error)
		return e.Error
	}

	// If we are rereading the current state, wipe what we know.
	if d.order == -1 {
		d.l.wipeLog(d.user)
	}

	// Is this a file we are watching?
	log.Debug.Printf("watch entry %s %v", e.Entry.Name, e)
	_, ok := d.l.lru.Get(lruKey{name: e.Entry.Name, glob: false})
	if !ok {
		// Not a file we are watching, how about in a directory we are watching?
		dirName := path.DropPath(e.Entry.Name, 1)
		if dirName == e.Entry.Name {
			return nil
		}
		_, ok := d.l.lru.Get(lruKey{name: dirName, glob: true})
		if !ok {
			return nil
		}
	}

	// This is an event we care about.
	d.order = e.Order
	op := lookupReq
	if e.Delete {
		op = deleteReq
	}
	d.l.logRequestWithOrder(op, e.Entry.Name, nil, e.Entry, e.Order)
	d.l.flush()
	return nil
}
