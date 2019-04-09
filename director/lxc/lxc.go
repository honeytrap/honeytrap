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
package lxc

import (
	"time"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("director/lxc")

type Delays struct {
	FreezeDelay      Delay `toml:"freeze_every"`
	StopDelay        Delay `toml:"stop_every"`
	HousekeeperDelay Delay `toml:"housekeeper_every"`
}

// Delay defines a duration type.
type Delay time.Duration

// Duration returns the type of the giving duration from the provided pointer.
func (t *Delay) Duration() time.Duration {
	return time.Duration(*t)
}

// UnmarshalText handles unmarshalling duration values from the provided slice.
func (t *Delay) UnmarshalText(text []byte) error {
	s := string(text)

	d, err := time.ParseDuration(s)
	if err != nil {
		log.Errorf("Error parsing duration (%s): %s", s, err.Error())
		return err
	}

	*t = Delay(d)
	return nil
}
