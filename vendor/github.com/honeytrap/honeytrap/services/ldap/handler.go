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
package ldap

import (
	"errors"
	"fmt"

	ber "github.com/go-asn1-ber/asn1-ber"
)

type requestHandler interface {
	handle(*ber.Packet, eventLog) []*ber.Packet
}

type resultCodeHandler struct {
	replyTypeID   int64 // the overall type of the response, e.g. 1 is BindResponse
	resultCode    int64 // the result code, i.e. 0 is success, 49 is invalid credentials, etc.
	matchedDN     string
	diagnosticMsg string
}

//Handle: the message envelope
func (h *resultCodeHandler) handle(p *ber.Packet, el eventLog) []*ber.Packet {

	id, _ := messageID(p)

	reply := replyEnvelope(id)

	bindResult := ber.Encode(
		ber.ClassApplication, ber.TypeConstructed, ber.Tag(h.replyTypeID), nil, "Response")
	bindResult.AppendChild(
		ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, h.resultCode, "Result Code"))
	// per the spec these are "matchedDN" and "diagnosticMessage", but we don't need them for this
	bindResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, h.matchedDN, "matched DN"))
	bindResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, h.diagnosticMsg, "Diagnostic message"))

	reply.AppendChild(bindResult)

	return []*ber.Packet{reply}
}

func replyEnvelope(msgid int64) *ber.Packet {

	reply := ber.Encode(
		ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	reply.AppendChild(
		ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, msgid, "MessageId"))

	return reply
}

func isUnbindRequest(p *ber.Packet) bool {

	if len(p.Children) > 1 {
		err := checkPacket(p.Children[1], ber.ClassApplication, ber.TypePrimitive, 0x02)
		if err == nil {
			return true
		}
	}
	return false
}

func messageID(p *ber.Packet) (int64, error) {

	// check overall packet header
	err := checkPacket(p, ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence)
	if err != nil {
		return 0, err
	}

	// check type of message id
	if len(p.Children) < 1 {
		return 0, errors.New("ldap: invalid request message")
	}

	err = checkPacket(p.Children[0], ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger)
	if err != nil {
		return 0, err
	}

	// return the message id
	return forceInt64(p.Children[0].Value), nil
}

//checkPacket: check a ber packet for correct class, type and tag
func checkPacket(p *ber.Packet, cl ber.Class, ty ber.Type, ta ber.Tag) error {
	if p.ClassType != cl {
		return fmt.Errorf("Check packet: Incorrect class, expected %v but got %v", cl, p.ClassType)
	}
	if p.TagType != ty {
		return fmt.Errorf("Check packet: Incorrect type, expected %v but got %v", cl, p.TagType)
	}
	if p.Tag != ta {
		return fmt.Errorf("Check packet: Incorrect tag, expected %v but got %v", cl, p.Tag)
	}

	return nil
}

// readVersion: Return the LDAP major version from the message
func readVersion(p *ber.Packet) int64 {

	if len(p.Children) > 0 && len(p.Children[1].Children) > 0 {
		err := checkPacket(p.Children[1].Children[0], ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger)
		if err != nil {
			log.Debugf("Error can not read the ldap version: %s", err)
			return -1
		}
		return forceInt64(p.Children[1].Children[0].Value)
	}

	log.Debug("Error can not read the ldap version")

	return -1
}

func forceInt64(v interface{}) int64 {
	switch v := v.(type) {
	case int64:
		return v
	case uint64:
		return int64(v)
	case int32:
		return int64(v)
	case uint32:
		return int64(v)
	case int:
		return int64(v)
	case byte:
		return int64(v)
	default:
		log.Panicf("forceInt64 doesn't understand values of type: %t", v)
	}
	return 0
}
