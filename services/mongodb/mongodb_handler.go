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
	"strconv"
	"strings"
)

var mapCommands = map[string]func(*mongodbService, *MsgData) []byte{
	"whatsmyuri":       (*mongodbService).uriMsg,
	"buildInfo":        (*mongodbService).buildInfoMsg,
	"buildinfo":        (*mongodbService).buildInfoMsg,
	"getLog":           (*mongodbService).getLogMsg,
	"ismaster":         (*mongodbService).isMasterMsg,
	"replSetGetStatus": (*mongodbService).replSetMsg,
	"listDatabases":    (*mongodbService).listDatabasesMsg,
	"isMaster":         (*mongodbService).isMasterMsg,
	"getnonce":         (*mongodbService).getnonceMsg,
	"saslStart":        (*mongodbService).saslstartMsg,
	"saslContinue":     (*mongodbService).conversationIdMsg,
	"ping":             (*mongodbService).pingMsg,
	// "serverStatus": (*mongodbService).serverStatusMsg,
	//...
}

func (s *mongodbService) reqHandler(bb *bytes.Buffer) ([]byte, map[string]interface{}) {

	md := &MsgData{}

	decodeMsgHeader(bb, md)

	ev := make(eventLog)

	switch md.mh.OpCode {
	case 2004:
		md.rq = parseOpQuery(bb, ev)
		s.findSmallestVersion(md)
		cmd := md.rq.(*OpQueryMsg).query.elements[0]
		switch cmd.(type) {
		case doubleStruct:
			md.cmd = cmd.(doubleStruct).name
		case int32Struct:
			md.cmd = cmd.(int32Struct).name
		case booleanStruct: // nodejs
			md.cmd = cmd.(booleanStruct).name
		}

	case 2010:
		md.rq = parseOpCommand(bb, ev)
		md.cmd = md.rq.(*OpCommandMsg).commandName
		// if md.cmd = "saslStart" {
		// 	parseUsername()
		// }

	default:
		log.Error("OpCode not implemented: %d", md.mh.OpCode)
		ev["mongodb.opCodeNotImplemented"] = md.mh.OpCode
		// TODO
	}

	return s.respHandler(md, ev), ev
}

func (s *mongodbService) findSmallestVersion(md *MsgData) {
	//TODO: Fieldbyname
	svV := strings.Split(s.Version, ".")
	if len(md.rq.(*OpQueryMsg).query.elements) == 1 {
		s.versionBoth = strings.Join(svV[:2], "")
	} else {

		clientV := md.rq.(*OpQueryMsg).query.elements[1].(documentStruct).query.elements[1].(documentStruct).query.elements[1].(stringStruct).value
		clV := strings.Split(clientV, ".")

		bothV := svV // incase it's ==

		for i := range svV {
			x, _ := strconv.Atoi(svV[i])
			y, _ := strconv.Atoi(clV[i])

			if x == y {
				continue
			}
			if x < y {
				bothV = svV
				break
			}
			bothV = clV
			break
		}

		s.versionBoth = strings.Join(bothV[:2], "")
	}
}

func (s *mongodbService) respHandler(md *MsgData) []byte {
	// fn, ok := mapCommands[strings.ToLower(md.cmd)]
	fn, ok := mapCommands[md.cmd]
	if !ok {
		log.Error("Error: command not implemented: %s", md.cmd)
		return s.errorMsgBadCmd(md)
	}
	return fn(s, md)
}
