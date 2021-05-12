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
package redis

type cmd func(*redisService, []interface{}) string

var mapCmds = map[string]cmd{
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
		return bulkString(s.infoSectionsMsg(), true)
	case 1:
		_word := args[0].(redisDatum)
		word, success := _word.ToString()
		if !success {
			return "Expected string argument, got something else"
		}
		fn, ok := mapInfoCmds[word]
		if ok {

			return bulkString(fn(s), true)
		}
		if word == "default" {
			return bulkString(s.infoSectionsMsg(), true)
		}
		if word == "all" {
			return bulkString(s.allSectionsMsg(), true)
		}
		return bulkString("", false)
	default:
		return errorMsg("syntax")
	}
}
