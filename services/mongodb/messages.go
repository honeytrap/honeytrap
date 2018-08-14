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
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func (s *mongodbService) isMasterMsg(md *MsgData) []byte {

	elements34 := []elem{
		booleanStruct{0x08, "ismaster", true},
		int32Struct{0x10, "maxBsonObjectSize", 16777216},
		int32Struct{0x10, "maxMessageSizeBytes", 48000000},
		int32Struct{0x10, "maxWriteBatchSize", 1000},
		datetimeStruct{0x09, "localTime", encodelocalTime()},
		int32Struct{0x10, "maxWireVersion", 5},
		int32Struct{0x10, "minWireVersion", 0},
		booleanStruct{0x08, "readOnly", false},
		doubleStruct{0x01, "ok", 61503}}

	// elements36 := []elem{
	// 	booleanStruct{0x08, "ismaster", true},
	// 	int32Struct{0x10, "maxBsonObjectSize", 16777216},
	// 	int32Struct{0x10, "maxMessageSizeBytes", 48000000},
	// 	int32Struct{0x10, "maxWriteBatchSize", 100000},
	// 	datetimeStruct{0x09, "localTime", encodelocalTime()},
	// 	int32Struct{0x10, "logicalSessionTimeoutMinutes", 30},
	// 	int32Struct{0x10, "minWireVersion", 0},
	// 	int32Struct{0x10, "maxWireVersion", 6},
	// 	booleanStruct{0x08, "readOnly", false},
	// 	doubleStruct{0x01, "ok", 61503}}

	// switch s.versionBoth {
	// case "34":
	// 	return s.encodeResponse(md, elements34)
	// case "36":
	// 	return s.encodeResponse(md, elements36)
	// }

	return s.encodeResponse(md, elements34)
}

func (s *mongodbService) uriMsg(md *MsgData) []byte {
	elements34 := []elem{
		stringStruct{0x02, "you", 0, s.clAddr.String()},
		doubleStruct{0x01, "ok", 61503}}

	// elements36 := []elem{
	// 	stringStruct{0x02, "you", 0, s.clAddr.String()}}

	// switch s.versionBoth {
	// case "34":
	// 	return s.encodeResponse(md, elements34)
	// case "36":
	// 	return s.encodeResponse(md, elements36)
	// }

	return s.encodeResponse(md, elements34)
}

func (s *mongodbService) buildInfoMsg(md *MsgData) []byte {

	vs := []int32{}
	for _, s := range strings.Split(s.Version, ".") {
		i, _ := strconv.Atoi(s)
		vs = append(vs, int32(i))
	}
	if len(vs) == 3 {
		vs = append(vs, int32(0))
	}

	versionDoc := document{
		elements: []elem{
			int32Struct{0x10, "0", vs[0]},
			int32Struct{0x10, "1", vs[1]},
			int32Struct{0x10, "2", vs[2]},
			int32Struct{0x10, "3", vs[3]}}}

	opensslDoc := document{
		elements: []elem{
			stringStruct{0x02, "running", 0, "OpenSSL 1.0.2o  27 Mar 2018"},
			stringStruct{0x02, "compiled", 0, "OpenSSL 1.0.2n  7 Dec 2017"}}}

	buildenvironmentDoc := document{
		elements: []elem{
			stringStruct{0x02, "distmod", 0, ""},
			stringStruct{0x02, "distarch", 0, "x86_64"},
			stringStruct{0x02, "cc", 0, "/usr/bin/clang: Apple LLVM version 9.0.0 (clang-900.0.39.2)"},
			stringStruct{0x02, "ccflags", 0, "-I/usr/local/opt/openssl/include -fno-omit-frame-pointer -fno-strict-aliasing -ggdb -pthread -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -O2 -Wno-unused-local-typedefs -Wno-unused-function -Wno-unused-private-field -Wno-deprecated-declarations -Wno-tautological-constant-out-of-range-compare -Wno-unused-const-variable -Wno-missing-braces -Wno-inconsistent-missing-override -Wno-potentially-evaluated-expression -fstack-protector-strong -Wno-null-conversion -mmacosx-version-min=10.13 -fno-builtin-memcmp"},
			stringStruct{0x02, "cxx", 0, "/usr/bin/clang++: Apple LLVM version 9.0.0 (clang-900.0.39.2)"},
			stringStruct{0x02, "cxxflags", 0, "-Woverloaded-virtual -Wpessimizing-move -Wredundant-move -Wno-undefined-var-template -std=c++11"},
			stringStruct{0x02, "linkflags", 0, "-L/usr/local/opt/openssl/lib -pthread -Wl,-bind_at_load -fstack-protector-strong -mmacosx-version-min=10.13"},
			stringStruct{0x02, "target_arch", 0, "x86_64"},
			stringStruct{0x02, "target_os", 0, "osx"}}}

	storageEnginesDoc := document{
		elements: []elem{
			stringStruct{0x02, "0", 0, "devnull"},
			stringStruct{0x02, "1", 0, "ephemeralForTest"},
			stringStruct{0x02, "2", 0, "mmapv1"},
			stringStruct{0x02, "3", 0, "wiredTiger"}}}

	elements := []elem{
		stringStruct{0x02, "version", 0, s.Version},
		stringStruct{0x02, "gitVersion", 0, "9586e557d54ef70f9ca4b43c26892cd55257e1a5"},
		arrayStruct{0x04, "modules", document{}},
		stringStruct{0x02, "allocator", 0, "system"},
		stringStruct{0x02, "javascriptEngine", 0, "mozjs"},
		stringStruct{0x02, "sysInfo", 0, "deprecated"},
		arrayStruct{0x04, "versionArray", versionDoc},
		documentStruct{0x03, "openssl", opensslDoc},
		documentStruct{0x03, "buildEnvironment", buildenvironmentDoc},
		int32Struct{0x10, "bits", 64},
		booleanStruct{0x08, "debug", false},
		int32Struct{0x10, "maxBsonObjectSize", 16777216},
		arrayStruct{0x04, "storageEngines", storageEnginesDoc},
		doubleStruct{0x01, "ok", 61503}}

	return s.encodeResponse(md, elements)
}

func (s *mongodbService) getLogMsg(md *MsgData) []byte {

	if s.authActivated { // TODO && s.logged
		return s.errorMsgNotAuthorized(md, "getLog: \"startupWarnings\"")
	}

	logDoc := document{
		elements: []elem{
			stringStruct{0x02, "0", 0, startupTime() + "I CONTROL  [initandlisten]"},
			stringStruct{0x02, "1", 0, startupTime() + "I CONTROL  [initandlisten] ** WARNING: Access control is not enabled for the database."},
			stringStruct{0x02, "2", 0, startupTime() + "I CONTROL  [initandlisten] **          Read and write access to data and configuration is unrestricted."},
			stringStruct{0x02, "3", 0, startupTime() + "I CONTROL  [initandlisten]"}}}

	elements := []elem{
		int32Struct{0x10, "totalLinesWritten", 4},
		arrayStruct{0x04, "log", logDoc},
		doubleStruct{0x01, "ok", 61503}}

	return s.encodeResponse(md, elements)
}

func (s *mongodbService) replSetMsg(md *MsgData) []byte {
	if s.authActivated { // TODO && !s.logged
		return s.errorMsgNotAuthorized(md, "replSetGetStatus: 1.0, forShell: 1.0")
	}
	elements := []elem{
		doubleStruct{0x01, "ok", 0},
		stringStruct{0x02, "errmsg", 0, "not running with --replSet"},
		int32Struct{0x10, "code", 76},
		stringStruct{0x02, "codeName", 0, "NoReplicationEnabled"}}

	return s.encodeResponse(md, elements)
}

func dbSize(s int64) []byte {
	h := math.Float64bits(float64(s))
	local := make([]byte, 8)
	binary.LittleEndian.PutUint64(local, h)
	return local
}

func encodeDb(Database Db, totalSize *int64) document {
	name := Database.Name
	size, _ := strconv.ParseInt(Database.SizeOnDisk, 10, 64)
	empty, _ := strconv.ParseBool(Database.Empty)
	sz := dbSize(size)
	doc := document{
		elements: []elem{
			stringStruct{0x02, "name", 0, name},
			dbSizeStruct{0x01, "sizeOnDisk", sz},
			booleanStruct{0x08, "empty", empty}}}

	*totalSize += size
	return doc
}

func (s *mongodbService) listDatabasesMsg(md *MsgData) []byte {

	if s.authActivated {
		if !s.logged {
			return s.errorMsgNotAuthorized(md, "listDatabases:  1.0")
		}
	}

	var totalSize int64
	databasesDoc := document{}
	for i, database := range s.Dbs {
		db := documentStruct{
			elemType: 0x03,
			name:     strconv.FormatInt(int64(i), 10),
			query:    encodeDb(database, &totalSize),
		}
		databasesDoc.elements = append(databasesDoc.elements, db)
	}
	sz := dbSize(totalSize)
	elements := []elem{
		arrayStruct{0x04, "databases", databasesDoc},
		dbSizeStruct{0x01, "totalSize", sz},
		doubleStruct{0x01, "ok", 61503},
	}
	return s.encodeResponse(md, elements)
}

func createNonce() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		log.Error("Error creating Nonce: %s", err.Error())
		return ""
	}
	return hex.EncodeToString(bytes)
}

func (s *mongodbService) getnonceMsg(md *MsgData) []byte {
	elements := []elem{
		stringStruct{0x02, "nonce", 0, createNonce()},
		doubleStruct{0x01, "ok", 61503}}
	return s.encodeResponse(md, elements)
}

func (s *mongodbService) saslstartMsg(md *MsgData) []byte {
	s.Client = Client{}
	payload := md.rq.(*OpCommandMsg).metadata.elements[2].(binaryStruct).payload //TODO improve elemnts[i]
	s.parsePayload(md, payload)
	if !s.userExists(md) {
		return s.authFailMsg(md)
	}
	return s.authHandShake(md)
}

func (s *mongodbService) pingMsg(md *MsgData) []byte {
	elements := []elem{
		doubleStruct{0x01, "ok", 61503}}
	return s.encodeResponse(md, elements)
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
			}
		case bytes.HasPrefix(slice, []byte("p=")):
			s.Client.clProof = string(slice[2:])
		}
	}
}

func (s *mongodbService) userExists(md *MsgData) bool {
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
		return false // or break ?
	}
	return false
}

func (s *mongodbService) authHandShake(md *MsgData) []byte {
	itercount := strconv.Itoa(s.itercounts)
	s.createServerNonce()
	payload := "r=" + s.Client.cbNonce + "," + "s=" + s.Client.salt + "," + "i=" + itercount
	elements := []elem{
		int32Struct{0x10, "conversationId", 1},
		booleanStruct{0x08, "done", false},
		binaryStruct{0x05, "payload", 0, 0x00, []byte(payload)},
		doubleStruct{0x01, "ok", 61503}}
	return s.encodeResponse(md, elements)
}

func (s *mongodbService) createServerNonce() {
	s.Client.svNonce = "/A+KIDOqoL+95VaItbNN9geuyPE6Hv52" //TODO generated randomly + ? saved in ev ?
	s.Client.cbNonce = s.Client.clNonce + s.Client.svNonce
}

func (s *mongodbService) conversationIdMsg(md *MsgData) []byte {

	if s.Client == (Client{}) {
		elements := []elem{
			doubleStruct{0x01, "ok", 0},
			stringStruct{0x02, "errmsg", 0, "No SASL session state found"},
			int32Struct{0x10, "code", 17},
			stringStruct{0x02, "codeName", 0, "ProtocolError"}}
		return s.encodeResponse(md, elements) //TODO errorMsg()
	}

	payload := md.rq.(*OpCommandMsg).metadata.elements[1].(binaryStruct).payload //TODO improve elemnts[i]
	l := md.rq.(*OpCommandMsg).metadata.elements[1].(binaryStruct).length        //TODO improve elemnts[i]

	if l != 0 { //1st salscontinue
		s.parsePayload(md, payload)
		return s.serverSignature(md)
	}
	// 2nd sasl continue with empty payload
	s.parsePayload(md, payload)

	// if !s.logged {
	// 	// when we directly send saslcontinue
	// }

	return s.finishAuthShake(md)
}

func (s *mongodbService) serverSignature(md *MsgData) []byte {
	serverSign, proved := s.scram()
	if !proved {
		log.Error("Check proof wrong!")
		return s.authFailMsg(md)
	}

	s.logged = true

	payload := "v=" + string(serverSign)

	elements := []elem{
		int32Struct{0x10, "conversationId", 1},
		booleanStruct{0x08, "done", false},
		binaryStruct{0x05, "payload", 0, 0x00, []byte(payload)},
		doubleStruct{0x01, "ok", 61503}}

	return s.encodeResponse(md, elements)
}

func (s *mongodbService) finishAuthShake(md *MsgData) []byte {
	elements := []elem{
		int32Struct{0x10, "conversationId", 1},
		booleanStruct{0x08, "done", true},
		binaryStruct{0x05, "payload", 0, 0x00, []byte("")},
		doubleStruct{0x01, "ok", 61503}}
	return s.encodeResponse(md, elements)
}

/////// Error Msg //////////// TODO regroup in one errorMsg(md, errmsg, code, codeName)
func (s *mongodbService) errorMsgNotAuthorized(md *MsgData, msg string) []byte {
	elements := []elem{
		doubleStruct{0x01, "ok", 0},
		stringStruct{0x02, "errmsg", 0, fmt.Sprintf("not authorized on %s to execute command { %s }", md.rq.(*OpCommandMsg).database, msg)},
		int32Struct{0x10, "code", 13},
		stringStruct{0x02, "codeName", 0, "Unauthorized"}}
	return s.encodeResponse(md, elements)
}

func (s *mongodbService) authFailMsg(md *MsgData) []byte {
	elements := []elem{
		doubleStruct{0x01, "ok", 0},
		stringStruct{0x02, "errmsg", 0, "Authentication failed."},
		int32Struct{0x10, "code", 18},
		stringStruct{0x02, "codeName", 0, "AuthenticationFailed"}}
	return s.encodeResponse(md, elements)
}

func (s *mongodbService) errorMsgBadCmd(md *MsgData) []byte {
	elements := []elem{
		doubleStruct{0x01, "ok", 0},
		stringStruct{0x02, "errmsg", 0, fmt.Sprintf("no such command: '%s', bad cmd: '{ %s: 1.0}'", md.cmd, md.cmd)},
		int32Struct{0x10, "code", 59},
		stringStruct{0x02, "codeName", 0, "CommandNotFound"}}
	return s.encodeResponse(md, elements)
}

/////////////////
func (s *mongodbService) errorMsg(md *MsgData, errmsg, codeName string, code int32) []byte {
	elements := []elem{
		doubleStruct{0x01, "ok", 0},
		stringStruct{0x02, "errmsg", 0, errmsg},
		int32Struct{0x10, "code", code},
		stringStruct{0x02, "codeName", 0, codeName}}
	return s.encodeResponse(md, elements)
}
