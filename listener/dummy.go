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

package listener

import (
	"context"
	"net"
)

func MustDummy(options ...func(Listener) error) Listener {
	l, _ := Dummy()
	return l
}

func Dummy(options ...func(Listener) error) (Listener, error) {
	return &dummyListener{}, nil
}

type dummyListener struct {
}

func (l *dummyListener) Close() error {
	return nil
}

func (l *dummyListener) Start(ctx context.Context) error {
	return nil
}

func (l *dummyListener) Accept() (net.Conn, error) {
	return nil, nil
}
