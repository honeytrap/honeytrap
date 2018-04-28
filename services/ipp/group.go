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

import "github.com/honeytrap/honeytrap/services/decoder"

type attribGroup struct {
	tag byte //begin-attribute-group-tag
	val []ValueType
}

func (ag *attribGroup) decode(dec decoder.Decoder) error {

	for vtag := dec.Byte(); vtag > unsupAttribTag; vtag = dec.Byte() { //Not a delimiter tag
		if err := dec.LastError(); err != nil {
			return err
		}

		var v ValueType

		switch vtag {
		case valInteger:
			v = &valInt{tag: vtag}
		case valBoolean:
			v = &valBool{tag: vtag}
		case valKeyword:
			v = &valStr{tag: vtag}
		case valCharSet:
			v = &valStr{tag: vtag}
		case valURI:
			v = &valStr{tag: vtag}
		case valRangeOfInt:
			v = &valInt{tag: vtag}
		case naturelLang:
			v = &valStr{tag: vtag}
		case mimeMediaType:
			v = &valStr{tag: vtag}
		case textWithoutLang:
			v = &valStr{tag: vtag}
		case valEnum:
			v = &valInt{tag: vtag}
		case nameWithoutLang:
			v = &valStr{tag: vtag}
		}

		v.decode(dec)
		ag.val = append(ag.val, v)
	}

	// Put back last read byte, because it is a delimiter tag
	dec.Seek(-1)

	return nil
}

func (ag *attribGroup) encode(buf decoder.EncoderType) error {
	buf.WriteUint8(ag.tag)

	if ag.val != nil {
		for _, vals := range ag.val {
			vals.encode(buf)
		}
	}
	return nil
}

var model = &attribGroup{
	tag: printerAttribTag,
	val: []ValueType{
		&valStr{valKeyword, "compression-supported", []string{"none"}},
		&valRangeInt{valRangeOfInt, "copies-supported", int32(1), int32(1)},
		&valStr{mimeMediaType, "document-format-supported", []string{
			"application/octet-stream",
			"image/pwg-raster",
			"application/pdf",
		}},
		&valStr{nameWithoutLang, "marker-colors", []string{"black", "cyan", "magenta", "yellow"}},
		&valInt{valInteger, "marker-high-levels", []int32{100, 100, 100, 100}},
		&valInt{valInteger, "marker-levels", []int32{80, 100, 100, 100}},
		&valInt{valInteger, "marker-low-levels", []int32{10, 10, 10, 10}},
		&valStr{valKeyword, "media-cols-supported", []string{"media-type", "media-size"}},
		&valInt{valEnum, "operations-supported", []int32{2, 4, 11}},
		&valStr{valKeyword, "print-color-mode-supported", []string{"auto", "color", "monochrome"}},
		&valBool{valBoolean, "printer-is-accepting-jobs", []bool{true}},
		&valInt{valEnum, "printer-state", []int32{3}},
		&valStr{valKeyword, "printer-state-reasons", []string{"none"}},
	},
}
