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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	zeroBS     = "00"
	doubleBS   = "01"
	stringBS   = "02"
	documentBS = "03"
	arrayBS    = "04"
	boolBS     = "08"
)

var mapCommands = map[string]func(*mongodbService, *MsgData) []byte{
	"whatsmyuri":       (*mongodbService).uriMsg,
	"buildinfo":        (*mongodbService).buildInfoMsg,
	"getlog":           (*mongodbService).getLogMsg,
	"ismaster":         (*mongodbService).forShellMsg,
	"replsetgetstatus": (*mongodbService).replSetMsg,
	"listdatabases":    (*mongodbService).listDatabases,
	"isMaster":         (*mongodbService).isMasterHandshakeMsg,
	"saslstart":        (*mongodbService).saslstartMsg,
	//...
}

func (s *mongodbService) respHandler(md *MsgData) []byte {

	// avoid problem of ismaster/isMaster with op_query
	if md.mh.OpCode != 2004 {
		md.cmd = strings.ToLower(md.cmd)
	}

	fn, ok := mapCommands[md.cmd]

	if !ok {
		log.Error("Error: command not implemented: %s", md.cmd)
	}
	src := fn(s, md)

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Error("Error encoding hex response: %s", err.Error())
	}

	return dst[:n]
}

func localTime() string {
	t := time.Now().UnixNano() / 1000000
	return encodeInt(t, 8)
}

func encodeInt(t interface{}, i int) string {
	src := make([]byte, i)
	switch i {
	case 4:
		binary.LittleEndian.PutUint32(src, uint32(t.(int32)))
	case 8:
		binary.LittleEndian.PutUint64(src, uint64(t.(int64)))
	}
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return string(dst)
}

func (s *mongodbService) saslstartMsg(md *MsgData) []byte {
	return []byte("")
}

func (s *mongodbService) isMasterHandshakeMsg(md *MsgData) []byte {
	t := localTime()
	// orm := OpReplyMsg{}
	a1 := "cd0000000000000000000000010000000800000000000000000000000000000001000000a900000008" +
		"69736d61737465720001106d617842736f6e4f626a65637453697a650000000001106d61784d657373" +
		"61676553697a65427974657300006cdc02106d61785772697465426174636853697a6500e803000009" +
		"6c6f63616c54696d6500"
	a2 := "106d61785769726556657273696f6e0005000000106d696e5769726556657273696f6e00000000000872" +
		"6561644f6e6c790000016f6b00000000000000f03f00"
	msg := a1 + t + a2
	return []byte(msg)

}

func (s *mongodbService) uriMsg(md *MsgData) []byte {
	src := []byte(md.port)
	port := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(port, src)
	len := "3f000000"
	a1 := "db0700002a00000002796f7500100000003132372e302e302e313a"
	a2 := "00016f6b00000000000000f03f000500000000"
	id := encodeInt(md.mh.RequestID, 4)
	msg := len + id + id + a1 + string(port) + a2

	return []byte(msg)
}

func (s *mongodbService) buildInfoMsg(md *MsgData) []byte {

	v := hex.EncodeToString([]byte(s.Version))
	vs := []string{}
	for _, s := range strings.Split(s.Version, ".") {
		i, _ := strconv.Atoi(s)
		vs = append(vs, encodeInt(int32(i), 4))
	}

	var a1, a11, a12, a13, a14, a2, a3, a4, a5 string

	id := encodeInt(md.mh.RequestID, 4)

	a11 = "db070000"

	a13 = "0276657273696f6e00"
	a14 = encodeInt(int32(len(s.Version)+1), 4) //  "07000000"     3.4.14 = 6 + 0x00 = 7

	// total len of document
	if a14 == "07000000" {
		a12 = "b5050000"
	} else if a14 == "06000000" {
		a12 = "b4050000"
	}

	a1 = a11 + a12 + a13 + a14

	a2 = "000267697456657273696f6e0029000000393538366535353764353465663730" +
		"6639636134623433633236383932636435353235376531613500046d6f64756c657300050000000" +
		"002616c6c6f6361746f72000700000073797374656d00026a617661736372697074456e67696e65" +
		"00060000006d6f7a6a730002737973496e666f000b0000006465707265636174656400047665727" +
		"3696f6e41727261790021000000103000"
	a3 = "103100"
	a4 = "103200"

	a5 = "1033000000000000036f70656e73736c00570000000272756e6e696e67001c0000004f70656e53534" +
		"c20312e302e326f20203237204d617220323031380002636f6d70696c6564001b0000004f70656e53" +
		"534c20312e302e326e2020372044656320323031370000036275696c64456e7669726f6e6d656e7400d903000002646973" +
		"746d6f6400010000000002646973746172636800070000007838365f363400026363003c0000002f7573722f62696e2f63" +
		"6c616e673a204170706c65204c4c564d2076657273696f6e20392e302e302028636c616e672d3930302e302e33392e3229" +
		"00026363666c61677300050200002d492f7573722f6c6f63616c2f6f70742f6f70656e73736c2f696e636c756465202d66" +
		"6e6f2d6f6d69742d6672616d652d706f696e746572202d666e6f2d7374726963742d616c696173696e67202d6767646220" +
		"2d70746872656164202d57616c6c202d577369676e2d636f6d70617265202d576e6f2d756e6b6e6f776e2d707261676d61" +
		"73202d57696e76616c69642d706368202d4f32202d576e6f2d756e757365642d6c6f63616c2d7479706564656673202d57" +
		"6e6f2d756e757365642d66756e6374696f6e202d576e6f2d756e757365642d707269766174652d6669656c64202d576e6f" +
		"2d646570726563617465642d6465636c61726174696f6e73202d576e6f2d746175746f6c6f676963616c2d636f6e737461" +
		"6e742d6f75742d6f662d72616e67652d636f6d70617265202d576e6f2d756e757365642d636f6e73742d7661726961626c" +
		"65202d576e6f2d6d697373696e672d627261636573202d576e6f2d696e636f6e73697374656e742d6d697373696e672d6f" +
		"76657272696465202d576e6f2d706f74656e7469616c6c792d6576616c75617465642d65787072657373696f6e202d6673" +
		"7461636b2d70726f746563746f722d7374726f6e67202d576e6f2d6e756c6c2d636f6e76657273696f6e202d6d6d61636f" +
		"73782d76657273696f6e2d6d696e3d31302e3133202d666e6f2d6275696c74696e2d6d656d636d700002637878003e0000" +
		"002f7573722f62696e2f636c616e672b2b3a204170706c65204c4c564d2076657273696f6e20392e302e302028636c616e" +
		"672d3930302e302e33392e32290002637878666c61677300600000002d576f7665726c6f616465642d7669727475616c20" +
		"2d5770657373696d697a696e672d6d6f7665202d57726564756e64616e742d6d6f7665202d576e6f2d756e646566696e65" +
		"642d7661722d74656d706c617465202d7374643d632b2b313100026c696e6b666c616773006c0000002d4c2f7573722f6c" +
		"6f63616c2f6f70742f6f70656e73736c2f6c6962202d70746872656164202d576c2c2d62696e645f61745f6c6f6164202d" +
		"66737461636b2d70726f746563746f722d7374726f6e67202d6d6d61636f73782d76657273696f6e2d6d696e3d31302e31" +
		"3300027461726765745f6172636800070000007838365f363400027461726765745f6f7300040000006f73780000106269" +
		"747300400000000864656275670000106d617842736f6e4f626a65637453697a6500000000010473746f72616765456e67" +
		"696e6573004c000000023000080000006465766e756c6c0002310011000000657068656d6572616c466f72546573740002" +
		"3200070000006d6d61707631000233000b000000776972656454696765720000016f6b00000000000000f03f000500000000"

	msg := id + id + a1 + v + a2 + vs[0] + a3 + vs[1] + a4 + vs[2] + a5
	len := encodeInt(int32(4+len(msg)/2), 4)
	msg = len + msg

	return []byte(msg)
}

func (s *mongodbService) forShellMsg(md *MsgData) []byte {
	lt := localTime()

	id := encodeInt(md.mh.RequestID, 4)
	len := "be000000"
	a1 := "db070000a90000000869736d61737465720001106d617842736f6e4f626a65637453697a6500000000011" +
		"06d61784d65737361676553697a65427974657300006cdc02106d61785772697465426174636853697a65" +
		"00e8030000096c6f63616c54696d6500"
	a2 := "106d61785769726556657273696f6e0005000000106d696e5769726556657273696f6e000000000008726" +
		"561644f6e6c790000016f6b00000000000000f03f000500000000"

	msg := len + id + id + a1 + lt + a2
	return []byte(msg)
}

func startupTime() string {
	src := []byte(time.Now().Format("2006-01-02T15:04:05.999-0700"))
	t := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(t, src)

	return string(t)
}

func (s *mongodbService) getLogMsg(md *MsgData) []byte {
	len := "ce010000"

	id := encodeInt(md.mh.RequestID, 4)
	a1 := "db070000b901000010746f74616c4c696e65735772697474656e0004000000046c6f67008c01000002300" +
		"039000000"

	a2 := "204920434f4e54524f4c" +
		"20205b696e6974616e646c697374656e5d200002310074000000"

	a3 := "204920434f4e54524f4c20205b696e6974616e646c697374656e5d202a2a20" +
		"5741524e494e473a2041636365737320636f6e74726f6c206973206e6f7420656e61626c656420666f722" +
		"07468652064617461626173652e0002320085000000"

	a4 := "204920434f4e54524f4c20205b696e6974616e646c697374656e5d202a2a20202020202" +
		"0202020205265616420616e642077726974652061636365737320746f206461746120616e6420636f6e66" +
		"696775726174696f6e20697320756e726573747269637465642e0002330039000000"

	a5 := "204920434f4e54524f4c20205b696e6974616e646c6973" +
		"74656e5d200000016f6b00000000000000f03f000500000000"

	msg := len + id + id + a1 + startupTime() + a2 + startupTime() + a3 + startupTime() + a4 + startupTime() + a5
	return []byte(msg)
}

func (s *mongodbService) replSetMsg(md *MsgData) []byte {
	len := "7a000000"

	id := encodeInt(md.mh.RequestID, 4)
	a := "db07000065000000016f6b000000000000000000026572726d7367001b0000006e6f742072756e6e696e" +
		"672077697468202d2d7265706c5365740010636f6465004c00000002636f64654e616d6500150000004e" +
		"6f5265706c69636174696f6e456e61626c656400000500000000"
	msg := len + id + id + a

	return []byte(msg)
}

func dbSize(s int32) string {
	h := math.Float64bits(float64(s))
	local := make([]byte, 8)
	binary.LittleEndian.PutUint64(local, h)
	l := make([]byte, hex.EncodedLen(len(local)))
	hex.Encode(l, local)
	return string(l)
}

func dbEmpty(b string) string {
	if b == "true" {
		return "01"
	}
	return "00"
}

func parseDB(d Db, kvv *int) (int32, string, int64) {
	headerDB := documentBS + fmt.Sprintf("%x", *kvv) + zeroBS
	nameDB := stringBS + "6e616d65" + zeroBS + encodeInt(int32(len(d.Name)+1), 4) + hex.EncodeToString([]byte(d.Name)) + zeroBS
	dsz, _ := strconv.ParseInt(d.SizeOnDisk, 10, 64)
	sizeDB := doubleBS + "73697a654f6e4469736b" + zeroBS + dbSize(int32(dsz))
	emptyDB := boolBS + "656d707479" + dbEmpty(d.Empty) + zeroBS
	lengthDB := len(headerDB+nameDB+sizeDB+emptyDB+zeroBS)/2 + 1 /*1 for length as int32*/
	s := headerDB + encodeInt(int32(lengthDB), 4) + nameDB + sizeDB + emptyDB + zeroBS
	return int32(lengthDB), s, dsz
}

func (s *mongodbService) listDatabases(md *MsgData) []byte {
	kvv := 47
	var totalSizeOnDiskValue int32
	var szArrayTotal int32
	var arrayTotal string

	args1 := encodeInt(md.mh.RequestID, 4)

	for _, Db := range s.Dbs {
		x, y, z := parseDB(Db, &kvv)
		totalSizeOnDiskValue += int32(z)
		szArrayTotal += x
		arrayTotal += y
	}
	// sectionArray
	docName := "646174616261736573" //"databases"
	sizeArrayTotal := encodeInt(int32(len(arrayTotal)/2+len("aabbccdd")/2)+1, 4)
	sectionArray := arrayBS + docName + zeroBS + sizeArrayTotal + arrayTotal + zeroBS

	// sectionTotalSize
	totalSizeOnDiskField := "746f74616c53697a65" // "totalSize"
	sectionTotalSize := doubleBS + totalSizeOnDiskField + zeroBS + dbSize(totalSizeOnDiskValue)

	// sectionOk
	okField := "6f6b"
	okValue := "000000000000f03f"
	sectionOk := doubleBS + okField + zeroBS + okValue + zeroBS

	// Total
	opCode := encodeInt(int32(2011), 4)

	sections := sectionArray + sectionTotalSize + sectionOk
	lengthDiffSections := encodeInt(int32(len(sections)/2+len("aabbccdd")/2), 4)

	object := args1 + args1 + opCode + lengthDiffSections + sections + "0500000000"

	lengthObject := encodeInt(int32(len(object)/2+len("aabbccdd")/2), 4)
	objectTotal := lengthObject + object

	return []byte(objectTotal)
}

/*

type OpReplyMsg struct {
	MsgHeader
	responseFlags  int32
	cursorID       int64
	startingFrom   int32
	numberReturned int32
	documents      []document
}

func (mh *MsgHeader) encodeMsgHeader() {
	mh.RequestID = decodeInt32(bb)
	mh.ResponseTo = decodeInt32(bb)
	mh.OpCode = encodeInt(t, 4)
}

func encodeDocument() {}

func encodeOpReply()

*/
