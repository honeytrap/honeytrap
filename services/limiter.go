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
package services

import (
	"net"
	"time"

	"golang.org/x/time/rate"

	"sync"
)

func NewLimiter() *Limiter {
	return &Limiter{
		interval: rate.Every(time.Minute * 10),
		burst:    4,
	}
}

type Limiter struct {
	m sync.Map

	interval rate.Limit
	burst    int
}

func (l *Limiter) Allow(ip net.Addr) bool {
	limiter := rate.NewLimiter(l.interval, l.burst)

	if ta, ok := ip.(*net.TCPAddr); ok {
		v, _ := l.m.LoadOrStore(ta.IP.String(), limiter)
		return v.(*rate.Limiter).Allow()
	} else if ua, ok := ip.(*net.UDPAddr); ok {
		v, _ := l.m.LoadOrStore(ua.IP.String(), limiter)
		return v.(*rate.Limiter).Allow()
	} else {
		return false
	}
}
