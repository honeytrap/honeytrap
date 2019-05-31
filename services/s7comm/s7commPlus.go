/* Copyright 2016-2019 DutchSec (https://dutchsec.com/)
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package s7comm

import (
	"github.com/google/netstack/rand"
	"github.com/honeytrap/honeytrap/services/decoder"
)
func (s7p *S7CommPlus) receiveCommand(m []byte) () {
}

func (s7p *S7CommPlus) connect(m []byte) (Data S7ComPlusData, resp []byte) {

	// We're skipping TKTP & COPT check
	dec := decoder.NewDecoder(m[7:])
	s7p.ID = dec.Byte()
	s7p.PDUType = dec.Byte()
	s7p.DataLen = dec.Uint16()
	s7p.Reserved = dec.Uint16()
	s7p.SubType = dec.Uint16()
	s7p.SecNum = dec.Uint32()

	/* Pattern Found! This is raw data sent by the s7-1200 Metasploit plugin
	i_1_ 2_ 3_4_ 5_ data -------------------------------------------------------------------------------------
	a382 21 0015 2c 313a3a3a362e303a3a5443502f4950202d3e20496e74656c2852292050524f2f31303030204d54204e2e2e2e <--- ,1:::6.0::TCP/IP -> Intel(R) PRO/1000 MT N...
	a382 28 0015 00
	a382 29 0015 00
	a382 2a 0015 0e 4841434b2d50435f383832333330 <--- HACK-PC_882330
	a382 2b 0004 01
	a382 2c 0012 01 c9c380
	a382 2d 0015 00 a1000000d3817f0000
	a381 69 0015 15 537562736372697074696f6e436f6e7461696e6572a 2a200000000
					SubscriptionContainer
	*/

	var S7PD S7ComPlusData
	for i := 0; i < len(m)- 1; i++ {

		// Searching for data block specifier: "0xa3, 0x8x"
		if m[i] == 0xa3 && m[i+1] > 0x7f && m[i+1] < 0x90 {

			// check for pattern inside datablock: "0x00, 0x15"
			if m[i+3] == 0x00 && m[i+4] == 0x15 {

				// the next byte contains message length
				mlen := int(m[i+5])
				if mlen > 0 {
					// extract slice from message length to index of message length + message length
					msg := string(m[i+6 : i+6+mlen])

					switch m[i+2] {
					case 0x69:
						S7PD.dataType = msg
					case 0x2a:
						S7PD.hostname = msg
					case 0x21:
						S7PD.networkInt = msg
					default:
					}

				}

			}

		}
	}
	//Sending back 25 bytes of random data
	resp = make([]byte, 25)
	rand.Read(resp)
	return S7PD, resp
}