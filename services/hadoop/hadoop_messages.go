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
package hadoop

import (
	"fmt"
	"time"
)

func getDayLetter() string {
	return fmt.Sprint(time.Now().Weekday(), time.Month(time.Now().Month()), time.Now().Day())
}

func getDayNumber() string {
	return fmt.Sprintf("%v-%d-%v", time.Now().Year(), time.Month(time.Now().Month()), time.Now().Day())
}

func getTime() string {
	return fmt.Sprintf("%v:%v:%v", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
}

func (s *hadoopService) htmlErrorPage(reqPath string) string {
	return fmt.Sprintln(`
		<html>
		<head>
		<meta http-equiv="Content-Type" content="text/html; charset=ISO-8859-1"/>
		<title>Error 404 NOT_FOUND</title>
		</head>
		<body><h2>HTTP ERROR 404</h2>
		<p>Problem accessing `, reqPath, `. Reason:
		<pre>    NOT_FOUND</pre></p><hr /><i><small>Powered by Jetty://</small></i><br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>
		<br/>

		</body>
		</html>
		`)
}

func (s *hadoopService) showEmpty() string {
	return `{ }`
}

func (s *hadoopService) showNothing() string {
	return `
{
	"beans": []
}`
}

func (s *hadoopService) showFSNamesystemState() string {
	return fmt.Sprintln(`
{
  "beans" : [ {
    "name" : "Hadoop:service=NameNode,name=FSNamesystemState",
    "modelerType" : "org.apache.hadoop.hdfs.server.namenode.FSNamesystem",
    "CapacityTotal" : 41083600896,
    "CapacityUsed" : 327680,
    "CapacityRemaining" : 14096678912,
    "TotalLoad" : 1,
    "SnapshotStats" : "{\"SnapshottableDirectories\":0,\"Snapshots\":0}",
    "BlocksTotal" : 31,
    "MaxObjects" : 0,
    "FilesTotal" : 35,
    "PendingReplicationBlocks" : 0,
    "UnderReplicatedBlocks" : 0,
    "ScheduledReplicationBlocks" : 0,
    "PendingDeletionBlocks" : 0,
    "BlockDeletionStartTime" : 1532001516964,
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
    "TopUserOpCounts" : "{\"timestamp\":\"`, getDayNumber(), `T`, getTime(), `-0400\",\"windows\":[{\"windowLenMs\":300000,\"ops\":[]},{\"windowLenMs\":1500000,\"ops\":[]},{\"windowLenMs\":60000,\"ops\":[]}]}"
  } ]
}`)
}

func (s *hadoopService) showNameNode() string {
	return fmt.Sprintln(`
{
  "beans" : [ {
    "name" : "Hadoop:service=NameNode,name=NameNodeInfo",
    "modelerType" : "org.apache.hadoop.hdfs.server.namenode.FSNamesystem",
    "UpgradeFinalized" : true,
    "ClusterId" : "CID-5e691286-4de5-4dde-800b-c02a7a8bf44a",
    "Version" : "`, s.Version, `, r15ecc87ccf4a0228f35af08fc56de536e6ce657a",
    "Used" : 327680,
    "Free" : 14100385792,
    "Safemode" : "",
    "NonDfsUsedSpace" : 26982887424,
    "PercentUsed" : 7.975932E-4,
    "BlockPoolUsedSpace" : 327680,
    "PercentBlockPoolUsed" : 7.975932E-4,
    "PercentRemaining" : 34.321205,
    "CacheCapacity" : 0,
    "CacheUsed" : 0,
    "TotalBlocks" : 31,
    "TotalFiles" : 35,
    "NumberOfMissingBlocks" : 0,
    "NumberOfMissingBlocksWithReplicationFactorOne" : 0,
    "LiveNodes" : "{\"4a41865ca8bc:50010\":{\"infoAddr\":\"172.17.0.2:50075\",\"infoSecureAddr\":\"172.17.0.2:0\",\"xferaddr\":\"172.17.0.2:50010\",\"lastContact\":0,\"usedSpace\":327680,\"adminState\":\"In Service\",\"nonDfsUsedSpace\":26982887424,\"capacity\":41083600896,\"numBlocks\":31,\"version\":\"`, s.Version, `\",\"used\":327680,\"remaining\":14100385792,\"blockScheduled\":0,\"blockPoolUsed\":327680,\"blockPoolUsedPercent\":7.975932E-4,\"volfails\":0}}",
    "DeadNodes" : "{}",
    "DecomNodes" : "{}",
    "BlockPoolId" : "BP-1961412683-172.17.0.32-1450036414523",
    "NameDirStatuses" : "{\"failed\":{},\"active\":{\"/tmp/hadoop-root/dfs/name\":\"IMAGE_AND_EDITS\"}}",
    "NodeUsage" : "{\"nodeUsage\":{\"min\":\"0.00%\",\"median\":\"0.00%\",\"max\":\"0.00%\",\"stdDev\":\"0.00%\"}}",
    "NameJournalStatus" : "[{\"stream\":\"EditLogFileOutputStream(/tmp/hadoop-root/dfs/name/current/edits_inprogress_0000000000000000192)\",\"manager\":\"FileJournalManager(root=/tmp/hadoop-root/dfs/name)\",\"required\":\"false\",\"disabled\":\"false\"}]",
    "JournalTransactionInfo" : "{\"LastAppliedOrWrittenTxId\":\"192\",\"MostRecentCheckpointTxId\":\"191\"}",
    "NNStarted" : `, getDayLetter(), `"07:58:36 EDT 2018",
    "CompileInfo" : "2015-06-29T06:04Z by jenkins from (detached from 15ecc87)",
    "CorruptFiles" : "[]",
    "DistinctVersionCount" : 1,
    "DistinctVersions" : [ {
      "key" : "`, s.Version, `",
      "value" : 1
    } ],
    "SoftwareVersion" : "`, s.Version, `",
    "RollingUpgradeStatus" : null,
    "Threads" : 33,
    "Total" : 41083600896
  } ]
}`)
}

func (s *hadoopService) showDataNode() string {
	return fmt.Sprintln(`
{
  "beans" : [ {
    "name" : "Hadoop:service=DataNode,name=DataNodeInfo",
    "modelerType" : "org.apache.hadoop.hdfs.server.datanode.DataNode",
    "XceiverCount" : 1,
    "DatanodeNetworkCounts" : [ ],
    "Version" : "`, s.Version, `",
    "RpcPort" : "50020",
    "HttpPort" : null,
    "NamenodeAddresses" : "{\"4a41865ca8bc\":\"BP-1961412683-172.17.0.32-1450036414523\"}",
    "VolumeInfo" : "{\"/tmp/hadoop-root/dfs/data/current\":{\"freeSpace\":14075846656,\"usedSpace\":327680,\"reservedSpace\":0}}",
    "ClusterId" : "CID-5e691286-4de5-4dde-800b-c02a7a8bf44a"
  } ]
}`)
}
