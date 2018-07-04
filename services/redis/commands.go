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
package redis

import (
	"fmt"
)

var mapCmds = map[string]func(*redisService, []interface{}) string{
	"info": (*redisService).infoCmd,
	// ...
}

var mapInfoCmds = map[string]func(*redisService) string{
	"server":      (*redisService).infoServerMsg,
	"clients":     (*redisService).infoClientsMsg,
	"memory":      (*redisService).infoMemoryMsg,
	"persistence": (*redisService).infoPersistenceMsg,
	"stats":       (*redisService).infoStatsMsg,
	"replication": (*redisService).infoReplicationMsg,
	"cpu":         (*redisService).infoCPUMsg,
	"cluster":     (*redisService).infoClusterMsg,
	"keyspace":    (*redisService).infoKeyspaceMsg,
}

func (s *redisService) infoCmd(args []interface{}) string {
	switch len(args) {
	case 0:
		return fmt.Sprintf(lenMsg(), len(s.infoSectionsMsg()), s.infoSectionsMsg())
	case 1:
		_word := args[0].(redisDatum)
		word, success := _word.ToString()
		if !success {
			return "Expected string argument, got something else"
		}
		fn, ok := mapInfoCmds[word]
		if ok {
			return fmt.Sprintf(lenMsg(), len(fn(s)), fn(s))
		}
		if word == "default" {
			return fmt.Sprintf(lenMsg(), len(s.infoSectionsMsg()), s.infoSectionsMsg())
		}
		if word == "all" {
			return fmt.Sprintf(lenMsg(), len(s.allSectionsMsg()), s.allSectionsMsg())
		}
		return fmt.Sprintf(lenMsg(), len(lineBreakMsg()), lineBreakMsg())
	default:
		return errorMsg("syntax")
	}
}
