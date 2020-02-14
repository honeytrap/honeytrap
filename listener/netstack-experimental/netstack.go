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
package xnetstack

import (
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	logging "github.com/op/go-logging"
)

// todo
// port /whitelist filtering (8022)
// custom (rst, irs )
// arm

var (
	SensorNetstack = event.Sensor("netstack")
)

var log = logging.MustGetLogger("listener/netstack")

var (
	_ = listener.Register("netstack-experimental", New)
)
