// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package perm

import (
	"testing"

	"upspin.io/access"
	"upspin.io/errors"
	"upspin.io/test/testenv"
	"upspin.io/upspin"
)

func setupStoreEnv(t *testing.T) (store *Store, ownerEnv *testenv.Env, wait, cleanup func()) {
	ownerEnv, wait, cleanup = setupEnv(t)

	var err error
	store, err = WrapStore(ownerEnv.Config, readyNow, ownerEnv.StoreServer)
	if err != nil {
		t.Fatal(err)
	}

	return
}

func TestStoreNoGroupFileAllowsAll(t *testing.T) {
	store, _, wait, cleanup := setupStoreEnv(t)
	defer cleanup()

	wait()

	// Everyone is allowed.
	for _, user := range []upspin.UserName{
		owner,
		writer,
		"foo@bar.com",
		"nobody@nobody.org",
	} {
		if !store.perm.IsWriter(user) {
			t.Errorf("user %q is not allowed; expected allowed", user)
		}
	}
}

func TestStoreAllowsOnlyOwner(t *testing.T) {
	store, ownerEnv, wait, cleanup := setupStoreEnv(t)
	defer cleanup()

	r := testenv.NewRunner()
	r.AddUser(ownerEnv.Config)

	r.As(owner)
	r.MakeDirectory(groupDir)
	r.Put(writersGroup, owner) // Only owner can write.
	if r.Failed() {
		t.Fatal(r.Diag())
	}

	wait() // Update call
	wait() // Watch event

	// Owner is allowed.
	if !store.perm.IsWriter(owner) {
		t.Errorf("Owner is not allowed, expected allowed")
	}

	// No one else is allowed.
	for _, user := range []upspin.UserName{
		writer,
		"foo@bar.com",
		"nobody@nobody.org",
	} {
		if store.perm.IsWriter(user) {
			t.Errorf("user %q is allowed; expected not allowed", user)
		}
	}
}

func TestStoreIncludeRemoteGroups(t *testing.T) {
	store, ownerEnv, wait, cleanup := setupStoreEnv(t)
	defer cleanup()

	writerEnv, err := testenv.New(&testenv.Setup{
		OwnerName: writer,
		Packing:   upspin.PlainPack,
		Kind:      "inprocess",
	})
	if err != nil {
		t.Fatal(err)
	}

	r := testenv.NewRunner()
	r.AddUser(ownerEnv.Config)
	r.AddUser(writerEnv.Config)

	const (
		randomDude = "random@dude.io"

		ownersContents = owner + ", otherGroupFile"

		otherGroupFile     = groupDir + "/otherGroupFile"
		otherGroupContents = writer + "/Group/family"

		writerGroupDir            = writer + "/Group"
		writerAccessFile          = writer + "/Group/Access"
		writerAccessContents      = "r: " + access.All
		writerFamilyGroupFile     = writer + "/Group/family"
		writerFamilyGroupContents = writer + "," + randomDude
	)

	r.As(writer)
	r.MakeDirectory(writerGroupDir)
	r.Put(writerAccessFile, writerAccessContents)
	r.Put(writerFamilyGroupFile, writerFamilyGroupContents)
	if r.Failed() {
		t.Fatal(r.Diag())
	}

	r.As(owner)
	r.MakeDirectory(groupDir)
	r.Put(otherGroupFile, otherGroupContents)
	r.Put(writersGroup, ownersContents)
	if r.Failed() {
		t.Fatal(r.Diag())
	}

	wait() // Update call
	wait() // Watch event

	// owner, writer and randomDude are allowed.
	for _, user := range []upspin.UserName{
		owner,
		writer,
		randomDude,
	} {
		if !store.perm.IsWriter(user) {
			t.Errorf("user %q is not allowed; expected allowed", user)
		}
	}

	// No one else is allowed.
	for _, user := range []upspin.UserName{
		"all@upspin.io",
		"foo@bar.com",
		"god@heaven.infinite",
		"nobody@nobody.org",
	} {
		if store.perm.IsWriter(user) {
			t.Errorf("user %q is allowed; expected not allowed", user)
		}
	}

	writerEnv.Exit()
}

func TestStoreLifeCycle(t *testing.T) {
	store, ownerEnv, wait, cleanup := setupStoreEnv(t)
	defer cleanup()

	writerEnv, err := testenv.New(&testenv.Setup{
		OwnerName: writer,
		Packing:   upspin.PlainPack,
		Kind:      "inprocess",
	})
	if err != nil {
		t.Fatal(err)
	}

	r := testenv.NewRunner()
	r.AddUser(ownerEnv.Config)
	r.AddUser(writerEnv.Config)

	wait()

	// Everyone is allowed at first.
	for _, user := range []upspin.UserName{
		owner,
		writer,
		"foo@bar.com",
		"nobody@nobody.org",
	} {
		if !store.perm.IsWriter(user) {
			t.Errorf("user %q is not allowed; expected allowed", user)
		}
	}

	r.As(owner)
	r.MakeDirectory(groupDir)
	r.Put(writersGroup, "*@example.com") // Anyone at example.com is allowed.
	if r.Failed() {
		t.Fatal(r.Diag())
	}

	wait()

	// Owner continues to be allowed, as well as others in the domain.
	for _, user := range []upspin.UserName{
		owner,
		"fred@example.com",
		"shirley@example.com",
	} {
		if !store.perm.IsWriter(user) {
			t.Errorf("User %s is not allowed, expected allowed", user)
		}
	}

	// But no one else is allowed.
	for _, user := range []upspin.UserName{
		writer,
		"foo@bar.com",
		"nobody@nobody.org",
	} {
		if store.perm.IsWriter(user) {
			t.Errorf("user %q is allowed; expected not allowed", user)
		}
	}

	writerEnv.Exit()
}

func TestStoreIntegration(t *testing.T) {
	ownerStore, ownerEnv, wait, cleanup := setupStoreEnv(t)
	defer cleanup()

	writerConfig, err := ownerEnv.NewUser(writer)
	if err != nil {
		t.Fatal(err)
	}

	r := testenv.NewRunner()
	r.AddUser(ownerEnv.Config)

	wait()

	// Dial the same server endpoint for writer.
	srv, err := ownerStore.Dial(writerConfig, ownerEnv.Config.StoreEndpoint())
	if err != nil {
		t.Fatal(err)
	}
	writerStore := srv.(upspin.StoreServer)

	// Everyone is allowed at first.
	for _, store := range []upspin.StoreServer{
		ownerStore,
		writerStore,
	} {
		ref, err := store.Put([]byte("data"))
		if err != nil {
			t.Fatal(err)
		}
		err = store.Delete(ref.Reference)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Allow only owner.
	r.As(owner)
	r.MakeDirectory(groupDir)
	r.Put(writersGroup, owner) // Only owner can write.
	if r.Failed() {
		t.Fatal(r.Diag())
	}

	wait()

	// Writing as owner succeeds.
	ref1, err := ownerStore.Put([]byte("123"))
	if err != nil {
		t.Fatal(err)
	}

	// Writing as other fails.
	_, err = writerStore.Put([]byte("456"))
	expectedErr := errors.E(errors.Permission, upspin.UserName(writer))
	if !errors.Match(expectedErr, err) {
		t.Fatalf("err = %s, want = %s", err, expectedErr)
	}

	// Deleting as other fails.
	err = writerStore.Delete(ref1.Reference)
	if !errors.Match(expectedErr, err) {
		t.Fatalf("err = %s, want = %s", err, expectedErr)
	}

	// Deleting as owner succeeds.
	err = ownerStore.Delete(ref1.Reference)
	if err != nil {
		t.Fatal(err)
	}
}
