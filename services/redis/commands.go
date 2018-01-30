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

type cmd func(*redisService, []string, string) (string, bool)

var mapCmds = map[string]cmd{
	"info":     (*redisService).infoCmd,
	"flushall": (*redisService).flushallCmd,
	"set":      (*redisService).setCmd,
	// ...
}

type infoSection func(*redisService) string

var mapInfoCmds = map[string]infoSection{
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

func (s *redisService) infoCmd(args []string, userCmd string) (string, bool) {
	switch len(args) {
	case 1:
		return fmt.Sprintf(lenMsg(), len(s.infoSectionsMsg()), s.infoSectionsMsg()), false
	case 2:
		if fn, ok := mapInfoCmds[args[1]]; ok {
			return fmt.Sprintf(lenMsg(), len(fn(s)), fn(s)), false
		} else if args[1] == "default" {
			return fmt.Sprintf(lenMsg(), len(s.infoSectionsMsg()), s.infoSectionsMsg()), false
		} else if args[1] == "all" {
			return fmt.Sprintf(lenMsg(), len(s.allSectionsMsg()), s.allSectionsMsg()), false
		} else {
			return fmt.Sprintf(lenMsg(), len(lineBreakMsg()), lineBreakMsg()), false
		}
	default:
		return errorMsg("syntax"), false
	}
}
func (s *redisService) flushallCmd(args []string, userCmd string) (string, bool) {
	switch len(args) {
	case 1:
		return okMsg(), false
	case 2:
		if args[1] == "async" {
			return okMsg(), false
		} else {
			return errorMsg("syntax"), false
		}
	default:
		return errorMsg("syntax"), false
	}
}

func (s *redisService) setCmd(args []string, userCmd string) (string, bool) {

	return okMsg(), false

	/* switch len(args) {
	case 1, 2:
		return fmt.Sprintf(errorMsg("nbargs"), args[0]), false
	case 3:
		return okMsg(), false
	default:
		return errorMsg("syntax"), false
	}*/
}
