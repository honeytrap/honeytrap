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
	return s.infoServerMsg() + s.infoClientsMsg() + s.infoMemoryMsg() + s.infoPersistenceMsg() + s.infoStatsMsg() + s.infoReplicationMsg() + s.infoCPUMsg() + s.infoClusterMsg()
}

func (s *redisService) infoSectionsMsg() string {
	return s.sectionsMsg() + s.infoKeyspaceMsg()
}

func (s *redisService) allSectionsMsg() string {
	return s.sectionsMsg() + s.infoCommandstatsMsg() + s.infoKeyspaceMsg()
}

func (s *redisService) infoServerMsg() string {
	return fmt.Sprintf(`# Server
redis_version:%s
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:f1060845dd32471a
redis_mode:standalone
os:%s
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.9.2
process_id:1
run_id:15444ca686daa4cfcf621e65a7aed097110bb598
tcp_port:6379
uptime_in_seconds:10396
uptime_in_days:0
hz:10
lru_clock:5820570
executable:/data/redis-server
config_file:

`, s.Version, s.Os)
}

func (s *redisService) infoClientsMsg() string {
	return `# Clients
connected_clients:1
client_longest_output_list:0
client_biggest_input_buf:0
blocked_clients:0

`
}

func (s *redisService) infoMemoryMsg() string {
	return `# Memory
used_memory:1828264
used_memory_human:808.85K
used_memory_rss:4120576
used_memory_rss_human:3.93M
used_memory_peak:828264
used_memory_peak_human:808.85K
used_memory_peak_perc:100.00%
used_memory_overhead:815150
used_memory_startup:765520
used_memory_dataset:13114
used_memory_dataset_perc:20.90%
total_system_memory:2096160768
total_system_memory_human:1.95G
used_memory_lua:37888
used_memory_lua_human:37.00K
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
mem_fragmentation_ratio:4.97
mem_allocator:jemalloc-4.0.3
active_defrag_running:0
lazyfree_pending_objects:0

`
}

func (s *redisService) infoPersistenceMsg() string {
	return `# Persistence
loading:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1515759614
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:-1
rdb_current_bgsave_time_sec:-1
rdb_last_cow_size:0
aof_enabled:0
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok
aof_last_cow_size:0

`
}

func (s *redisService) infoStatsMsg() string {
	return `total_connections_received:2
total_commands_processed:1
instantaneous_ops_per_sec:0
total_net_input_bytes:14
total_net_output_bytes:2664
instantaneous_input_kbps:0.00
instantaneous_output_kbps:0.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
evicted_keys:0
keyspace_hits:0
keyspace_misses:0
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:0
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0

`
}

func (s *redisService) infoReplicationMsg() string {
	return `# Replication
role:master
connected_slaves:0
master_replid:29e814284ae0619c1b2c09175f4b5b6a5aafff48
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

`
}

func (s *redisService) infoCPUMsg() string {
	return `# CPU
used_cpu_sys:20.83
used_cpu_user:3.02
used_cpu_sys_children:0.00
used_cpu_user_children:0.00

`
}

func (s *redisService) infoCommandstatsMsg() string {
	return `# Commandstats
cmdstat_info:calls=3,usec=181,usec_per_call=60.33

`
}

func (s *redisService) infoClusterMsg() string {
	return `# Cluster
cluster_enabled:0

`
}

func (s *redisService) infoKeyspaceMsg() string {
	return `# Keyspace

`
}

func errorMsg(errType string) string {
	switch errType {
	case "syntax":
		return "-ERR syntax error\r\n"
	default:
		return "-ERR unknown command '%s'\r\n"
	}
}

func bulkString(text string, convertToCRLF bool) string {
	if convertToCRLF {
		text = strings.Replace(text, "\n", "\r\n", -1)
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(text), text)
}