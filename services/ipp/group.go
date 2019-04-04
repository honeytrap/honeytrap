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
