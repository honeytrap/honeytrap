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

package config

import (
	"strconv"
	"time"
)

//convertToUint64 wraps the internal int converter
func convertToUint64(target string, def uint64) uint64 {
	fo, err := strconv.Atoi(target)
	if err != nil {
		return def
	}
	return uint64(fo)
}

// MakeDuration should become internal functions, config should return time.Duration
func MakeDuration(target string, def uint64) time.Duration {
	if !elapso.MatchString(target) {
		return time.Duration(def)
	}

	matchs := elapso.FindAllStringSubmatch(target, -1)

	if len(matchs) <= 0 {
		return time.Duration(def)
	}

	match := matchs[0]

	if len(match) < 3 {
		return time.Duration(def)
	}

	dur := time.Duration(convertToUint64(match[1], def))

	mtype := match[2]

	switch mtype {
	case "s":
		log.Infof("Setting %d in seconds", dur)
		return dur * time.Second
	case "mcs":
		log.Infof("Setting %d in Microseconds", dur)
		return dur * time.Microsecond
	case "ns":
		log.Infof("Setting %d in Nanoseconds", dur)
		return dur * time.Nanosecond
	case "ms":
		log.Infof("Setting %d in Milliseconds", dur)
		return dur * time.Millisecond
	case "m":
		log.Infof("Setting %d in Minutes", dur)
		return dur * time.Minute
	case "h":
		log.Infof("Setting %d in Hours", dur)
		return dur * time.Hour
	default:
		log.Infof("Defaul %d to Seconds", dur)
		return time.Duration(dur) * time.Second
	}

}
