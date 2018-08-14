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
package mongodb

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const (
	zeroBS     = "00"
	doubleBS   = "01"
	stringBS   = "02"
	documentBS = "03"
	arrayBS    = "04"
	boolBS     = "08"
)

// mechanism := md.rq.(*OpCommandMsg).metadata.elements[1].(stringStruct).name

var mapCommands = map[string]func(*mongodbService, *MsgData, map[string]interface{}) []byte{
	"whatsmyuri":       (*mongodbService).uriMsg,
	"buildinfo":        (*mongodbService).buildInfoMsg,
	"buildInfo":        (*mongodbService).buildInfoMsg,
	"getLog":           (*mongodbService).getLogMsg,
	"isMaster":         (*mongodbService).forShellMsg,
	"replSetGetStatus": (*mongodbService).replSetMsg,
	"listDatabases":    (*mongodbService).listDatabases,
	"saslStart":        (*mongodbService).saslstartMsg,
	"getnonce":         (*mongodbService).getnonceMsg,
	//...
	"saslContinue": (*mongodbService).conversationId,
}

var mapHandShake = map[string]func(*mongodbService, *MsgData) []byte{
	"isMaster": (*mongodbService).isMasterHandshakeMsg,
	"ismaster": (*mongodbService).isMasterHandshakeMsg,
}

func (s *mongodbService) respHandler(md *MsgData, ev map[string]interface{}) []byte {

	src := []byte{}
	if md.mh.OpCode == 2004 {
		if md.cmd == "isMaster" || md.cmd == "ismaster" {
			src = s.isMasterHandshakeMsg(md)
		} else {
			log.Error("OpCode 2004 with command: %s", md.cmd)
			// TODO check what is usually returned
		}
	} else {
		fn, ok := mapCommands[md.cmd]
		if !ok {
			log.Error("Error: command not implemented: %s", md.cmd)
			ev["mongodb.commandNotImplemented"] = md.cmd
			src = errorMsgBadCmd(md)
		} else {
			ev["mongodb.command"] = md.cmd
			src = fn(s, md, ev)
		}
	}

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Error("Error encoding hex response: %s", err.Error())
		//TODO check what to return
	}
	return dst[:n]
}

func (s *mongodbService) parsePayload(md *MsgData, payload []byte) {
	parts := bytes.Split(payload, []byte(","))
	for _, slice := range parts {
		switch {
		case bytes.HasPrefix(slice, []byte("n=")):
			s.Client.username = string(slice[2:])
		case bytes.HasPrefix(slice, []byte("r=")):
			if s.Client.clNonce == "" {
				s.Client.clNonce = string(slice[2:])
			} //  else {
			// 	s.Client.cbNonce = string(slice[2:]) // combinedNonce
			// }
		case bytes.HasPrefix(slice, []byte("p=")):
			s.Client.clProof = string(slice[2:])
			//case bytes.HasPrefix(slice, []byte("a / m / s =")):   we should'nt receive these ones
		}
	}
}

func createPass(password string) string {

	// md5 := md5.Sum([]byte(str))

	credsum := md5.New()
	credsum.Write([]byte("user" + ":mongo:" + "pencil"))
	hex := hex.EncodeToString(credsum.Sum(nil))

	fmt.Println("---------------------------------------------------------")
	fmt.Println(hex)
	fmt.Println("---------------------------------------------------------")
	return hex
}

func (s *mongodbService) userExists(md *MsgData, ev map[string]interface{}) bool {
	for _, Db := range s.Dbs {
		if Db.Name != md.rq.(*OpCommandMsg).database {
			continue
		}
		for _, User := range Db.Users {
			if s.Client.username != User.username {
				continue
			}
			s.Client.password = User.password
			s.Client.salt = User.salt
			return true
		}
		ev["mongodb.failAuth"] = "username not found"
		return false
	}
	ev["mongodb.failAuth"] = "database not found"
	return false
}

func (s *mongodbService) createSvNonce() {
	s.Client.svNonce = "/A+KIDOqoL+95VaItbNN9geuyPE6Hv52"  //TODO generated randomly + ? saved in ev ?
	s.Client.cbNonce = s.Client.clNonce + s.Client.svNonce // combinedNonce
}

func withHeader(md *MsgData, aw *bytes.Buffer) []byte {
	// only db070000
	id := encodeInt(md.mh.RequestID, 4)
	str := aw.String()
	aw.Reset()
	aw.WriteString(id)
	aw.WriteString(id)
	aw.WriteString("db070000")

	aw.WriteString(str)

	l := []byte(encodeInt(int32(4+aw.Len()/2), 4))
	withHeader := append(l, aw.Bytes()...)
	return withHeader
}

func withLengthHeader(aw *bytes.Buffer) []byte {
	l := encodeInt(int32(4+aw.Len()/2-5), 4)
	str := aw.String()
	aw.Reset()
	aw.WriteString(l)
	aw.WriteString(str)
	return aw.Bytes()
}

func (s *mongodbService) authHandShake(md *MsgData) []byte {

	aw := bytes.NewBuffer([]byte(""))
	s.createSvNonce()

	// 	s.Client.cbNonce = string(slice[2:]) // combinedNonce

	src := []byte(s.Client.cbNonce)
	cbnonce := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(cbnonce, src)

	// src := []byte(s.Client.clNonce)
	// clnonce := make([]byte, hex.EncodedLen(len(src)))
	// hex.Encode(clnonce, src)

	// src = []byte(s.Client.svNonce)
	// svnonce := make([]byte, hex.EncodedLen(len(src)))
	// hex.Encode(svnonce, src)

	src = []byte(s.Client.salt)
	salt := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(salt, src)

	payload := "00723d" + string(cbnonce) + "2c733d" + string(salt) + "2c693d3130303030" // 4096: 2c693d34303936
	lenpayload := encodeInt(int32(len(payload)/2)-1, 4)                                  // -1 because of the 00 we add

	aw.WriteString("10636f6e766572736174696f6e4964000100000008646f6e650000057061796c6f616400")
	aw.WriteString(lenpayload)
	aw.WriteString(payload)
	aw.WriteString("016f6b00000000000000f03f000500000000")

	l := encodeInt(int32(4+aw.Len()/2-5), 4)
	str := aw.String()
	aw.Reset()
	aw.WriteString(l)
	aw.WriteString(str)

	// l := []byte(encodeInt(int32(4+aw.Len()/2), 4))
	// withHeader := append(l, aw.Bytes()...)

	// len(aw.Bytes())/2 - 3

	return withHeader(md, aw)
	// return []byte(msg)
}

//////////////

// when saslContinue:
func (s *mongodbService) conversationId(md *MsgData, ev map[string]interface{}) []byte {

	if s.Client == (Client{}) {
		// check if a client exists: if not, what happen ? did someone try to send saslcontinue without saslstart ?
	}

	payload := md.rq.(*OpCommandMsg).metadata.elements[1].(binaryStruct).payload //TODO improve elemnts[i]
	l := md.rq.(*OpCommandMsg).metadata.elements[1].(binaryStruct).length        //TODO improve elemnts[i]

	if l != 0 { //1st salscontinue
		s.parsePayload(md, payload)
		return s.serverSignature(md)
	}
	// 2nd sasl continue with empty payload
	s.parsePayload(md, payload)

	if !s.logged {
		// when we directly send saslcontinue
	}

	return finishAuthShake(md)
}

func (s *mongodbService) serverSignature(md *MsgData) []byte {
	fmt.Println()
	fmt.Println("serverSignature()")

	// serverSign, proved := s.checkProof(md)
	serverSign, proved := s.scram()
	if !proved {
		log.Error("Check proof wrong!")
		return authFailMsg(md)
	}

	s.logged = true

	v := hex.EncodeToString(serverSign)
	l := "6d000000"
	id := encodeInt(md.mh.RequestID, 4)

	a1 := "db0700005800000010636f6e766572736174696f6e4964000100000008646f6e650000057061796c6f6164001e00000000763d"

	a2 := "016f6b00000000000000f03f000500000000"
	msg := l + id + id + a1 + v + a2

	return []byte(msg)
}

func finishAuthShake(md *MsgData) []byte {

	l := "4f000000"
	id := encodeInt(md.mh.RequestID, 4)
	a := "db0700003a00000010636f6e766572736174696f6e4964000100000008646f6e650001057061796c6f6164000000000000016f6b00000000000000f03f000500000000"
	msg := l + id + id + a
	return []byte(msg)
}

func authFailMsg(md *MsgData) []byte {
	msg := "db07000061000000016f6b000000000000000000026572726d7367001700000041757468656e74696361" +
		"74696f6e206661696c65642e0010636f6465001200000002636f64654e616d650015000000417574686" +
		"56e7469636174696f6e4661696c656400000500000000"
	l := "76000000"
	id := encodeInt(md.mh.RequestID, 4)
	msg = l + id + id + msg
	return []byte(msg)
}

func errorMsgBadCmd(md *MsgData) []byte { //TODO ev

	id := encodeInt(md.mh.RequestID, 4)
	l := len(md.cmd)

	// len1
	opCode := "db070000"
	// len2
	a1 := "016f6b000000000000000000026572726d736700"
	// len3
	a11 := "6e6f207375636820636f6d6d616e643a2027"
	a2 := "272c2062616420636d643a20277b20"
	a3 := "3a20312e30207d270010636f6465003b00000002636f64654e616" +
		"d650010000000436f6d6d616e644e6f74466f756e6400000500000000"

	src := []byte(md.cmd)
	cmd := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(cmd, src)

	len3 := encodeInt(int32(42+2*l), 4)
	len2 := encodeInt(int32(111+2*l), 4)
	len1 := encodeInt(int32(132+2*l), 4)

	msg := len1 + id + id + opCode + len2 + a1 + len3 + a11 + string(cmd) + a2 + string(cmd) + a3

	return []byte(msg)
}

func errorMsgNotAuthorized(md *MsgData, c string) []byte {

	a1 := "016f6b000000000000000000026572726d736700"
	a2 := "6e6f7420617574686f72697a6564206f6e2061646d696e20746f206578656375746520636f6d6d616e64207b20"
	a3 := "207d0010636f6465000d00000002636f64654e616d65000d000000556e617574686f72697a656400000500000000"

	l := len(c)
	src := []byte(c)
	cmd := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(cmd, src)

	id := encodeInt(md.mh.RequestID, 4)
	opCode := "db070000"

	len1 := encodeInt(int32(135+l), 4)
	len2 := encodeInt(int32(114+l), 4)
	len3 := encodeInt(int32(48+l), 4)

	msg := len1 + id + id + opCode + len2 + a1 + len3 + a2 + string(cmd) + a3

	return []byte(msg)
}
