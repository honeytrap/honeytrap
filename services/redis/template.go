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

type RedisServiceConfiguration struct {
	ServerVersion
	ClientsVersion
	MemoryVersion
	PersistenceVersion
	StatsVersion
	ReplicationVersion
	CpuVersion
	CommandStatsVersion
	ClusterVersion
	KeyspaceVersion
	ConnectedClientsVersion
}

type ServerVersion struct {
	Redis_version     string
	Redis_git_sha1    string
	Redis_git_dirty   string
	Redis_build_id    string
	Redis_mode        string
	Os                string
	Arch_bits         string
	Multiplexing_api  string
	Atomicvar_api     string
	Gcc_version       string
	Process_id        string
	Run_id            string
	Tcp_port          string
	Uptime_in_seconds string
	Uptime_in_days    string
	Hz                string
	Lru_clock         string
	Executable        string
	Config_file       string
}

type ClientsVersion struct {
	Connected_clients          string
	Client_longest_output_list string
	Client_biggest_input_buf   string
	Blocked_clients            string
}

type MemoryVersion struct {
	Used_memory               string
	Used_memory_human         string
	Used_memory_rss           string
	Used_memory_rss_human     string
	Used_memory_peak          string
	Used_memory_peak_human    string
	Used_memory_peak_perc     string
	Used_memory_overhead      string
	Used_memory_startup       string
	Used_memory_dataset       string
	Used_memory_dataset_perc  string
	Total_system_memory       string
	Total_system_memory_human string
	Used_memory_lua           string
	Used_memory_lua_human     string
	Maxmemory                 string
	Maxmemory_human           string
	Maxmemory_policy          string
	Mem_fragmentation_ratio   string
	Mem_allocator             string
	Active_defrag_running     string
	Lazyfree_pending_objects  string
}

type PersistenceVersion struct {
	Loading                      string
	Rdb_changes_since_last_save  string
	Rdb_bgsave_in_progress       string
	Rdb_last_save_time           string
	Rdb_last_bgsave_status       string
	Rdb_last_bgsave_time_sec     string
	Rdb_current_bgsave_time_sec  string
	Rdb_last_cow_size            string
	Aof_enabled                  string
	Aof_rewrite_in_progress      string
	Aof_rewrite_scheduled        string
	Aof_last_rewrite_time_sec    string
	Aof_current_rewrite_time_sec string
	Aof_last_bgrewrite_status    string
	Aof_last_write_status        string
	Aof_last_cow_size            string
}

type StatsVersion struct {
	Total_connections_received string
	Total_commands_processed   string
	Instantaneous_ops_per_sec  string
	Total_net_input_bytes      string
	Total_net_output_bytes     string
	Instantaneous_input_kbps   string
	Instantaneous_output_kbps  string
	Rejected_connections       string
	Sync_full                  string
	Sync_partial_ok            string
	Sync_partial_err           string
	Expired_keys               string
	Evicted_keys               string
	Keyspace_hits              string
	Keyspace_misses            string
	Pubsub_channels            string
	Pubsub_patterns            string
	Latest_fork_usec           string
	Migrate_cached_sockets     string
	Slave_expires_tracked_keys string
	Active_defrag_hits         string
	Active_defrag_misses       string
	Active_defrag_key_hits     string
	Active_defrag_key_misses   string
}

type ReplicationVersion struct {
	Role                           string
	Connected_slaves               string
	Master_replid                  string
	Master_replid2                 string
	Master_repl_offset             string
	Second_repl_offset             string
	Repl_backlog_active            string
	Repl_backlog_size              string
	Repl_backlog_first_byte_offset string
	Repl_backlog_histlen           string
}

type CpuVersion struct {
	Used_cpu_sys           string
	Used_cpu_user          string
	Used_cpu_sys_children  string
	Used_cpu_user_children string
}

type CommandStatsVersion struct {
	Cmdstat_info string
}

type ClusterVersion struct {
	Cluster_enabled string
}

type KeyspaceVersion struct {
}

type ConnectedClientsVersion struct {
}

var RedisDefault = RedisServiceConfiguration{
	ServerVersion: ServerVersion{
		Redis_version:     "1.0.0;U;4.0.8",
		Redis_git_sha1:    "1.0.0;TBD;00000000",
		Redis_git_dirty:   "1.0.0;TBD;0",
		Redis_build_id:    "1.0.0;TBD;f1060845dd32471a",
		Redis_mode:        "1.0.0;TBD;standalone",
		Os:                "1.0.0;TBD;Linux 4.9.49-moby x86_64",
		Arch_bits:         "1.0.0;TBD;64",
		Multiplexing_api:  "1.0.0;TBD;epoll",
		Atomicvar_api:     "1.0.0;TBD;atomic-builtin",
		Gcc_version:       "1.0.0;TBD;4.9.2",
		Process_id:        "1.0.0;TBD;1",
		Run_id:            "1.0.0;TBD;15444ca686daa4cfcf621e65a7aed097110bb598",
		Tcp_port:          "1.0.0;TBD;6379",
		Uptime_in_seconds: "1.0.0;TBD;10396",
		Uptime_in_days:    "1.0.0;TBD;0",
		Hz:                "1.0.0;TBD;10",
		Lru_clock:         "1.0.0;TBD;8584119",
		Executable:        "1.0.0;TBD;/data/redis-server",
		Config_file:       "1.0.0;TBD;",
	},
	ClientsVersion: ClientsVersion{
		Connected_clients:          "3.0.0;TBD;0",
		Client_longest_output_list: "1.0.0;TBD;0",
		Client_biggest_input_buf:   "1.0.0;TBD;0",
		Blocked_clients:            "2.4.0;TBD;0",
	},
	MemoryVersion: MemoryVersion{
		Used_memory:               "1.0.0;TBD;1828264",
		Used_memory_human:         "1.0.0;TBD;808.85K",
		Used_memory_rss:           "1.0.0;TBD;4120576",
		Used_memory_rss_human:     "3.2.8;TBD;3.93M",
		Used_memory_peak:          "1.0.0;TBD;828264",
		Used_memory_peak_human:    "1.0.0;TBD;808.85K",
		Used_memory_peak_perc:     "3.9.103;TBD;100.00%",
		Used_memory_overhead:      "3.9.103;TBD;815150",
		Used_memory_startup:       "3.9.103;TBD;765520",
		Used_memory_dataset:       "3.9.103;TBD;13114",
		Used_memory_dataset_perc:  "3.9.103;TBD;20.90%",
		Total_system_memory:       "3.2.8;TBD;2096160768",
		Total_system_memory_human: "3.2.8;TBD;1.95G",
		Used_memory_lua:           "1.0.0;TBD;37888",
		Used_memory_lua_human:     "3.2.8;TBD;37.00K",
		Maxmemory:                 "3.2.8;TBD;0",
		Maxmemory_human:           "3.2.8;TBD;0B",
		Maxmemory_policy:          "3.2.8;TBD;noeviction",
		Mem_fragmentation_ratio:   "1.0.0;TBD;4.97",
		Mem_allocator:             "1.0.0;TBD;jemalloc-4.0.3",
		Active_defrag_running:     "3.9.103;TBD;0",
		Lazyfree_pending_objects:  "3.9.103;TBD;0",
	},
	PersistenceVersion: PersistenceVersion{
		Loading:                      "1.0.0;TBD;0",
		Rdb_changes_since_last_save:  "1.0.0;TBD;0",
		Rdb_bgsave_in_progress:       "1.0.0;TBD;0",
		Rdb_last_save_time:           "1.0.0;TBD;1515759614",
		Rdb_last_bgsave_status:       "1.0.0;TBD;ok",
		Rdb_last_bgsave_time_sec:     "1.0.0;TBD;-1",
		Rdb_current_bgsave_time_sec:  "1.0.0;TBD;-1",
		Rdb_last_cow_size:            "1.0.0;TBD;0",
		Aof_enabled:                  "1.0.0;TBD;0",
		Aof_rewrite_in_progress:      "1.0.0;TBD;0",
		Aof_rewrite_scheduled:        "1.0.0;TBD;0",
		Aof_last_rewrite_time_sec:    "1.0.0;TBD;-1",
		Aof_current_rewrite_time_sec: "1.0.0;TBD;-1",
		Aof_last_bgrewrite_status:    "1.0.0;TBD;ok",
		Aof_last_write_status:        "1.0.0;TBD;ok",
		Aof_last_cow_size:            "1.0.0;TBD;0",
	},
	StatsVersion: StatsVersion{
		Total_connections_received: "1.0.0;TBD;2",
		Total_commands_processed:   "1.0.0;TBD;1",
		Instantaneous_ops_per_sec:  "1.0.0;TBD;0",
		Total_net_input_bytes:      "1.0.0;TBD;14",
		Total_net_output_bytes:     "1.0.0;TBD;2664",
		Instantaneous_input_kbps:   "1.0.0;TBD;0.00",
		Instantaneous_output_kbps:  "1.0.0;TBD;0.00",
		Rejected_connections:       "1.0.0;TBD;0",
		Sync_full:                  "1.0.0;TBD;0",
		Sync_partial_ok:            "1.0.0;TBD;0",
		Sync_partial_err:           "1.0.0;TBD;0",
		Expired_keys:               "1.0.0;TBD;0",
		Evicted_keys:               "1.0.0;TBD;0",
		Keyspace_hits:              "1.0.0;TBD;0",
		Keyspace_misses:            "1.0.0;TBD;0",
		Pubsub_channels:            "1.0.0;TBD;0",
		Pubsub_patterns:            "1.0.0;TBD;0",
		Latest_fork_usec:           "1.0.0;TBD;0",
		Migrate_cached_sockets:     "1.0.0;TBD;0",
		Slave_expires_tracked_keys: "1.0.0;TBD;0",
		Active_defrag_hits:         "1.0.0;TBD;0",
		Active_defrag_misses:       "1.0.0;TBD;0",
		Active_defrag_key_hits:     "1.0.0;TBD;0",
		Active_defrag_key_misses:   "1.0.0;TBD;0",
	},
	ReplicationVersion: ReplicationVersion{
		Role:                           "1.0.0;TBD;master",
		Connected_slaves:               "1.0.0;TBD;0",
		Master_replid:                  "1.0.0;TBD;29e814284ae0619c1b2c09175f4b5b6a5aafff48",
		Master_replid2:                 "1.0.0;TBD;0000000000000000000000000000000000000000",
		Master_repl_offset:             "1.0.0;TBD;0",
		Second_repl_offset:             "1.0.0;TBD;-1",
		Repl_backlog_active:            "1.0.0;TBD;0",
		Repl_backlog_size:              "1.0.0;TBD;1048576",
		Repl_backlog_first_byte_offset: "1.0.0;TBD;0",
		Repl_backlog_histlen:           "1.0.0;TBD;0",
	},
	CpuVersion: CpuVersion{
		Used_cpu_sys:           "1.0.0;TBD;20.83",
		Used_cpu_user:          "1.0.0;TBD;3.02",
		Used_cpu_sys_children:  "1.0.0;TBD;0.00",
		Used_cpu_user_children: "1.0.0;TBD;0.00",
	},
	CommandStatsVersion: CommandStatsVersion{
		Cmdstat_info: "1.0.0;TBD;calls=3,usec=181,usec_per_call=60.33",
	},
	ClusterVersion: ClusterVersion{
		Cluster_enabled: "1.0.0;TBD;0",
	},
	KeyspaceVersion:         KeyspaceVersion{},
	ConnectedClientsVersion: ConnectedClientsVersion{},
}
