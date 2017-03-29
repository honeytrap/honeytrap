// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"strings"
	"testing"

	"upspin.io/cloud/storage"
	"upspin.io/cloud/storage/storagetest"
	"upspin.io/errors"

	// Import needed storage backend.
	_ "upspin.io/cloud/storage/gcs"
)

const (
	expectedRef   = "978F93921702F861CF941AAACE56B83AE17C8F6845FD674263FFF374A2696A4F"
	serverBaseURL = "http://go-download-from-gcp.goog.com"
	contents      = "contents of our file"
)

func TestPutAndGet(t *testing.T) {
	s := newStoreServer(nil)

	refdata, err := s.Put([]byte(contents))
	if err != nil {
		t.Fatal(err)
	}
	ref := refdata.Reference
	if ref != expectedRef {
		t.Errorf("Expected reference %q, got %q", expectedRef, ref)
	}

	data, _, locs, err := s.Get(ref)
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("Expected data to be non-nil")
	}
	if len(locs) != 0 {
		t.Fatalf("Expected 0 location, got %d", len(locs))
	}
	if string(data) != contents {
		t.Errorf("Got data %q, want %q", data, contents)
	}
}

func TestDelete(t *testing.T) {
	s := newStoreServer(nil)

	err := s.Delete(expectedRef)
	if err != nil {
		t.Fatal(err)
	}
	gotRef := s.storage.(*testGCP).deletedRef
	if gotRef != expectedRef {
		t.Errorf("Expected delete call to %q, got %q", gotRef, expectedRef)
	}
}

// Test some error conditions.

func TestGetInvalidRef(t *testing.T) {
	s := newStoreServer(nil)

	_, _, _, err := s.Get("bla bla bla")
	if err == nil {
		t.Fatal("Expected error")
	}
	want := errors.E(errors.NotExist)
	if !errors.Match(want, err) {
		t.Errorf("Expected error %q, got %q", want, err)
	}
}

func TestNew(t *testing.T) {
	_, err := New("dance=the macarena")
	if err == nil {
		t.Fatalf("Expected error")
	}
	expected := "invalid operation"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected %q, got %q", expected, err)
	}

	_, err = New("backend=disk,dance=the macarena")
	if err == nil {
		t.Fatalf("Expected error")
	}
	expected = "invalid operation"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected %q, got %q", expected, err)
	}

	if testing.Short() {
		t.Skip("skipping part of test when network unavailable; depends on credential availability")
	}
	_, err = New("backend=GCS", "defaultACL=publicRead", "gcpBucketName=zee bucket")
	if err != nil {
		t.Fatal(err)
	}
}

func newStoreServer(s storage.Storage) *server {
	if s == nil {
		s = &testGCP{
			Storage: &storagetest.ExpectDownloadCapturePut{
				Ref:  []string{expectedRef},
				Data: [][]byte{[]byte(contents)},
			},
		}
	}
	return &server{
		storage: s,
	}
}

type storeTestServer struct {
	server *server
}

type testGCP struct {
	storage.Storage
	deletedRef string
}

// Delete implements GCP.
func (t *testGCP) Delete(ref string) error {
	t.deletedRef = ref // Capture the ref
	return nil
}
