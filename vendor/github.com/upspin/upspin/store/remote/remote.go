// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package remote implements an inprocess store server that uses RPC to
// connect to a remote store server.
package remote // import "upspin.io/store/remote"

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"upspin.io/bind"
	"upspin.io/errors"
	"upspin.io/log"
	"upspin.io/rpc"
	"upspin.io/upspin"
	"upspin.io/upspin/proto"
)

// dialConfig contains the destination and authenticated user of the dial.
type dialConfig struct {
	endpoint upspin.Endpoint
	userName upspin.UserName
}

// remote implements upspin.StoreServer.
type remote struct {
	rpc.Client // For sessions, Ping, and Close.
	cfg        dialConfig

	// If non-empty, the base HTTP URL under which references for this
	// server may be found.
	baseURL string
}

var _ upspin.StoreServer = (*remote)(nil)

// Get implements upspin.StoreServer.Get.
func (r *remote) Get(ref upspin.Reference) ([]byte, *upspin.Refdata, []upspin.Location, error) {
	op := r.opf("Get", "%q", ref)

	if r.baseURL != "" {
		// If we can fetch this by HTTP, do so.
		u := r.baseURL + string(ref)
		resp, err := http.Get(u)
		if err != nil {
			return nil, nil, nil, op.error(err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, nil, nil, op.error(errors.Errorf("fetching %s: %s", u, resp.Status))
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, nil, nil, op.error(err)
		}
		refData := &upspin.Refdata{
			Reference: ref,
			Volatile:  false,
			Duration:  0,
		}
		return body, refData, nil, nil
	}

	req := &proto.StoreGetRequest{
		Reference: string(ref),
	}
	resp := new(proto.StoreGetResponse)
	if err := r.Invoke("Store/Get", req, resp, nil, nil); err != nil {
		return nil, nil, nil, op.error(err)
	}
	if len(resp.Error) != 0 {
		return nil, nil, nil, errors.UnmarshalError(resp.Error)
	}
	return resp.Data, proto.UpspinRefdata(resp.Refdata), proto.UpspinLocations(resp.Locations), nil
}

// Put implements upspin.StoreServer.Put.
func (r *remote) Put(data []byte) (*upspin.Refdata, error) {
	op := r.opf("Put", "%v bytes", len(data))

	req := &proto.StorePutRequest{
		Data: data,
	}
	resp := new(proto.StorePutResponse)
	if err := r.Invoke("Store/Put", req, resp, nil, nil); err != nil {
		return nil, op.error(err)
	}
	return proto.UpspinRefdata(resp.Refdata), op.error(errors.UnmarshalError(resp.Error))
}

// Delete implements upspin.StoreServer.Delete.
func (r *remote) Delete(ref upspin.Reference) error {
	op := r.opf("Delete", "%q", ref)

	req := &proto.StoreDeleteRequest{
		Reference: string(ref),
	}
	resp := new(proto.StoreDeleteResponse)
	if err := r.Invoke("Store/Delete", req, resp, nil, nil); err != nil {
		return op.error(err)
	}
	return op.error(errors.UnmarshalError(resp.Error))
}

// Endpoint implements upspin.StoreServer.Endpoint.
func (r *remote) Endpoint() upspin.Endpoint {
	return r.cfg.endpoint
}

func dialCache(op *operation, config upspin.Config, proxyFor upspin.Endpoint) upspin.Service {
	// Are we using a cache?
	ce := config.CacheEndpoint()
	if ce.Transport == upspin.Unassigned {
		return nil
	}

	// Call the cache. The cache is local so don't bother with TLS.
	authClient, err := rpc.NewClient(config, ce.NetAddr, rpc.NoSecurity, proxyFor)
	if err != nil {
		// On error dial direct.
		op.error(errors.IO, err)
		return nil
	}

	return &remote{
		Client: authClient,
		cfg: dialConfig{
			endpoint: proxyFor,
			userName: config.UserName(),
		},
	}
}

// Dial implements upspin.Service.
func (r *remote) Dial(config upspin.Config, e upspin.Endpoint) (upspin.Service, error) {
	op := r.opf("Dial", "%q, %q", config.UserName(), e)

	if e.Transport != upspin.Remote {
		return nil, op.error(errors.Invalid, errors.Str("unrecognized transport"))
	}

	// First try a cache
	if svc := dialCache(op, config, e); svc != nil {
		return svc, nil
	}

	// Call the server directly.
	authClient, err := rpc.NewClient(config, e.NetAddr, rpc.Secure, upspin.Endpoint{})
	if err != nil {
		return nil, op.error(errors.IO, err)
	}

	r2 := &remote{
		Client: authClient,
		cfg: dialConfig{
			endpoint: e,
			userName: config.UserName(),
		},
	}
	if err := r2.probeDirect(); err != nil {
		op.error(err)
	}

	return r2, nil
}

// probeDirect performs a Get request to the remote server for the reference
// httpBaseRef. The server may respond with an HTTP URL that may be used as a
// base for fetching objects directly by HTTP (from Google Cloud Storage, for
// instance).
func (r *remote) probeDirect() error {
	const op = "store/remote.probeDirect"

	b, _, _, err := r.Get(upspin.HTTPBaseMetadata)
	if errors.Match(errors.E(errors.NotExist), err) {
		return nil
	} else if err != nil {
		return errors.E(op, err)
	}
	s := string(b)

	u, err := url.Parse(s)
	if err != nil {
		return errors.E(op, errors.Errorf("parsing %q: %v", s, err))
	}

	// We have a valid URL. Use it as a base.
	r.baseURL = u.String()
	return nil
}

const transport = upspin.Remote

func init() {
	r := &remote{} // uninitialized until Dial time.
	bind.RegisterStoreServer(transport, r)
}

func (r *remote) opf(method string, format string, args ...interface{}) *operation {
	ep := r.cfg.endpoint.String()
	s := fmt.Sprintf("store/remote.%s(%q)", method, ep)
	op := &operation{s, fmt.Sprintf(format, args...)}
	log.Debug.Print(op)
	return op
}

type operation struct {
	op   string
	args string
}

func (op *operation) String() string {
	return fmt.Sprintf("%s(%s)", op.op, op.args)
}

func (op *operation) error(args ...interface{}) error {
	if len(args) == 0 {
		panic("error called with zero args")
	}
	if len(args) == 1 {
		if e, ok := args[0].(error); ok && e == upspin.ErrFollowLink {
			return e
		}
		if args[0] == nil {
			return nil
		}
	}
	log.Debug.Printf("%v error: %v", op, errors.E(args...))
	return errors.E(append([]interface{}{op.op}, args...)...)
}
