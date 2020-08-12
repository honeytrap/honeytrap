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

package server

import "net"

//Controller defines the generic properties for controller
type Controller interface {
	Name() string
	Device() (string, error)
	// TODO:
	// Maybe we should have the container be responsible for itself, so idle timing etc.
	// that way we can create also a simple honeypot (low interaction) controller only with stream as well
	// then IsIdle, Device, Name, Etc are not important anymore, or we can also just solve this
	// with having an other SSHProxyListener, so that will be solved already then....
	IsIdle() bool
	Dial(string) (net.Conn, error)
	CleanUp() error
}
