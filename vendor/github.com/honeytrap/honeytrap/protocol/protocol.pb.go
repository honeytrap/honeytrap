// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package protocol

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type PingMessage struct {
	Token            *string `protobuf:"bytes,1,req,name=token" json:"token,omitempty"`
	LocalAddress     *string `protobuf:"bytes,2,req,name=localAddress" json:"localAddress,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PingMessage) Reset()         { *m = PingMessage{} }
func (m *PingMessage) String() string { return proto.CompactTextString(m) }
func (*PingMessage) ProtoMessage()    {}

func (m *PingMessage) GetToken() string {
	if m != nil && m.Token != nil {
		return *m.Token
	}
	return ""
}

func (m *PingMessage) GetLocalAddress() string {
	if m != nil && m.LocalAddress != nil {
		return *m.LocalAddress
	}
	return ""
}

type PayloadMessage struct {
	Token            *string `protobuf:"bytes,1,req,name=token" json:"token,omitempty"`
	LocalAddress     *string `protobuf:"bytes,2,req,name=localAddress" json:"localAddress,omitempty"`
	RemoteAddress    *string `protobuf:"bytes,3,req,name=remoteAddress" json:"remoteAddress,omitempty"`
	Protocol         *string `protobuf:"bytes,4,req,name=protocol" json:"protocol,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PayloadMessage) Reset()         { *m = PayloadMessage{} }
func (m *PayloadMessage) String() string { return proto.CompactTextString(m) }
func (*PayloadMessage) ProtoMessage()    {}

func (m *PayloadMessage) GetToken() string {
	if m != nil && m.Token != nil {
		return *m.Token
	}
	return ""
}

func (m *PayloadMessage) GetLocalAddress() string {
	if m != nil && m.LocalAddress != nil {
		return *m.LocalAddress
	}
	return ""
}

func (m *PayloadMessage) GetRemoteAddress() string {
	if m != nil && m.RemoteAddress != nil {
		return *m.RemoteAddress
	}
	return ""
}

func (m *PayloadMessage) GetProtocol() string {
	if m != nil && m.Protocol != nil {
		return *m.Protocol
	}
	return ""
}
