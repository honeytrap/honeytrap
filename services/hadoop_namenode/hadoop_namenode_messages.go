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

package hadoop_namenode

import "fmt"

func (s *hadoopService) showEmpty() string {
	return fmt.Sprintln(`
		{}
		`)
}

func (s *hadoopService) showNothing() string {
	return fmt.Sprintln(`
		{
  "beans" : [ ]
}
		`)
}

func (s *hadoopService) showFSNamesystemState() string {
	// the querry "/jmx?qry=Hadoop:service=NameNode,name=FSNamesystemState" works only on port 50070
	return fmt.Sprintln(`
		{
		  "beans" : [ {
		    "name" : "Hadoop:service=NameNode,name=FSNamesystemState",
		    "modelerType" : "org.apache.hadoop.hdfs.server.namenode.FSNamesystem",
		    "CapacityTotal" : 41083600896,
		    "CapacityUsed" : 327680,
		    "CapacityRemaining" : 14133362688,
		    "TotalLoad" : 1,
		    "SnapshotStats" : "{\"SnapshottableDirectories\":0,\"Snapshots\":0}",
		    "BlocksTotal" : 31,
		    "MaxObjects" : 0,
		    "FilesTotal" : 35,
		    "PendingReplicationBlocks" : 0,
		    "UnderReplicatedBlocks" : 0,
		    "ScheduledReplicationBlocks" : 0,
		    "PendingDeletionBlocks" : 0,
		    "BlockDeletionStartTime" : 1530696472555,
		    "FSState" : "Operational",
		    "NumLiveDataNodes" : 1,
		    "NumDeadDataNodes" : 0,
		    "NumDecomLiveDataNodes" : 0,
		    "NumDecomDeadDataNodes" : 0,
		    "VolumeFailuresTotal" : 0,
		    "EstimatedCapacityLostTotal" : 0,
		    "NumDecommissioningDataNodes" : 0,
		    "NumStaleDataNodes" : 0,
		    "NumStaleStorages" : 0,
		    "TopUserOpCounts" : "{\"timestamp\":\"2018-07-04T05:48:36-0400\",\"windows\":[{\"windowLenMs\":300000,\"ops\":[]},{\"windowLenMs\":1500000,\"ops\":[]},{\"windowLenMs\":60000,\"ops\":[]}]}"
		  } ]
		}
		`)
}

func (s *hadoopService) showNamenode() string {
	return fmt.Sprintln(`
		"beans" : [ {
		"name" : "Hadoop:service=NameNode,name=NameNodeInfo",
		"modelerType" : "org.apache.hadoop.hdfs.server.namenode.FSNamesystem",
		"UpgradeFinalized" : true,
		"ClusterId" : "CID-5e691286-4de5-4dde-800b-c02a7a8bf44a",
		"Version" : `, s.Version, `,
		"Used" : 327680,
		"Free" : 14516498432,
		"Safemode" : "",
		"NonDfsUsedSpace" : 26566774784,
		"PercentUsed" : 7.975932E-4,
		"BlockPoolUsedSpace" : 327680,
		"PercentBlockPoolUsed" : 7.975932E-4,
		"PercentRemaining" : 35.334045,
		"CacheCapacity" : 0,
		"CacheUsed" : 0,
		"TotalBlocks" : 31,
		"TotalFiles" : 35,
		"NumberOfMissingBlocks" : 0,
		"NumberOfMissingBlocksWithReplicationFactorOne" : 0,
		"LiveNodes" : "{\"fdaebb1fae81:50010\":{\"infoAddr\":\"172.17.0.2:50075\",\"infoSecureAddr\":\"172.17.0.2:0\",\"xferaddr\":\"172.17.0.2:50010\",\"lastContact\":0,\"usedSpace\":327680,\"adminState\":\"In Service\",\"nonDfsUsedSpace\":26566774784,\"capacity\":41083600896,\"numBlocks\":31,\"version\":\"`, s.Version, `\",\"used\":327680,\"remaining\":14516498432,\"blockScheduled\":0,\"blockPoolUsed\":327680,\"blockPoolUsedPercent\":7.975932E-4,\"volfails\":0}}",
		"DeadNodes" : "{}",
		"DecomNodes" : "{}",
		"BlockPoolId" : "BP-1961412683-172.17.0.32-1450036414523",
		"NameDirStatuses" : "{\"failed\":{},\"active\":{\"/tmp/hadoop-root/dfs/name\":\"IMAGE_AND_EDITS\"}}",
		"NodeUsage" : "{\"nodeUsage\":{\"min\":\"0.00%\",\"median\":\"0.00%\",\"max\":\"0.00%\",\"stdDev\":\"0.00%\"}}",
		"NameJournalStatus" : "[{\"stream\":\"EditLogFileOutputStream(/tmp/hadoop-root/dfs/name/current/edits_inprogress_0000000000000000192)\",\"manager\":\"FileJournalManager(root=/tmp/hadoop-root/dfs/name)\",\"required\":\"false\",\"disabled\":\"false\"}]",
		"JournalTransactionInfo" : "{\"LastAppliedOrWrittenTxId\":\"192\",\"MostRecentCheckpointTxId\":\"191\"}",
		"NNStarted" : "Mon Jul 02 05:45:49 EDT 2018",
		"CompileInfo" : "2015-06-29T06:04Z by jenkins from (detached from 15ecc87)",
		"CorruptFiles" : "[]",
		"DistinctVersionCount" : 1,
		"DistinctVersions" : [ {
			"key" : "`, s.Version, `",
			"value" : 1
		} ],
		"SoftwareVersion" : `, s.Version, `,
		"RollingUpgradeStatus" : null,
		"Threads" : 34,
		"Total" : 41083600896
	} ]
		`)
}

func (s *hadoopService) showWithoutQuerry() string {
	return fmt.Sprintln(`
		{
  "beans" : [ {
    "name" : "java.lang:type=Memory",
    "modelerType" : "sun.management.MemoryImpl",
    "Verbose" : false,
    "HeapMemoryUsage" : {
      "committed" : 114294784,
      "init" : 64449664,
      "max" : 932184064,
      "used" : 24682376
    },
    "NonHeapMemoryUsage" : {
      "committed" : 31391744,
      "init" : 24576000,
      "max" : 136314880,
      "used" : 29944432
    },
    "ObjectPendingFinalizationCount" : 0,
    "ObjectName" : "java.lang:type=Memory"
  }, {
    "name" : "java.lang:type=MemoryPool,name=PS Eden Space",
    "modelerType" : "sun.management.MemoryPoolImpl",
    "CollectionUsage" : {
      "committed" : 62914560,
      "init" : 16252928,
      "max" : 330825728,
      "used" : 0
    },
    "CollectionUsageThreshold" : 0,
    "CollectionUsageThresholdCount" : 0,
    "MemoryManagerNames" : [ "PS MarkSweep", "PS Scavenge" ],
    "PeakUsage" : {
      "committed" : 65011712,
      "init" : 16252928,
      "max" : 344457216,
      "used" : 65011712
    },
    "Usage" : {
      "committed" : 62914560,
      "init" : 16252928,
      "max" : 330825728,
      "used" : 4136568
    },
    "CollectionUsageThresholdExceeded" : false,
    "CollectionUsageThresholdSupported" : true,
    "UsageThresholdSupported" : false,
    "Name" : "PS Eden Space",
    "Type" : "HEAP",
    "Valid" : true,
    "ObjectName" : "java.lang:type=MemoryPool,name=PS Eden Space"
  }, {
    "name" : "java.lang:type=MemoryPool,name=PS Survivor Space",
    "modelerType" : "sun.management.MemoryPoolImpl",
    "CollectionUsage" : {
      "committed" : 8388608,
      "init" : 2621440,
      "max" : 8388608,
      "used" : 8231776
    },
    "CollectionUsageThreshold" : 0,
    "CollectionUsageThresholdCount" : 0,
    "MemoryManagerNames" : [ "PS MarkSweep", "PS Scavenge" ],
    "PeakUsage" : {
      "committed" : 8388608,
      "init" : 2621440,
      "max" : 8388608,
      "used" : 8231776
    },
    "Usage" : {
      "committed" : 8388608,
      "init" : 2621440,
      "max" : 8388608,
      "used" : 8231776
    },
    "CollectionUsageThresholdExceeded" : false,
    "CollectionUsageThresholdSupported" : true,
    "UsageThresholdSupported" : false,
    "Name" : "PS Survivor Space",
    "Type" : "HEAP",
    "Valid" : true,
    "ObjectName" : "java.lang:type=MemoryPool,name=PS Survivor Space"
  }, {
    "name" : "java.lang:type=GarbageCollector,name=PS MarkSweep",
    "modelerType" : "sun.management.GarbageCollectorImpl",
    "LastGcInfo" : null,
    "CollectionCount" : 0,
    "CollectionTime" : 0,
    "MemoryPoolNames" : [ "PS Eden Space", "PS Survivor Space", "PS Old Gen", "PS Perm Gen" ],
    "Name" : "PS MarkSweep",
    "Valid" : true,
    "ObjectName" : "java.lang:type=GarbageCollector,name=PS MarkSweep"
  }, {
    "name" : "java.nio:type=BufferPool,name=mapped",
    "modelerType" : "sun.management.ManagementFactoryHelper$1",
    "MemoryUsed" : 2144,
    "TotalCapacity" : 2144,
    "Name" : "mapped",
    "Count" : 1,
    "ObjectName" : "java.nio:type=BufferPool,name=mapped"
  }, {
    "name" : "java.lang:type=Compilation",
    "modelerType" : "sun.management.CompilationImpl",
    "CompilationTimeMonitoringSupported" : true,
    "TotalCompilationTime" : 7128,
    "Name" : "HotSpot 64-Bit Tiered Compilers",
    "ObjectName" : "java.lang:type=Compilation"
  }, {
    "name" : "Hadoop:service=DataNode,name=FSDatasetState-null",
    "modelerType" : "org.apache.hadoop.hdfs.server.datanode.fsdataset.impl.FsDatasetImpl",
    "Remaining" : 14485221376,
    "StorageInfo" : "FSDataset{dirpath='[/tmp/hadoop-root/dfs/data/current]'}",
    "Capacity" : 41083600896,
    "DfsUsed" : 327680,
    "CacheCapacity" : 0,
    "CacheUsed" : 0,
    "NumFailedVolumes" : 0,
    "FailedStorageLocations" : [ ],
    "LastVolumeFailureDate" : 0,
    "EstimatedCapacityLostTotal" : 0,
    "NumBlocksCached" : 0,
    "NumBlocksFailedToCache" : 0,
    "NumBlocksFailedToUncache" : 0
  }, {
    "name" : "Hadoop:service=DataNode,name=RpcActivityForPort50020",
    "modelerType" : "RpcActivityForPort50020",
    "tag.port" : "50020",
    "tag.Context" : "rpc",
    "tag.Hostname" : "00c3c26f8980",
    "ReceivedBytes" : 0,
    "SentBytes" : 0,
    "RpcQueueTimeNumOps" : 0,
    "RpcQueueTimeAvgTime" : 0.0,
    "RpcProcessingTimeNumOps" : 0,
    "RpcProcessingTimeAvgTime" : 0.0,
    "RpcAuthenticationFailures" : 0,
    "RpcAuthenticationSuccesses" : 0,
    "RpcAuthorizationFailures" : 0,
    "RpcAuthorizationSuccesses" : 0,
    "NumOpenConnections" : 0,
    "CallQueueLength" : 0
  }, {
    "name" : "java.lang:type=OperatingSystem",
    "modelerType" : "com.sun.management.UnixOperatingSystem",
    "MaxFileDescriptorCount" : 1048576,
    "OpenFileDescriptorCount" : 240,
    "CommittedVirtualMemorySize" : 1612926976,
    "FreePhysicalMemorySize" : 113266688,
    "FreeSwapSpaceSize" : 480120832,
    "ProcessCpuLoad" : 0.0037368210329640997,
    "ProcessCpuTime" : 18410000000,
    "SystemCpuLoad" : 0.17589750433738155,
    "TotalPhysicalMemorySize" : 4124778496,
    "TotalSwapSpaceSize" : 1071640576,
    "AvailableProcessors" : 2,
    "Arch" : "amd64",
    "SystemLoadAverage" : 0.5,
    "Name" : "`, s.Os, `",
    "Version" : "4.4.0-128-generic",
    "ObjectName" : "java.lang:type=OperatingSystem"
  }, {
    "name" : "Hadoop:service=DataNode,name=DataNodeActivity-00c3c26f8980-50010",
    "modelerType" : "DataNodeActivity-00c3c26f8980-50010",
    "tag.SessionId" : null,
    "tag.Context" : "dfs",
    "tag.Hostname" : "00c3c26f8980",
    "BytesWritten" : 0,
    "TotalWriteTime" : 0,
    "BytesRead" : 0,
    "TotalReadTime" : 0,
    "BlocksWritten" : 0,
    "BlocksRead" : 0,
    "BlocksReplicated" : 0,
    "BlocksRemoved" : 0,
    "BlocksVerified" : 0,
    "BlockVerificationFailures" : 0,
    "BlocksCached" : 0,
    "BlocksUncached" : 0,
    "ReadsFromLocalClient" : 0,
    "ReadsFromRemoteClient" : 0,
    "WritesFromLocalClient" : 0,
    "WritesFromRemoteClient" : 0,
    "BlocksGetLocalPathInfo" : 0,
    "RemoteBytesRead" : 0,
    "RemoteBytesWritten" : 0,
    "RamDiskBlocksWrite" : 0,
    "RamDiskBlocksWriteFallback" : 0,
    "RamDiskBytesWrite" : 0,
    "RamDiskBlocksReadHits" : 0,
    "RamDiskBlocksEvicted" : 0,
    "RamDiskBlocksEvictedWithoutRead" : 0,
    "RamDiskBlocksEvictionWindowMsNumOps" : 0,
    "RamDiskBlocksEvictionWindowMsAvgTime" : 0.0,
    "RamDiskBlocksLazyPersisted" : 0,
    "RamDiskBlocksDeletedBeforeLazyPersisted" : 0,
    "RamDiskBytesLazyPersisted" : 0,
    "RamDiskBlocksLazyPersistWindowMsNumOps" : 0,
    "RamDiskBlocksLazyPersistWindowMsAvgTime" : 0.0,
    "FsyncCount" : 0,
    "VolumeFailures" : 0,
    "DatanodeNetworkErrors" : 0,
    "ReadBlockOpNumOps" : 0,
    "ReadBlockOpAvgTime" : 0.0,
    "WriteBlockOpNumOps" : 0,
    "WriteBlockOpAvgTime" : 0.0,
    "BlockChecksumOpNumOps" : 0,
    "BlockChecksumOpAvgTime" : 0.0,
    "CopyBlockOpNumOps" : 0,
    "CopyBlockOpAvgTime" : 0.0,
    "ReplaceBlockOpNumOps" : 0,
    "ReplaceBlockOpAvgTime" : 0.0,
    "HeartbeatsNumOps" : 530,
    "HeartbeatsAvgTime" : 8.000000000000002,
    "BlockReportsNumOps" : 1,
    "BlockReportsAvgTime" : 160.0,
    "IncrementalBlockReportsNumOps" : 0,
    "IncrementalBlockReportsAvgTime" : 0.0,
    "CacheReportsNumOps" : 0,
    "CacheReportsAvgTime" : 0.0,
    "PacketAckRoundTripTimeNanosNumOps" : 0,
    "PacketAckRoundTripTimeNanosAvgTime" : 0.0,
    "FlushNanosNumOps" : 0,
    "FlushNanosAvgTime" : 0.0,
    "FsyncNanosNumOps" : 0,
    "FsyncNanosAvgTime" : 0.0,
    "SendDataPacketBlockedOnNetworkNanosNumOps" : 0,
    "SendDataPacketBlockedOnNetworkNanosAvgTime" : 0.0,
    "SendDataPacketTransferNanosNumOps" : 0,
    "SendDataPacketTransferNanosAvgTime" : 0.0
  }, {
    "name" : "java.lang:type=MemoryManager,name=CodeCacheManager",
    "modelerType" : "sun.management.MemoryManagerImpl",
    "MemoryPoolNames" : [ "Code Cache" ],
    "Name" : "CodeCacheManager",
    "Valid" : true,
    "ObjectName" : "java.lang:type=MemoryManager,name=CodeCacheManager"
  }, {
    "name" : "Hadoop:service=DataNode,name=MetricsSystem,sub=Stats",
    "modelerType" : "MetricsSystem,sub=Stats",
    "tag.Context" : "metricssystem",
    "tag.Hostname" : "00c3c26f8980",
    "NumActiveSources" : 5,
    "NumAllSources" : 5,
    "NumActiveSinks" : 0,
    "NumAllSinks" : 0,
    "SnapshotNumOps" : 0,
    "SnapshotAvgTime" : 0.0,
    "PublishNumOps" : 0,
    "PublishAvgTime" : 0.0,
    "DroppedPubAll" : 0
  }, {
    "name" : "Hadoop:service=DataNode,name=UgiMetrics",
    "modelerType" : "UgiMetrics",
    "tag.Context" : "ugi",
    "tag.Hostname" : "00c3c26f8980",
    "LoginSuccessNumOps" : 0,
    "LoginSuccessAvgTime" : 0.0,
    "LoginFailureNumOps" : 0,
    "LoginFailureAvgTime" : 0.0,
    "GetGroupsNumOps" : 0,
    "GetGroupsAvgTime" : 0.0
  }, {
    "name" : "java.lang:type=MemoryPool,name=Code Cache",
    "modelerType" : "sun.management.MemoryPoolImpl",
    "CollectionUsage" : null,
    "MemoryManagerNames" : [ "CodeCacheManager" ],
    "PeakUsage" : {
      "committed" : 2555904,
      "init" : 2555904,
      "max" : 50331648,
      "used" : 1615168
    },
    "Usage" : {
      "committed" : 2555904,
      "init" : 2555904,
      "max" : 50331648,
      "used" : 1615168
    },
    "UsageThreshold" : 0,
    "UsageThresholdCount" : 0,
    "CollectionUsageThresholdSupported" : false,
    "UsageThresholdExceeded" : false,
    "UsageThresholdSupported" : true,
    "Name" : "Code Cache",
    "Type" : "NON_HEAP",
    "Valid" : true,
    "ObjectName" : "java.lang:type=MemoryPool,name=Code Cache"
  }, {
    "name" : "java.lang:type=Runtime",
    "modelerType" : "sun.management.RuntimeImpl",
    "BootClassPath" : "/usr/java/jdk1.7.0_71/jre/lib/resources.jar:/usr/java/jdk1.7.0_71/jre/lib/rt.jar:/usr/java/jdk1.7.0_71/jre/lib/sunrsasign.jar:/usr/java/jdk1.7.0_71/jre/lib/jsse.jar:/usr/java/jdk1.7.0_71/jre/lib/jce.jar:/usr/java/jdk1.7.0_71/jre/lib/charsets.jar:/usr/java/jdk1.7.0_71/jre/lib/jfr.jar:/usr/java/jdk1.7.0_71/jre/classes",
    "LibraryPath" : "/usr/local/hadoop/lib/native",
    "VmName" : "Java HotSpot(TM) 64-Bit Server VM",
    "VmVendor" : "Oracle Corporation",
    "VmVersion" : "24.71-b01",
    "BootClassPathSupported" : true,
    "InputArguments" : [ "-Dproc_datanode", "-Xmx1000m", "-Djava.net.preferIPv4Stack=true", "-Dhadoop.log.dir=/usr/local/hadoop/logs", "-Dhadoop.log.file=hadoop.log", "-Dhadoop.home.dir=/usr/local/hadoop", "-Dhadoop.id.str=root", "-Dhadoop.root.logger=INFO,console", "-Djava.library.path=/usr/local/hadoop/lib/native", "-Dhadoop.policy.file=hadoop-policy.xml", "-Djava.net.preferIPv4Stack=true", "-Djava.net.preferIPv4Stack=true", "-Djava.net.preferIPv4Stack=true", "-Dhadoop.log.dir=/usr/local/hadoop/logs", "-Dhadoop.log.file=hadoop-root-datanode-00c3c26f8980.log", "-Dhadoop.home.dir=/usr/local/hadoop", "-Dhadoop.id.str=root", "-Dhadoop.root.logger=INFO,RFA", "-Djava.library.path=/usr/local/hadoop/lib/native", "-Dhadoop.policy.file=hadoop-policy.xml", "-Djava.net.preferIPv4Stack=true", "-Dhadoop.security.logger=ERROR,RFAS", "-Dhadoop.security.logger=ERROR,RFAS", "-Dhadoop.security.logger=ERROR,RFAS", "-Dhadoop.security.logger=INFO,RFAS" ],
    "ManagementSpecVersion" : "1.2",
    "SpecName" : "Java Virtual Machine Specification",
    "SpecVendor" : "Oracle Corporation",
    "SpecVersion" : "1.7",
    "SystemProperties" : [ {
      "key" : "java.vm.version",
      "value" : "24.71-b01"
    }, {
      "key" : "java.vendor.url",
      "value" : "http://java.oracle.com/"
    }, {
      "key" : "sun.jnu.encoding",
      "value" : "ANSI_X3.4-1968"
    }, {
      "key" : "java.vm.info",
      "value" : "mixed mode"
    }, {
      "key" : "user.dir",
      "value" : "/usr/local/hadoop-`, s.Version, `"
    }, {
      "key" : "sun.cpu.isalist",
      "value" : ""
    }, {
      "key" : "java.awt.graphicsenv",
      "value" : "sun.awt.X11GraphicsEnvironment"
    }, {
      "key" : "sun.os.patch.level",
      "value" : "unknown"
    }, {
      "key" : "hadoop.log.dir",
      "value" : "/usr/local/hadoop/logs"
    }, {
      "key" : "java.io.tmpdir",
      "value" : "/tmp"
    }, {
      "key" : "sun.nio.ch.bugLevel",
      "value" : ""
    }, {
      "key" : "user.home",
      "value" : "/root"
    }, {
      "key" : "java.awt.printerjob",
      "value" : "sun.print.PSPrinterJob"
    }, {
      "key" : "java.version",
      "value" : "1.7.0_71"
    }, {
      "key" : "file.encoding.pkg",
      "value" : "sun.io"
    }, {
      "key" : "java.vendor.url.bug",
      "value" : "http://bugreport.sun.com/bugreport/"
    }, {
      "key" : "file.encoding",
      "value" : "ANSI_X3.4-1968"
    }, {
      "key" : "line.separator",
      "value" : "\n"
    }, {
      "key" : "sun.java.command",
      "value" : "org.apache.hadoop.hdfs.server.datanode.DataNode"
    }, {
      "key" : "java.vm.specification.vendor",
      "value" : "Oracle Corporation"
    }, {
      "key" : "hadoop.id.str",
      "value" : "root"
    }, {
      "key" : "java.vm.vendor",
      "value" : "Oracle Corporation"
    }, {
      "key" : "hadoop.security.logger",
      "value" : "INFO,RFAS"
    }, {
      "key" : "java.class.path",
      "value" : "/usr/local/hadoop/etc/hadoop/:/usr/local/hadoop/share/hadoop/common/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/common/lib/zookeeper-3.4.6.jar:/usr/local/hadoop/share/hadoop/common/lib/paranamer-2.3.jar:/usr/local/hadoop/share/hadoop/common/lib/apacheds-i18n-2.0.0-M15.jar:/usr/local/hadoop/share/hadoop/common/lib/apacheds-kerberos-codec-2.0.0-M15.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-beanutils-core-1.8.0.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-compress-1.4.1.jar:/usr/local/hadoop/share/hadoop/common/lib/java-xmlbuilder-0.4.jar:/usr/local/hadoop/share/hadoop/common/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/common/lib/httpcore-4.2.5.jar:/usr/local/hadoop/share/hadoop/common/lib/gson-2.2.4.jar:/usr/local/hadoop/share/hadoop/common/lib/hadoop-auth-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-collections-3.2.1.jar:/usr/local/hadoop/share/hadoop/common/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/common/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-httpclient-3.1.jar:/usr/local/hadoop/share/hadoop/common/lib/jetty-6.1.26.jar:/usr/local/hadoop/share/hadoop/common/lib/jaxb-impl-2.2.3-1.jar:/usr/local/hadoop/share/hadoop/common/lib/xz-1.0.jar:/usr/local/hadoop/share/hadoop/common/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-digester-1.8.jar:/usr/local/hadoop/share/hadoop/common/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/common/lib/jsch-0.1.42.jar:/usr/local/hadoop/share/hadoop/common/lib/mockito-all-1.8.5.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-logging-1.1.3.jar:/usr/local/hadoop/share/hadoop/common/lib/hadoop-annotations-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/xmlenc-0.52.jar:/usr/local/hadoop/share/hadoop/common/lib/slf4j-log4j12-1.7.10.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-beanutils-1.7.0.jar:/usr/local/hadoop/share/hadoop/common/lib/httpclient-4.2.5.jar:/usr/local/hadoop/share/hadoop/common/lib/jersey-json-1.9.jar:/usr/local/hadoop/share/hadoop/common/lib/jetty-util-6.1.26.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-codec-1.4.jar:/usr/local/hadoop/share/hadoop/common/lib/curator-client-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-configuration-1.6.jar:/usr/local/hadoop/share/hadoop/common/lib/slf4j-api-1.7.10.jar:/usr/local/hadoop/share/hadoop/common/lib/jettison-1.1.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-net-3.1.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-lang-2.6.jar:/usr/local/hadoop/share/hadoop/common/lib/jaxb-api-2.2.2.jar:/usr/local/hadoop/share/hadoop/common/lib/activation-1.1.jar:/usr/local/hadoop/share/hadoop/common/lib/htrace-core-3.1.0-incubating.jar:/usr/local/hadoop/share/hadoop/common/lib/jsr305-3.0.0.jar:/usr/local/hadoop/share/hadoop/common/lib/snappy-java-1.0.4.1.jar:/usr/local/hadoop/share/hadoop/common/lib/api-asn1-api-1.0.0-M20.jar:/usr/local/hadoop/share/hadoop/common/lib/avro-1.7.4.jar:/usr/local/hadoop/share/hadoop/common/lib/servlet-api-2.5.jar:/usr/local/hadoop/share/hadoop/common/lib/jsp-api-2.1.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-xc-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-math3-3.1.1.jar:/usr/local/hadoop/share/hadoop/common/lib/stax-api-1.0-2.jar:/usr/local/hadoop/share/hadoop/common/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/common/lib/junit-4.11.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/hamcrest-core-1.3.jar:/usr/local/hadoop/share/hadoop/common/lib/jets3t-0.9.0.jar:/usr/local/hadoop/share/hadoop/common/lib/curator-recipes-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-jaxrs-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/guava-11.0.2.jar:/usr/local/hadoop/share/hadoop/common/lib/api-util-1.0.0-M20.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-cli-1.2.jar:/usr/local/hadoop/share/hadoop/common/lib/curator-framework-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/hadoop-nfs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/hadoop-common-`, s.Version, `-tests.jar:/usr/local/hadoop/share/hadoop/common/hadoop-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/hdfs:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jetty-6.1.26.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/leveldbjni-all-1.8.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-logging-1.1.3.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/xmlenc-0.52.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jetty-util-6.1.26.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-codec-1.4.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-lang-2.6.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/xercesImpl-2.9.1.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-daemon-1.0.13.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/htrace-core-3.1.0-incubating.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jsr305-3.0.0.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/xml-apis-1.3.04.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/servlet-api-2.5.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/netty-all-4.0.23.Final.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/guava-11.0.2.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-cli-1.2.jar:/usr/local/hadoop/share/hadoop/hdfs/hadoop-hdfs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/hdfs/hadoop-hdfs-`, s.Version, `-tests.jar:/usr/local/hadoop/share/hadoop/hdfs/hadoop-hdfs-nfs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/yarn/lib/zookeeper-3.4.6.jar:/usr/local/hadoop/share/hadoop/yarn/lib/guice-3.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-compress-1.4.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-client-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-collections-3.2.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-guice-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jetty-6.1.26.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jaxb-impl-2.2.3-1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/xz-1.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/leveldbjni-all-1.8.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-logging-1.1.3.jar:/usr/local/hadoop/share/hadoop/yarn/lib/javax.inject-1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/zookeeper-3.4.6-tests.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-json-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jetty-util-6.1.26.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-codec-1.4.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jettison-1.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-lang-2.6.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jaxb-api-2.2.2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/activation-1.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jsr305-3.0.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/servlet-api-2.5.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-xc-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/stax-api-1.0-2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/guice-servlet-3.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-jaxrs-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/guava-11.0.2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/aopalliance-1.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-cli-1.2.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-applicationhistoryservice-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-tests-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-registry-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-resourcemanager-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-applications-unmanaged-am-launcher-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-web-proxy-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-client-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-applications-distributedshell-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-nodemanager-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-sharedcachemanager-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-api-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/paranamer-2.3.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/guice-3.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/commons-compress-1.4.1.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jersey-guice-1.9.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/xz-1.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/leveldbjni-all-1.8.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/hadoop-annotations-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/javax.inject-1.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/snappy-java-1.0.4.1.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/avro-1.7.4.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/junit-4.11.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/hamcrest-core-1.3.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/guice-servlet-3.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/aopalliance-1.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-examples-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-shuffle-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-app-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-jobclient-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-jobclient-`, s.Version, `-tests.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-hs-plugins-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-hs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-core-`, s.Version, `.jar:/usr/local/hadoop/contrib/capacity-scheduler/*.jar:/usr/local/hadoop/contrib/capacity-scheduler/*.jar:/usr/local/hadoop/contrib/capacity-scheduler/*.jar"
    }, {
      "key" : "sun.io.unicode.encoding",
      "value" : "UnicodeLittle"
    }, {
      "key" : "os.arch",
      "value" : "amd64"
    }, {
      "key" : "user.name",
      "value" : "root"
    }, {
      "key" : "user.language",
      "value" : "en"
    }, {
      "key" : "java.runtime.version",
      "value" : "1.7.0_71-b14"
    }, {
      "key" : "sun.boot.class.path",
      "value" : "/usr/java/jdk1.7.0_71/jre/lib/resources.jar:/usr/java/jdk1.7.0_71/jre/lib/rt.jar:/usr/java/jdk1.7.0_71/jre/lib/sunrsasign.jar:/usr/java/jdk1.7.0_71/jre/lib/jsse.jar:/usr/java/jdk1.7.0_71/jre/lib/jce.jar:/usr/java/jdk1.7.0_71/jre/lib/charsets.jar:/usr/java/jdk1.7.0_71/jre/lib/jfr.jar:/usr/java/jdk1.7.0_71/jre/classes"
    }, {
      "key" : "hadoop.log.file",
      "value" : "hadoop-root-datanode-00c3c26f8980.log"
    }, {
      "key" : "sun.cpu.endian",
      "value" : "little"
    }, {
      "key" : "awt.toolkit",
      "value" : "sun.awt.X11.XToolkit"
    }, {
      "key" : "hadoop.root.logger",
      "value" : "INFO,RFA"
    }, {
      "key" : "sun.boot.library.path",
      "value" : "/usr/java/jdk1.7.0_71/jre/lib/amd64"
    }, {
      "key" : "java.vm.name",
      "value" : "Java HotSpot(TM) 64-Bit Server VM"
    }, {
      "key" : "java.home",
      "value" : "/usr/java/jdk1.7.0_71/jre"
    }, {
      "key" : "java.endorsed.dirs",
      "value" : "/usr/java/jdk1.7.0_71/jre/lib/endorsed"
    }, {
      "key" : "java.net.preferIPv4Stack",
      "value" : "true"
    }, {
      "key" : "sun.management.compiler",
      "value" : "HotSpot 64-Bit Tiered Compilers"
    }, {
      "key" : "java.runtime.name",
      "value" : "Java(TM) SE Runtime Environment"
    }, {
      "key" : "java.library.path",
      "value" : "/usr/local/hadoop/lib/native"
    }, {
      "key" : "file.separator",
      "value" : "/"
    }, {
      "key" : "java.specification.vendor",
      "value" : "Oracle Corporation"
    }, {
      "key" : "java.vm.specification.version",
      "value" : "1.7"
    }, {
      "key" : "hadoop.home.dir",
      "value" : "/usr/local/hadoop"
    }, {
      "key" : "sun.java.launcher",
      "value" : "SUN_STANDARD"
    }, {
      "key" : "user.timezone",
      "value" : "America/New_York"
    }, {
      "key" : "os.name",
      "value" : "`, s.Os, `"
    }, {
      "key" : "path.separator",
      "value" : ":"
    }, {
      "key" : "proc_datanode",
      "value" : ""
    }, {
      "key" : "java.ext.dirs",
      "value" : "/usr/java/jdk1.7.0_71/jre/lib/ext:/usr/java/packages/lib/ext"
    }, {
      "key" : "sun.arch.data.model",
      "value" : "64"
    }, {
      "key" : "java.specification.name",
      "value" : "Java Platform API Specification"
    }, {
      "key" : "os.version",
      "value" : "4.4.0-128-generic"
    }, {
      "key" : "hadoop.policy.file",
      "value" : "hadoop-policy.xml"
    }, {
      "key" : "user.country",
      "value" : "US"
    }, {
      "key" : "java.class.version",
      "value" : "51.0"
    }, {
      "key" : "java.vendor",
      "value" : "Oracle Corporation"
    }, {
      "key" : "java.vm.specification.name",
      "value" : "Java Virtual Machine Specification"
    }, {
      "key" : "java.specification.version",
      "value" : "1.7"
    } ],
    "Name" : "208@00c3c26f8980",
    "ClassPath" : "/usr/local/hadoop/etc/hadoop/:/usr/local/hadoop/share/hadoop/common/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/common/lib/zookeeper-3.4.6.jar:/usr/local/hadoop/share/hadoop/common/lib/paranamer-2.3.jar:/usr/local/hadoop/share/hadoop/common/lib/apacheds-i18n-2.0.0-M15.jar:/usr/local/hadoop/share/hadoop/common/lib/apacheds-kerberos-codec-2.0.0-M15.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-beanutils-core-1.8.0.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-compress-1.4.1.jar:/usr/local/hadoop/share/hadoop/common/lib/java-xmlbuilder-0.4.jar:/usr/local/hadoop/share/hadoop/common/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/common/lib/httpcore-4.2.5.jar:/usr/local/hadoop/share/hadoop/common/lib/gson-2.2.4.jar:/usr/local/hadoop/share/hadoop/common/lib/hadoop-auth-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-collections-3.2.1.jar:/usr/local/hadoop/share/hadoop/common/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/common/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-httpclient-3.1.jar:/usr/local/hadoop/share/hadoop/common/lib/jetty-6.1.26.jar:/usr/local/hadoop/share/hadoop/common/lib/jaxb-impl-2.2.3-1.jar:/usr/local/hadoop/share/hadoop/common/lib/xz-1.0.jar:/usr/local/hadoop/share/hadoop/common/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-digester-1.8.jar:/usr/local/hadoop/share/hadoop/common/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/common/lib/jsch-0.1.42.jar:/usr/local/hadoop/share/hadoop/common/lib/mockito-all-1.8.5.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-logging-1.1.3.jar:/usr/local/hadoop/share/hadoop/common/lib/hadoop-annotations-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/xmlenc-0.52.jar:/usr/local/hadoop/share/hadoop/common/lib/slf4j-log4j12-1.7.10.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-beanutils-1.7.0.jar:/usr/local/hadoop/share/hadoop/common/lib/httpclient-4.2.5.jar:/usr/local/hadoop/share/hadoop/common/lib/jersey-json-1.9.jar:/usr/local/hadoop/share/hadoop/common/lib/jetty-util-6.1.26.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-codec-1.4.jar:/usr/local/hadoop/share/hadoop/common/lib/curator-client-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-configuration-1.6.jar:/usr/local/hadoop/share/hadoop/common/lib/slf4j-api-1.7.10.jar:/usr/local/hadoop/share/hadoop/common/lib/jettison-1.1.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-net-3.1.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-lang-2.6.jar:/usr/local/hadoop/share/hadoop/common/lib/jaxb-api-2.2.2.jar:/usr/local/hadoop/share/hadoop/common/lib/activation-1.1.jar:/usr/local/hadoop/share/hadoop/common/lib/htrace-core-3.1.0-incubating.jar:/usr/local/hadoop/share/hadoop/common/lib/jsr305-3.0.0.jar:/usr/local/hadoop/share/hadoop/common/lib/snappy-java-1.0.4.1.jar:/usr/local/hadoop/share/hadoop/common/lib/api-asn1-api-1.0.0-M20.jar:/usr/local/hadoop/share/hadoop/common/lib/avro-1.7.4.jar:/usr/local/hadoop/share/hadoop/common/lib/servlet-api-2.5.jar:/usr/local/hadoop/share/hadoop/common/lib/jsp-api-2.1.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-xc-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-math3-3.1.1.jar:/usr/local/hadoop/share/hadoop/common/lib/stax-api-1.0-2.jar:/usr/local/hadoop/share/hadoop/common/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/common/lib/junit-4.11.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/hamcrest-core-1.3.jar:/usr/local/hadoop/share/hadoop/common/lib/jets3t-0.9.0.jar:/usr/local/hadoop/share/hadoop/common/lib/curator-recipes-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/lib/jackson-jaxrs-1.9.13.jar:/usr/local/hadoop/share/hadoop/common/lib/guava-11.0.2.jar:/usr/local/hadoop/share/hadoop/common/lib/api-util-1.0.0-M20.jar:/usr/local/hadoop/share/hadoop/common/lib/commons-cli-1.2.jar:/usr/local/hadoop/share/hadoop/common/lib/curator-framework-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/hadoop-nfs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/common/hadoop-common-`, s.Version, `-tests.jar:/usr/local/hadoop/share/hadoop/common/hadoop-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/hdfs:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jetty-6.1.26.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/leveldbjni-all-1.8.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-logging-1.1.3.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/xmlenc-0.52.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jetty-util-6.1.26.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-codec-1.4.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-lang-2.6.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/xercesImpl-2.9.1.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-daemon-1.0.13.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/htrace-core-3.1.0-incubating.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jsr305-3.0.0.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/xml-apis-1.3.04.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/servlet-api-2.5.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/netty-all-4.0.23.Final.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/guava-11.0.2.jar:/usr/local/hadoop/share/hadoop/hdfs/lib/commons-cli-1.2.jar:/usr/local/hadoop/share/hadoop/hdfs/hadoop-hdfs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/hdfs/hadoop-hdfs-`, s.Version, `-tests.jar:/usr/local/hadoop/share/hadoop/hdfs/hadoop-hdfs-nfs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/yarn/lib/zookeeper-3.4.6.jar:/usr/local/hadoop/share/hadoop/yarn/lib/guice-3.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-compress-1.4.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-client-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-collections-3.2.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-guice-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jetty-6.1.26.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jaxb-impl-2.2.3-1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/xz-1.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/leveldbjni-all-1.8.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-logging-1.1.3.jar:/usr/local/hadoop/share/hadoop/yarn/lib/javax.inject-1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/zookeeper-3.4.6-tests.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jersey-json-1.9.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jetty-util-6.1.26.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-codec-1.4.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jettison-1.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-lang-2.6.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jaxb-api-2.2.2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/activation-1.1.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jsr305-3.0.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/servlet-api-2.5.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-xc-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/stax-api-1.0-2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/guice-servlet-3.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/jackson-jaxrs-1.9.13.jar:/usr/local/hadoop/share/hadoop/yarn/lib/guava-11.0.2.jar:/usr/local/hadoop/share/hadoop/yarn/lib/aopalliance-1.0.jar:/usr/local/hadoop/share/hadoop/yarn/lib/commons-cli-1.2.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-applicationhistoryservice-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-tests-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-registry-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-resourcemanager-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-applications-unmanaged-am-launcher-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-web-proxy-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-client-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-applications-distributedshell-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-nodemanager-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-server-sharedcachemanager-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/yarn/hadoop-yarn-api-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/commons-io-2.4.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/paranamer-2.3.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/guice-3.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/commons-compress-1.4.1.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jersey-server-1.9.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/protobuf-java-2.5.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jersey-guice-1.9.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/log4j-1.2.17.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/xz-1.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/leveldbjni-all-1.8.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jersey-core-1.9.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/netty-3.6.2.Final.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/hadoop-annotations-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/javax.inject-1.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/snappy-java-1.0.4.1.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/avro-1.7.4.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jackson-core-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/asm-3.2.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/junit-4.11.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/jackson-mapper-asl-1.9.13.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/hamcrest-core-1.3.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/guice-servlet-3.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/lib/aopalliance-1.0.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-examples-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-shuffle-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-common-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-app-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-jobclient-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-jobclient-`, s.Version, `-tests.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-hs-plugins-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-hs-`, s.Version, `.jar:/usr/local/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-client-core-`, s.Version, `.jar:/usr/local/hadoop/contrib/capacity-scheduler/*.jar:/usr/local/hadoop/contrib/capacity-scheduler/*.jar:/usr/local/hadoop/contrib/capacity-scheduler/*.jar",
    "Uptime" : 1594468,
    "StartTime" : 1530624916811,
    "ObjectName" : "java.lang:type=Runtime"
  }, {
    "name" : "java.nio:type=BufferPool,name=direct",
    "modelerType" : "sun.management.ManagementFactoryHelper$1",
    "MemoryUsed" : 100029,
    "TotalCapacity" : 100028,
    "Name" : "direct",
    "Count" : 9,
    "ObjectName" : "java.nio:type=BufferPool,name=direct"
  }, {
    "name" : "Hadoop:service=DataNode,name=DataNodeInfo",
    "modelerType" : "org.apache.hadoop.hdfs.server.datanode.DataNode",
    "XceiverCount" : 1,
    "DatanodeNetworkCounts" : [ ],
    "Version" : "`, s.Version, `",
    "RpcPort" : "50020",
    "HttpPort" : null,
    "NamenodeAddresses" : "{\"00c3c26f8980\":\"BP-1961412683-172.17.0.32-1450036414523\"}",
    "VolumeInfo" : "{\"/tmp/hadoop-root/dfs/data/current\":{\"freeSpace\":14485192704,\"usedSpace\":327680,\"reservedSpace\":0}}",
    "ClusterId" : "CID-5e691286-4de5-4dde-800b-c02a7a8bf44a"
  }, {
    "name" : "java.lang:type=ClassLoading",
    "modelerType" : "sun.management.ClassLoadingImpl",
    "TotalLoadedClassCount" : 4498,
    "Verbose" : false,
    "LoadedClassCount" : 4498,
    "UnloadedClassCount" : 0,
    "ObjectName" : "java.lang:type=ClassLoading"
  }, {
    "name" : "java.lang:type=Threading",
    "modelerType" : "sun.management.ThreadImpl",
    "ThreadAllocatedMemoryEnabled" : true,
    "ThreadAllocatedMemorySupported" : true,
    "ThreadCount" : 37,
    "TotalStartedThreadCount" : 52,
    "ThreadCpuTimeSupported" : true,
    "ThreadCpuTimeEnabled" : true,
    "DaemonThreadCount" : 29,
    "PeakThreadCount" : 39,
    "CurrentThreadCpuTimeSupported" : true,
    "ObjectMonitorUsageSupported" : true,
    "SynchronizerUsageSupported" : true,
    "ThreadContentionMonitoringSupported" : true,
    "CurrentThreadCpuTime" : 286534459,
    "CurrentThreadUserTime" : 220000000,
    "ThreadContentionMonitoringEnabled" : false,
    "AllThreadIds" : [ 58, 27, 26, 25, 24, 56, 52, 49, 48, 46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 30, 33, 14, 34, 32, 31, 29, 20, 19, 18, 16, 13, 4, 3, 2, 1 ],
    "ObjectName" : "java.lang:type=Threading"
  }, {
    "name" : "Hadoop:service=DataNode,name=RpcDetailedActivityForPort50020",
    "modelerType" : "RpcDetailedActivityForPort50020",
    "tag.port" : "50020",
    "tag.Context" : "rpcdetailed",
    "tag.Hostname" : "00c3c26f8980"
  }, {
    "name" : "java.util.logging:type=Logging",
    "modelerType" : "sun.management.ManagementFactoryHelper$PlatformLoggingImpl",
    "ObjectName" : "java.util.logging:type=Logging",
    "LoggerNames" : [ "javax.management.snmp", "global", "javax.management.notification", "javax.management.modelmbean", "javax.management.timer", "com.google.common.collect.MapMakerInternalMap", "com.google.common.cache.CacheBuilder", "javax.management", "com.google.common.cache.LocalCache", "javax.management.mlet", "javax.management.mbeanserver", "javax.management.snmp.daemon", "javax.management.relation", "javax.management.monitor", "javax.management.misc", "" ]
  }, {
    "name" : "Hadoop:service=DataNode,name=JvmMetrics",
    "modelerType" : "JvmMetrics",
    "tag.Context" : "jvm",
    "tag.ProcessName" : "DataNode",
    "tag.SessionId" : null,
    "tag.Hostname" : "00c3c26f8980",
    "MemNonHeapUsedM" : 28.569443,
    "MemNonHeapCommittedM" : 29.9375,
    "MemNonHeapMaxM" : 130.0,
    "MemHeapUsedM" : 23.816345,
    "MemHeapCommittedM" : 109.0,
    "MemHeapMaxM" : 889.0,
    "MemMaxM" : 889.0,
    "GcCount" : 6,
    "GcTimeMillis" : 113,
    "GcNumWarnThresholdExceeded" : 0,
    "GcNumInfoThresholdExceeded" : 0,
    "GcTotalExtraSleepTime" : 2379,
    "ThreadsNew" : 0,
    "ThreadsRunnable" : 12,
    "ThreadsBlocked" : 0,
    "ThreadsWaiting" : 3,
    "ThreadsTimedWaiting" : 22,
    "ThreadsTerminated" : 0,
    "LogFatal" : 0,
    "LogError" : 7,
    "LogWarn" : 1,
    "LogInfo" : 58
  }, {
    "name" : "com.sun.management:type=HotSpotDiagnostic",
    "modelerType" : "sun.management.HotSpotDiagnostic",
    "DiagnosticOptions" : [ {
      "name" : "HeapDumpBeforeFullGC",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "HeapDumpAfterFullGC",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "HeapDumpOnOutOfMemoryError",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "HeapDumpPath",
      "origin" : "DEFAULT",
      "value" : "",
      "writeable" : true
    }, {
      "name" : "CMSAbortablePrecleanWaitMillis",
      "origin" : "DEFAULT",
      "value" : "100",
      "writeable" : true
    }, {
      "name" : "CMSWaitDuration",
      "origin" : "DEFAULT",
      "value" : "2000",
      "writeable" : true
    }, {
      "name" : "PrintGC",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "PrintGCDetails",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "PrintGCDateStamps",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "PrintGCTimeStamps",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "PrintClassHistogramBeforeFullGC",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "PrintClassHistogramAfterFullGC",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "PrintClassHistogram",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "MinHeapFreeRatio",
      "origin" : "DEFAULT",
      "value" : "0",
      "writeable" : true
    }, {
      "name" : "MaxHeapFreeRatio",
      "origin" : "DEFAULT",
      "value" : "100",
      "writeable" : true
    }, {
      "name" : "PrintConcurrentLocks",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    }, {
      "name" : "UnlockCommercialFeatures",
      "origin" : "DEFAULT",
      "value" : "false",
      "writeable" : true
    } ],
    "ObjectName" : "com.sun.management:type=HotSpotDiagnostic"
  }, {
    "name" : "java.lang:type=MemoryPool,name=PS Perm Gen",
    "modelerType" : "sun.management.MemoryPoolImpl",
    "CollectionUsage" : {
      "committed" : 0,
      "init" : 22020096,
      "max" : 85983232,
      "used" : 0
    },
    "CollectionUsageThreshold" : 0,
    "CollectionUsageThresholdCount" : 0,
    "MemoryManagerNames" : [ "PS MarkSweep" ],
    "PeakUsage" : {
      "committed" : 28835840,
      "init" : 22020096,
      "max" : 85983232,
      "used" : 28341080
    },
    "Usage" : {
      "committed" : 28835840,
      "init" : 22020096,
      "max" : 85983232,
      "used" : 28341080
    },
    "UsageThreshold" : 0,
    "UsageThresholdCount" : 0,
    "CollectionUsageThresholdExceeded" : false,
    "CollectionUsageThresholdSupported" : true,
    "UsageThresholdExceeded" : false,
    "UsageThresholdSupported" : true,
    "Name" : "PS Perm Gen",
    "Type" : "NON_HEAP",
    "Valid" : true,
    "ObjectName" : "java.lang:type=MemoryPool,name=PS Perm Gen"
  }, {
    "name" : "Hadoop:service=DataNode,name=MetricsSystem,sub=Control",
    "modelerType" : "org.apache.hadoop.metrics2.impl.MetricsSystemImpl"
  }, {
    "name" : "java.lang:type=GarbageCollector,name=PS Scavenge",
    "modelerType" : "sun.management.GarbageCollectorImpl",
    "LastGcInfo" : {
      "GcThreadCount" : 2,
      "duration" : 34,
      "endTime" : 1404821,
      "id" : 6,
      "memoryUsageAfterGc" : [ {
        "key" : "Code Cache",
        "value" : {
          "committed" : 2555904,
          "init" : 2555904,
          "max" : 50331648,
          "used" : 1577024
        }
      }, {
        "key" : "PS Survivor Space",
        "value" : {
          "committed" : 8388608,
          "init" : 2621440,
          "max" : 8388608,
          "used" : 8231776
        }
      }, {
        "key" : "PS Perm Gen",
        "value" : {
          "committed" : 28835840,
          "init" : 22020096,
          "max" : 85983232,
          "used" : 28334088
        }
      }, {
        "key" : "PS Old Gen",
        "value" : {
          "committed" : 42991616,
          "init" : 42991616,
          "max" : 698875904,
          "used" : 12410976
        }
      }, {
        "key" : "PS Eden Space",
        "value" : {
          "committed" : 62914560,
          "init" : 16252928,
          "max" : 330825728,
          "used" : 0
        }
      } ],
      "memoryUsageBeforeGc" : [ {
        "key" : "Code Cache",
        "value" : {
          "committed" : 2555904,
          "init" : 2555904,
          "max" : 50331648,
          "used" : 1577024
        }
      }, {
        "key" : "PS Survivor Space",
        "value" : {
          "committed" : 2621440,
          "init" : 2621440,
          "max" : 2621440,
          "used" : 2588752
        }
      }, {
        "key" : "PS Perm Gen",
        "value" : {
          "committed" : 28835840,
          "init" : 22020096,
          "max" : 85983232,
          "used" : 28334088
        }
      }, {
        "key" : "PS Old Gen",
        "value" : {
          "committed" : 42991616,
          "init" : 42991616,
          "max" : 698875904,
          "used" : 12410976
        }
      }, {
        "key" : "PS Eden Space",
        "value" : {
          "committed" : 65011712,
          "init" : 16252928,
          "max" : 337641472,
          "used" : 65011712
        }
      } ],
      "startTime" : 1404787
    },
    "CollectionCount" : 6,
    "CollectionTime" : 113,
    "MemoryPoolNames" : [ "PS Eden Space", "PS Survivor Space" ],
    "Name" : "PS Scavenge",
    "Valid" : true,
    "ObjectName" : "java.lang:type=GarbageCollector,name=PS Scavenge"
  }, {
    "name" : "java.lang:type=MemoryPool,name=PS Old Gen",
    "modelerType" : "sun.management.MemoryPoolImpl",
    "CollectionUsage" : {
      "committed" : 0,
      "init" : 42991616,
      "max" : 698875904,
      "used" : 0
    },
    "CollectionUsageThreshold" : 0,
    "CollectionUsageThresholdCount" : 0,
    "MemoryManagerNames" : [ "PS MarkSweep" ],
    "PeakUsage" : {
      "committed" : 42991616,
      "init" : 42991616,
      "max" : 698875904,
      "used" : 12410976
    },
    "Usage" : {
      "committed" : 42991616,
      "init" : 42991616,
      "max" : 698875904,
      "used" : 12410976
    },
    "UsageThreshold" : 0,
    "UsageThresholdCount" : 0,
    "CollectionUsageThresholdExceeded" : false,
    "CollectionUsageThresholdSupported" : true,
    "UsageThresholdExceeded" : false,
    "UsageThresholdSupported" : true,
    "Name" : "PS Old Gen",
    "Type" : "HEAP",
    "Valid" : true,
    "ObjectName" : "java.lang:type=MemoryPool,name=PS Old Gen"
  }, {
    "name" : "JMImplementation:type=MBeanServerDelegate",
    "modelerType" : "javax.management.MBeanServerDelegate",
    "MBeanServerId" : "00c3c26f8980_1530624919441",
    "SpecificationName" : "Java Management Extensions",
    "SpecificationVersion" : "1.4",
    "SpecificationVendor" : "Oracle Corporation",
    "ImplementationName" : "JMX",
    "ImplementationVersion" : "1.7.0_71-b14",
    "ImplementationVendor" : "Oracle Corporation"
  } ]
}
		`)
}
