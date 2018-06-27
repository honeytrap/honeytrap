// Copyright 2016 The Netstack Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcp

import "github.com/google/netstack/tcpip/seqnum"

type Metadata interface {
	IRS() seqnum.Value
}
