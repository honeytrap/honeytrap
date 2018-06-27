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

type createMsg func(*mongodbService, []string) []byte

var mapRequests = map[string]createMsg{
	"whatsmyuri":       (*mongodbService).uriMsg,
	"buildinfo":        (*mongodbService).buildInfoMsg,
	"getlog":           (*mongodbService).getLogMsg,
	"ismaster":         (*mongodbService).forShellMsg,
	"replsetgetstatus": (*mongodbService).replSetMsg,
	"listdatabases":    (*mongodbService).listDatabases,
	"isMaster":         (*mongodbService).isMasterMsg,
	//...
}

func (s *mongodbService) encodeResp(m *mongoMsg, port string) []byte {
	reqID := encodeHex(m.mh.RequestID+1, 4)
	respID := encodeHex(m.mh.RequestID, 4)

	arg := []string{reqID, respID, port}

	// avoid problem of ismaster/isMaster with op_query
	if m.mh.OpCode != 2004 {
		m.rq = strings.ToLower(m.rq)
	}

	ok := false
	for cmd := range mapRequests {
		if strings.Contains(m.rq, cmd) {
			m.rq = cmd
			ok = true
			break
		}
	}

	// if the request data is not in the mapRequests
	if !ok {
		log.Error("Error: command not implemented: %s", m.rq)
		return []byte("")
	}

	fn := mapRequests[m.rq]
	src := fn(s, arg)

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		fmt.Println(err)
	}
	return dst[:n]
}

func encodeHex(t interface{}, i int) string {
	local := make([]byte, i)
	switch i {
	case 4:
		binary.LittleEndian.PutUint32(local, uint32(t.(int32)))
	case 8:
		binary.LittleEndian.PutUint64(local, uint64(t.(int64)))
	}
	l := make([]byte, hex.EncodedLen(len(local)))
	hex.Encode(l, local)
	return string(l)
}

func localTime() string {
	var t = time.Now().UnixNano() / 1000000
	return encodeHex(t, 8)
}

func (s *mongodbService) isMasterMsg(arg []string) []byte {
	lt := localTime()
	a1 := "ef000000"
	a2 := "010000000800000000000000000000000000000001000000cb0000000869736d" +
		"61737465720001106d617842736f6e4f626a65637453697a650000000001106d61784d657373616" +
		"76553697a65427974657300006cdc02106d61785772697465426174636853697a6500a086010009" +
		"6c6f63616c54696d6500"
	a3 := "106c6f676963616c53657373696f6e54696d656f75744d696e75746573001e00" +
		"0000106d696e5769726556657273696f6e0000000000106d61785769726556657273696f6e00060" +
		"0000008726561644f6e6c790000016f6b00000000000000f03f00"
	return []byte(a1 + arg[0] + arg[1] + a2 + lt + a3)
}

func (s *mongodbService) uriMsg(args []string) []byte {
	src := []byte(args[2])
	port := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(port, src)
	len := "3f000000"
	a1 := "dd07000000000000002a00000002796f7500100000003132372e302e302e313a"
	a2 := "00016f6b00000000000000f03f00"
	return []byte(len + args[0] + args[1] + a1 + string(port) + a2)
}

func (s *mongodbService) forShellMsg(arg []string) []byte {
	len := "e0000000"
	lt := localTime()
	a1 := "dd0700000000000000cb0000000869736d61737465720001106d617842736f6e4" +
		"f626a65637453697a650000000001106d61784d65737361676553697a65427974657300006cdc021" +
		"06d61785772697465426174636853697a6500a0860100096c6f63616c54696d6500"
	a2 := "106c6f676963616c53657373696f6e54696d656f75744d696e75746573001e000" +
		"000106d696e5769726556657273696f6e0000000000106d61785769726556657273696f6e0006000" +
		"00008726561644f6e6c790000016f6b00000000000000f03f00"
	return []byte(len + arg[0] + arg[1] + a1 + lt + a2)
}

func (s *mongodbService) buildInfoMsg(arg []string) []byte {
	len := "d4050000"
	v := hex.EncodeToString([]byte(s.Version))
	vs := []string{}
	for _, s := range strings.Split(s.Version, ".") {
		i, _ := strconv.Atoi(s)
		vs = append(vs, encodeHex(int32(i), 4))
	}
	a1 := "dd0700000000000000bf0500000276657273696f6e0006000000"
	a2 := "000267697456657273696f6e0029000000393538366535353764353465663730" +
		"6639636134623433633236383932636435353235376531613500046d6f64756c657300050000000" +
		"002616c6c6f6361746f72000700000073797374656d00026a617661736372697074456e67696e65" +
		"00060000006d6f7a6a730002737973496e666f000b0000006465707265636174656400047665727" +
		"3696f6e41727261790021000000103000"
	a3 := "103100"
	a4 := "103200"
	a5 := "1033000000000000036f70656e73736c00570000000272756e6e696e67001c0000004f70656e53534c20312e3" +
		"02e326f20203237204d617220323031380002636f6d70696c6564001b0000004f70656e53534c20" +
		"312e302e326e2020372044656320323031370000036275696c64456e7669726f6e6d656e7400e40" +
		"3000002646973746d6f6400010000000002646973746172636800070000007838365f3634000263" +
		"63003c0000002f7573722f62696e2f636c616e673a204170706c65204c4c564d2076657273696f6" +
		"e20392e302e302028636c616e672d3930302e302e33392e322900026363666c61677300e5010000" +
		"2d492f7573722f6c6f63616c2f6f70742f6f70656e73736c2f696e636c756465202d666e6f2d6f6" +
		"d69742d6672616d652d706f696e746572202d666e6f2d7374726963742d616c696173696e67202d" +
		"67676462202d70746872656164202d57616c6c202d577369676e2d636f6d70617265202d576e6f2" +
		"d756e6b6e6f776e2d707261676d6173202d57696e76616c69642d706368202d4f32202d576e6f2d" +
		"756e757365642d6c6f63616c2d7479706564656673202d576e6f2d756e757365642d66756e63746" +
		"96f6e202d576e6f2d756e757365642d707269766174652d6669656c64202d576e6f2d6465707265" +
		"63617465642d6465636c61726174696f6e73202d576e6f2d746175746f6c6f676963616c2d636f6" +
		"e7374616e742d6f75742d6f662d72616e67652d636f6d70617265202d576e6f2d756e757365642d" +
		"636f6e73742d7661726961626c65202d576e6f2d6d697373696e672d627261636573202d576e6f2" +
		"d696e636f6e73697374656e742d6d697373696e672d6f76657272696465202d576e6f2d706f7465" +
		"6e7469616c6c792d6576616c75617465642d65787072657373696f6e202d576e6f2d65786365707" +
		"4696f6e73202d66737461636b2d70726f746563746f722d7374726f6e67202d666e6f2d6275696c" +
		"74696e2d6d656d636d700002637878004d0000002f7573722f62696e2f636c616e672b2b202d737" +
		"4646c69623d6c6962632b2b3a204170706c65204c4c564d2076657273696f6e20392e302e302028" +
		"636c616e672d3930302e302e33392e32290002637878666c616773009e0000002d576f7665726c6" +
		"f616465642d7669727475616c202d576572726f723d756e757365642d726573756c74202d577065" +
		"7373696d697a696e672d6d6f7665202d57726564756e64616e742d6d6f7665202d576e6f2d756e6" +
		"46566696e65642d7661722d74656d706c617465202d576e6f2d696e7374616e74696174696f6e2d" +
		"61667465722d7370656369616c697a6174696f6e202d7374643d632b2b313400026c696e6b666c6" +
		"1677300480000002d4c2f7573722f6c6f63616c2f6f70742f6f70656e73736c2f6c6962202d576c" +
		"2c2d62696e645f61745f6c6f6164202d66737461636b2d70726f746563746f722d7374726f6e670" +
		"0027461726765745f6172636800070000007838365f363400027461726765745f6f730006000000" +
		"6d61634f530000106269747300400000000864656275670000106d617842736f6e4f626a6563745" +
		"3697a6500000000010473746f72616765456e67696e6573004c000000023000080000006465766e" +
		"756c6c0002310011000000657068656d6572616c466f725465737400023200070000006d6d61707" +
		"631000233000b000000776972656454696765720000016f6b00000000000000f03f00"
	return []byte(len + arg[0] + arg[1] + a1 + v + a2 + vs[0] + a3 + vs[1] + a4 + vs[2] + a5)
}

func (s *mongodbService) getLogMsg(arg []string) []byte {
	len := "ce010000"
	a := "dd0700000000000000b901000010746f74616c4c696e65735772697474656e000" +
		"4000000046c6f67008c01000002300039000000323031382d30352d33305431313a32383a33342e" +
		"3735362b30323030204920434f4e54524f4c20205b696e6974616e646c697374656e5d200002310" +
		"074000000323031382d30352d33305431313a32383a33342e3735362b30323030204920434f4e54" +
		"524f4c20205b696e6974616e646c697374656e5d202a2a205741524e494e473a204163636573732" +
		"0636f6e74726f6c206973206e6f7420656e61626c656420666f7220746865206461746162617365" +
		"2e0002320085000000323031382d30352d33305431313a32383a33342e3735362b3032303020492" +
		"0434f4e54524f4c20205b696e6974616e646c697374656e5d202a2a202020202020202020205265" +
		"616420616e642077726974652061636365737320746f206461746120616e6420636f6e666967757" +
		"26174696f6e20697320756e726573747269637465642e0002330039000000323031382d30352d33" +
		"305431313a32383a33342e3735362b30323030204920434f4e54524f4c20205b696e6974616e646" +
		"c697374656e5d200000016f6b00000000000000f03f00"
	return []byte(len + arg[0] + arg[1] + a)
}

func (s *mongodbService) replSetMsg(arg []string) []byte {
	len := "7a000000"
	a := "dd070000000000000065000000016f6b000000000000000000026572726d73670" +
		"01b0000006e6f742072756e6e696e672077697468202d2d7265706c5365740010636f6465004c00" +
		"000002636f64654e616d6500150000004e6f5265706c69636174696f6e456e61626c65640000"
	return []byte(len + arg[0] + arg[1] + a)
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
	nameDB := stringBS + "6e616d65" + zeroBS + encodeHex(int32(len(d.Name)+1), 4) + hex.EncodeToString([]byte(d.Name)) + zeroBS
	dsz, _ := strconv.ParseInt(d.SizeOnDisk, 10, 64)
	sizeDB := doubleBS + "73697a654f6e4469736b" + zeroBS + dbSize(int32(dsz))
	emptyDB := boolBS + "656d707479" + dbEmpty(d.Empty) + zeroBS
	lengthDB := len(headerDB+nameDB+sizeDB+emptyDB+zeroBS)/2 + 1 /*1 for length as int32*/
	s := headerDB + encodeHex(int32(lengthDB), 4) + nameDB + sizeDB + emptyDB + zeroBS
	return int32(lengthDB), s, dsz
}

func (s *mongodbService) listDatabases(arg []string) []byte {
	kvv := 47
	var totalSizeOnDiskValue int32
	var szArrayTotal int32
	var arrayTotal string

	for _, Db := range s.Dbs {
		x, y, z := parseDB(Db, &kvv)
		totalSizeOnDiskValue += int32(z)
		szArrayTotal += x
		arrayTotal += y
	}
	// sectionArray
	docName := "646174616261736573" //"databases"
	sizeArrayTotal := encodeHex(int32(len(arrayTotal)/2+len("aabbccdd")/2)+1, 4)
	sectionArray := arrayBS + docName + zeroBS + sizeArrayTotal + arrayTotal + zeroBS

	// sectionTotalSize
	totalSizeOnDiskField := "746f74616c53697a65" // "totalSize"
	sectionTotalSize := doubleBS + totalSizeOnDiskField + zeroBS + dbSize(totalSizeOnDiskValue)

	// sectionOk
	okField := "6f6b"
	okValue := "000000000000f03f"
	sectionOk := doubleBS + okField + zeroBS + okValue + zeroBS

	// Total
	opCode := encodeHex(int32(2013), 4)
	flagsBits := encodeHex(int32(0), 4)

	sections := sectionArray + sectionTotalSize + sectionOk
	lengthDiffSections := encodeHex(int32(len(sections)/2+len("aabbccdd")/2), 4)

	object := arg[0] + arg[1] + opCode + zeroBS + flagsBits + lengthDiffSections + sections

	lengthObject := encodeHex(int32(len(object)/2+len("aabbccdd")/2), 4)
	objectTotal := lengthObject + object

	return []byte(objectTotal)
}
