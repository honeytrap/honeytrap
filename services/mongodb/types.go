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

type eventLog map[string]interface{}

type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        int32
}

type MsgData struct {
	mh  MsgHeader
	rq  interface{}
	cmd string
	// data []interface{}
}

type dbSizeStruct struct {
	elemType byte
	name     string
	size     []byte
}

type fullCollectionName struct {
	databaseName   string
	collectionName string
}

type document struct {
	length   int32
	elements []elem
}

type elements []elem

type elem interface {
}

/////// Messages Types ////////
// opCode: 1
type OpReplyMsg struct {
	MsgHeader
	responseFlags  int32
	cursorID       int64
	startingFrom   int32
	numberReturned int32
	documents      []document
}

// opCode: 2004
type OpQueryMsg struct {
	flags int32
	fullCollectionName
	numberToSkip         int32
	numberToReturn       int32
	query                document
	returnFieldsSelector document //optional
}

// opCode: 2010
type OpCommandMsg struct { //
	MsgHeader
	database    string
	commandName string
	metadata    document
	commandArgs document
	inputDocs   []document // zero or more documents
}

// opCode: 2011
type OpCommandReplyMsg struct {
	MsgHeader
	metadata     document
	commandReply document
	outputDocs   document // not currently in use
}

// opCode: 2013
type OpMsgMsg struct {
	MsgHeader
	flagBits uint32
	sections []elem
	checksum uint32 //optional
}

/////// Bson Types ////////
//byte: 0x01
type doubleStruct struct {
	elemType byte
	name     string
	value    int64
}

// byte: 0x02
type stringStruct struct {
	elemType byte
	name     string
	length   int32
	value    string
}

// byte: 0x03
type documentStruct struct {
	elemType byte
	name     string
	query    document
}

// byte: 0x04
type arrayStruct struct {
	elemType byte
	name     string
	doc      document
}

// byte: 0x05
type binaryStruct struct {
	elemType byte
	name     string
	length   int32
	subtype  byte
	payload  []byte
}

// byte: 0x08
type booleanStruct struct {
	elemType byte
	name     string
	value    bool
}

// byte: 0x09
type datetimeStruct struct {
	elemType byte
	name     string
	value    int64
}

// byte: 0x10
type int32Struct struct {
	elemType byte
	name     string
	value    int32
}
