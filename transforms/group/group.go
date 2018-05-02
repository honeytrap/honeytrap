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
package group

import (
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/transforms"
)

func GroupBy(features []string, period time.Duration) transforms.TransformFunc {
	// hashmap that counts occurrences
	memory := make(map[string]uint)
	// used to synchronize updates and deletions
	locks := make(map[string]sync.Mutex)
	return func(state transforms.State, e event.Event, send func(event.Event)) {
		hash := ""
		reference := make(map[string]string)
		// transfer only the features (keys) we're interested in
		for _, feat := range features {
			hash += feat + "=" + e.Get(feat) + ";"
			reference[feat] = e.Get(feat)
		}
		if _, ok := locks[hash]; !ok {
			locks[hash] = sync.Mutex{}
		}
		lock := locks[hash]
		lock.Lock()
		if _, ok := memory[hash]; !ok {
			memory[hash] = 0
			time.AfterFunc(period, func() {
				lock.Lock()
				evt := event.New(
					event.Sensor("transforms"),
					event.Category("group"),
					event.Custom("group.count", memory[hash]),
				)
				for key, val := range reference {
					evt.Store("original-" + key, val)
				}
				send(evt)
				delete(memory, hash)
				lock.Unlock()
			})
		}
		memory[hash]++
		lock.Unlock()
	}
}