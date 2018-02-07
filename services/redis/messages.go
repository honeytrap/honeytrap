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
	"strings"
)

func (s *redisService) sectionsMsg() string {
	return s.infoServerMsg() + "\r\n" + s.infoClientsMsg() + "\r\n" + s.infoMemoryMsg() + "\r\n" + s.infoPersistenceMsg() + "\r\n" + s.infoStatsMsg() + "\r\n" + s.infoReplicationMsg() + "\r\n" + s.infoCPUMsg() + "\r\n" + s.infoClusterMsg()
}

func (s *redisService) infoSectionsMsg() string {
	return s.sectionsMsg() + "\r\n" + s.infoKeyspaceMsg()
}

func (s *redisService) allSectionsMsg() string {
	return s.sectionsMsg() + "\r\n" + s.infoCommandstatsMsg() + "\r\n" + s.infoKeyspaceMsg()
}

var msgs = map[string]string{}

func init() {

	msgs["ServerVersion"] = "# Server\r\n"
	msgs["ClientsVersion"] = "# Clients\r\n"
	msgs["MemoryVersion"] = "# Memory\r\n"
	msgs["PersistenceVersion"] = "# Persistence\r\n"
	msgs["StatsVersion"] = "# Stats\r\n"
	msgs["ReplicationVersion"] = "# Replication\r\n"
	msgs["CpuVersion"] = "# CPU\r\n"
	msgs["CommandStatsVersion"] = "# Commandstats\r\n"
	msgs["ClusterVersion"] = "# Cluster\r\n"
	msgs["KeyspaceVersion"] = "# Keyspace\r\n"
	msgs["ConnectedClientsVersion"] = "# Connected Clients\r\n"
}

func createMsg(section string) string {

	msg := msgs[section]
	RedisdefSection := Redisdef.FieldByName(section)

	for y := 0; y < RedisdefSection.NumField(); y++ {

		field := strings.ToLower(RedisdefSection.Type().Field(y).Name)
		value := RedisdefSection.Field(y).Interface().(string)

		if value != "__" {
			msg += fmt.Sprintf("%s:%s\r\n", field, value)
		}
	}
	return msg
}

func (s *redisService) infoServerMsg() string {
	return createMsg("ServerVersion")
}

func (s *redisService) infoClientsMsg() string {
	return createMsg("ClientsVersion")
}

func (s *redisService) infoMemoryMsg() string {
	return createMsg("MemoryVersion")
}

func (s *redisService) infoPersistenceMsg() string {
	return createMsg("PersistenceVersion")
}

func (s *redisService) infoStatsMsg() string {
	return createMsg("StatsVersion")
}

func (s *redisService) infoReplicationMsg() string {
	return createMsg("ReplicationVersion")
}

func (s *redisService) infoCPUMsg() string {
	return createMsg("CpuVersion")
}

func (s *redisService) infoCommandstatsMsg() string {
	return createMsg("CommandStatsVersion")
}

func (s *redisService) infoClusterMsg() string {
	return createMsg("ClusterVersion")
}

func (s *redisService) infoKeyspaceMsg() string {
	return createMsg("KeyspaceVersion")
}

func (s *redisService) infoConnectedClientsMsg() string {
	return createMsg("ConnectedClientsVersion")
}

func lenMsg() string {
	return "$%d\n%s\n"
}

func lineBreakMsg() string {
	return ""
}

func errorMsg(errType string) string {
	switch errType {
	case "syntax":
		return "-ERR syntax error\n"
	case "nbargs":
		return "-ERR wrong number of arguments for '%s' command\n"
	default:
		return "-ERR unknown command '%s'\n"
	}
}

func okMsg() string {
	return "+OK\n"
}
