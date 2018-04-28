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
package ipp

import (
	"testing"

	"github.com/honeytrap/honeytrap/services/decoder"
)

/*
func TestIPP(t *testing.T) {
	s := IPP()
}
*/

func TestValStrDecode(t *testing.T) {
	v := &valStr{
		tag:  textWithoutLang,
		name: "test",
		val:  []string{"one", "two"},
	}

	enc := decoder.NewEncoder()

	v.encode(enc)
	vlen := 20 //length in bytes of v

	if l := enc.Len(); l != vlen {
		t.Errorf("valStr.encode: wrong length. want %v got %v", vlen, l)
	}

	ebytes := enc.Bytes()
	dec := decoder.NewDecoder(ebytes)

	vnew := &valStr{tag: dec.Byte()}

	if dec.Available() != vlen-1 {
		t.Errorf("Decoder: failed read tag. data len is %v, want %v", dec.Available(), vlen-1)
	}
	if err := dec.LastError(); err != nil {
		t.Errorf("decoder.Byte(): wrong offset: %v", err.Error())
	}
	vnew.decode(dec)
	if vnew.name != v.name {
		t.Errorf("Decode, wrong name. Want %v, got %v", v.name, vnew.name)
	}
	if vnew.val[0] != v.val[0] {
		t.Errorf("Decode, wrong value. Want %v, got %v", v.val[0], vnew.val[0])
	}
	if vnew.val[1] != v.val[1] {
		t.Errorf("Decode, wrong value. Want %v, got %v. All:%v", v.val[1], vnew.val[1], vnew.val)
	}
}

func TestGroupDecode(t *testing.T) {

	grp := &attribGroup{
		tag: opAttribTag,
	}
	grp.val = append(grp.val, &valStr{
		tag:  textWithoutLang,
		name: "test",
		val:  []string{"one", "two"},
	})

	enc := decoder.NewEncoder()
	if err := grp.encode(enc); err != nil {
		t.Errorf("Encode length error: %v", err.Error())
	}

	//grp ipp encoded is 21 bytes long
	grplen := 21
	if l := enc.Len(); l != grplen {
		t.Errorf("attribGroup.encode: wrong length. want %v got %v", grplen, l)
	}

	dec := decoder.NewDecoder(enc.Bytes())

	dtag := dec.Byte()
	if dtag != opAttribTag {
		t.Errorf("Decoded wrong group tag. want %X, got %X", opAttribTag, dtag)
	}
	if e := dec.LastError(); e != nil {
		t.Errorf("Decoding length Error: %v", e.Error())
	}

	grpout := &attribGroup{tag: dtag}
	_ = grpout.decode(dec)

	tval, _ := grpout.val[0].(*valStr)
	ival, _ := grp.val[0].(*valStr)
	if tval.name != ival.name {
		t.Errorf("Decode, wrong name. Want %v, got %v", ival.name, tval.name)
	}
	if tval.val[1] != ival.val[1] {
		t.Errorf("Decode, wrong value. Want %v, got %v", ival.val[1], tval.val[1])
	}
}

func TestIPPDecode(t *testing.T) {
	//byte length: 34
	ipp := &ippMsg{
		versionMajor: 2,
		versionMinor: 0,
		statusCode:   123, //bogus code
		requestID:    23,
	}
	grp := &attribGroup{
		tag: opAttribTag,
		val: []ValueType{
			&valStr{textWithoutLang, "test", []string{"one", "two"}},
		},
	}

	ipp.attributes = append(ipp.attributes, grp)
	ipp.attributes = append(ipp.attributes, &attribGroup{tag: endAttribTag})

	enc := ipp.encode()
	ilen := 30
	if l := enc.Len(); l != ilen {
		t.Errorf("IPP Encode: wrong size. Want %v, got %v", ilen, l)
	}

	ippnew := &ippMsg{}

	_ = ippnew.decode(enc.Bytes())

	if ippnew.versionMajor != ipp.versionMajor {
		t.Errorf("IPP Decoding error: versionMajor is %v, want %v", ippnew.versionMajor, ipp.versionMajor)
	}
	if ippnew.requestID != ipp.requestID {
		t.Errorf("IPP Decoding error: requestID is %v, want %v", ippnew.requestID, ipp.requestID)
	}
	if len(ipp.attributes) != len(ippnew.attributes) {
		t.Errorf("IPP Decoding: Amount of groups is %v, want %v", len(ippnew.attributes), len(ipp.attributes))
	}
}
