/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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
