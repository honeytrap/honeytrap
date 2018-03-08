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
package storers

import "github.com/honeytrap/honeytrap/event"

type StoredFile struct {
	Reference string
	Backend string
}

func (f *StoredFile) ToMap() map[string]interface{} {
	out := make(map[string]interface{})
	out["reference"] = f.Reference
	out["backend"] = f.Backend
	return out
}

func (f *StoredFile) ToEvent() event.Event {
	var out event.Event
	out.Store("reference", f.Reference)
	out.Store("backend", f.Backend)
	return out
}

type Storer interface {
	Push([]byte) StoredFile
	Fetch(StoredFile) ([]byte, error)
}