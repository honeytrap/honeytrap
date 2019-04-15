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
package profiler

import (
	"net/http"
	_ "net/http/pprof"

	logging "github.com/op/go-logging"
	"github.com/pkg/profile"
)

var log = logging.MustGetLogger("honeytrap/profiler")

type Profiler interface {
	Start()
	Stop()
}

func Dummy() *dummyProfiler {
	return &dummyProfiler{}
}

type dummyProfiler struct {
}

func (p *dummyProfiler) Start() {
}

func (p *dummyProfiler) Stop() {
}

func New(options ...func(*profile.Profile)) *profiler {
	return &profiler{
		options: append(options, profile.ProfilePath("."), profile.NoShutdownHook),
	}
}

type profiler struct {
	p interface {
		Stop()
	}

	options []func(*profile.Profile)
}

func (p *profiler) Start() {
	go func() {
		http.ListenAndServe("127.0.0.1:6060", nil)
	}()

	p.p = profile.Start(p.options...)
	log.Info("Profiler started.")
}

func (p *profiler) Stop() {
	p.p.Stop()
}
