// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package storagetest implements simple types and utility functions to help test
// implementations of storage.S.
package storagetest // import "upspin.io/cloud/storage/storagetest"

import (
	"upspin.io/cloud/storage"
	"upspin.io/errors"
)

// DummyStorage returns a storage.Storage that does nothing.
func DummyStorage(*storage.Opts) (storage.Storage, error) {
	return &dummyStorage{}, nil
}

type dummyStorage struct{}

var _ storage.Storage = (*dummyStorage)(nil)

func (m *dummyStorage) LinkBase() (base string, err error)    { return "", nil }
func (m *dummyStorage) Download(ref string) ([]byte, error)   { return nil, nil }
func (m *dummyStorage) Put(ref string, contents []byte) error { return nil }
func (m *dummyStorage) Delete(ref string) error               { return nil }
func (m *dummyStorage) Close()                                {}

// ExpectDownloadCapturePut inspects all calls to Download with the
// given Ref and if it matches, it returns Data. Ref matches are strictly sequential.
// It also captures all Put requests.
type ExpectDownloadCapturePut struct {
	dummyStorage
	// Expectations for calls to Download
	Ref  []string
	Data [][]byte
	// Storage for calls to Put
	PutRef      []string
	PutContents [][]byte

	pos int // position of the next Ref to match
}

// Download implements storage.Storage.
func (e *ExpectDownloadCapturePut) Download(ref string) ([]byte, error) {
	if e.pos < len(e.Ref) && ref == e.Ref[e.pos] {
		data := e.Data[e.pos]
		e.pos++
		return data, nil
	}
	return nil, errors.E(errors.NotExist)
}

// Put implements storage.Storage.
func (e *ExpectDownloadCapturePut) Put(ref string, contents []byte) error {
	e.PutRef = append(e.PutRef, ref)
	e.PutContents = append(e.PutContents, contents)
	return nil
}
