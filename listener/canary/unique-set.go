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
package canary

type equalFn func(interface{}, interface{}) bool

// NewUniqueSet returns a new instance of UniqueSet.
func NewUniqueSet(fn equalFn) *UniqueSet {
	return &UniqueSet{
		uniqueFunc: fn,
	}
}

// UniqueSet defines a type to create a unique set of values.
type UniqueSet struct {
	items []interface{}

	uniqueFunc equalFn
}

// Count returns count of all elements.
func (us *UniqueSet) Count() int {
	return len(us.items)
}

// Add adds the given item into the set if it not yet included.
func (us *UniqueSet) Add(item interface{}) interface{} {
	for _, item2 := range us.items {
		if !us.uniqueFunc(item, item2) {
			continue
		}

		return item2
	}

	us.items = append(us.items, item)
	return item
}

// Remove removes the item from the interal set.
func (us *UniqueSet) Remove(item interface{}) {
	for i, item2 := range us.items {
		if item != item2 {
			continue
		}

		us.items[i] = nil

		us.items = append(us.items[:i], us.items[i+1:]...)
		return
	}
}

// Each runs the function against all items in set.
func (us *UniqueSet) Each(fn func(int, interface{})) {
	items := us.items[:]

	for i, item := range items {
		fn(i, item)
	}
}

// Find runs the function against all items in set.
func (us *UniqueSet) Find(fn func(interface{}) bool) interface{} {
	for _, item := range us.items {
		if fn(item) {
			return item
		}
	}

	return nil
}
