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

 BRIEF EXPLANATION OF THE S7COMM PROTOCOL

                                       +--------+-------+------------+----------+
 S7 Telegram / PDU                     | HEADER | PARAM | PARAM DATA | DATA     |
                                       +--------+-------+------------+----------+
                                       ^                                        ^
                         +------+------+----------------------------------------+
 ISO on TCP              | TKTP | COTP |            S7 PDU                      |
                         +------+------+----------------------------------------+
                         ^                                                      ^
              +----------+------------------------------------------------------+
 TCP/IP       | HEADER   |  ISO TCP TELEGRAM                                    |
              +----------+------------------------------------------------------+

source: http://gmiru.com/article/s7comm/

COTP: RFC905
TPKT: RFC1006

*/

package s7comm

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"strconv"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/op/go-logging"
)

var (
	_ = services.Register("s7comm", S7)
)

var log = logging.MustGetLogger("services")

func S7(options ...services.ServicerFunc) services.Servicer {
	s := &s7commService{
		s7commServiceConfig: s7commServiceConfig{},
	}

	for _, o := range options {
		_ = o(s)
	}
	return s
}

type s7commServiceConfig struct {
	Hardware  string `toml:"basic_hardware"`
	SysName   string `toml:"system_name"`
	Copyright string `toml:"copyright"`
	Version   string `toml:"version"`
	ModType   string `toml:"module_type"`
	Mod       string `toml:"module"`
	SerialNum string `toml:"serial_number"`
	PlantID   string `toml:"plant_identification"`
	CPUType   string `toml:"cpu_type"`
}

type s7commService struct {
	s7commServiceConfig
	channel  pushers.Channel
	s7packet S7Packet
	iot      Packet
}

func (s7service *s7commService) SetChannel(c pushers.Channel) {
	s7service.channel = c
}

func (s7service *s7commService) Handle(ctx context.Context, conn net.Conn) error {
	s7service.parseUserInput()
	var isCotp, isS7Connected = false, false
	var err error
	for {
		buf := make([]byte, 4096)
		buflen, err := conn.Read(buf)
		buf = buf[:buflen]
		if errorCheck(err) {
			break
		}
		if len(buf) < 1 {
			break
		}
		if !isCotp {
			response := s7service.s7packet.C.connect(buf)
			if response != nil {

				s7service.channel.Send(event.New(
					services.EventOptions,
					event.Category("s7comm"),
					event.Type("ics"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("request.type", "COTP connection request"),
					event.Payload(buf),
				))

				length, err := conn.Write(response)
				if err != nil || length < 1 {
					break
				}
				isCotp = true
			}
		}
		if isCotp {
			//Check if ID is 32 or 72

			if len(buf) < 8 {
				break
			}

			if buf[7] != 0x72 {

				P, isS7 := s7service.s7packet.deserialize(buf)
				if isS7 && !isS7Connected {
					if P.S7.Parameter.SetupCom.Function == S7ConReq {

						s7service.channel.Send(event.New(
							services.EventOptions,
							event.Category("s7comm"),
							event.Type("ics"),
							event.SourceAddr(conn.RemoteAddr()),
							event.DestinationAddr(conn.LocalAddr()),
							event.Custom("request.type", "S7comm job request"),
							event.Payload(buf),
						))

						response := s7service.s7packet.connect(P)
						length, err := conn.Write(response)
						if err != nil || length < 1 {
							break
						}
						isS7Connected = true
					}
				}
				if isS7 && isS7Connected {

					if P.S7.Header.MessageType == 1 {
						requestType := lookupJobRequest(P.S7)
						if requestType != "" {
							s7service.channel.Send(event.New(
								services.EventOptions,
								event.Category("s7comm"),
								event.Type("ics"),
								event.SourceAddr(conn.RemoteAddr()),
								event.DestinationAddr(conn.LocalAddr()),
								event.Custom("request.type", requestType),
								event.Payload(buf),
							))
						}

					} else {
						reqID := P.S7.Data.SZLID
						if reqID != 0 {
							r := s7service.handleEvent(reqID, conn, buf)
							if r != nil {
								length, err := conn.Write(r)
								if err != nil || length < 1 {
									break
								}
							}
						}
					}
				}
			} else if buf[7] == 0x72 {
				var s7Plus Plus

				if buf[8] == 0x01 {
					S7CPD, resp := s7Plus.connect(buf)

					if resp != nil {
						length, err := conn.Write(resp)
						if err != nil || length < 1 {
							break
						}
					}

					s7service.channel.Send(event.New(
						services.EventOptions,
						event.Category("s7comm"),
						event.Type("ics"),
						event.SourceAddr(conn.RemoteAddr()),
						event.DestinationAddr(conn.LocalAddr()),
						event.Custom("request.type", "Received S7CommPlus Connection Request"),
						event.Custom("hostname", S7CPD.hostname),
						event.Custom("interface", S7CPD.networkInt),
						event.Custom("data_type", S7CPD.dataType),
						event.Payload(buf),
					))

				} else  {

					resp := s7Plus.randomResponse(buf)
					if resp != nil {
						length, err := conn.Write(resp)
						if err != nil || length < 1 {
							break
						}
					}

					s7service.channel.Send(event.New(
						services.EventOptions,
						event.Category("s7comm"),
						event.Type("ics"),
						event.SourceAddr(conn.RemoteAddr()),
						event.DestinationAddr(conn.LocalAddr()),
						event.Custom("request.type", "Received S7CommPlus Request"),
						event.Payload(buf),
					))


				}
			}
		}
		if !isCotp && !isS7Connected {
			s7service.channel.Send(event.New(
				services.EventOptions,
				event.Category("s7comm"),
				event.Type("ics"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("request.type", "Received unknown request"),
				event.Payload(buf),
			))

		}
	}

	return err
}

func lookupJobRequest(packet S7Packet) (rt string) {

	switch packet.Parameter.SetupCom.Function {
	case 0x00:
		log.Info("Diagnostics request")
		rt = "diagnostics request"
	case 0x04:
		log.Info("Read request")
		rt = "read request"
	case 0x05:
		log.Info("Write request")
		rt = "write request"
	case 0x1a:
		log.Info("Download request")
		rt = "download request"
	case 0x1b:
		log.Info("Download block")
		rt = "download block"
	case 0x1c:
		log.Info("End download")
		rt = "end download"
	case 0x1d:
		log.Info("Start upload")
		rt = "start upload"
	case 0x1e:
		log.Info("Upload")
		rt = "upload"
	case 0x1f:
		log.Info("End upload")
		rt = "end upload"
	case 0x28:
		log.Info("Insert block")
		rt = "insert block"
	default:
		rt = ""

	}
	return
}

func (s7service *s7commService) handleEvent(reqID uint16, conn net.Conn, b []byte) (r []byte) {

	var rt string
	var resp []byte

	switch reqID {
	case 0x11:
		log.Info("Module ID list requested")
		rt = "module ID request"
		resp = s7service.s7packet.primRes(s7service.iot)
	case 0x1c:
		log.Info("Component ID list requested")
		rt = "component ID request"
		resp = s7service.s7packet.secRes(s7service.iot)
	default:
		log.Info("Received unknown request")
		rt = "unknown request"
		resp = nil
	}

	s7service.channel.Send(event.New(
		services.EventOptions,
		event.Category("s7comm"),
		event.Type("ics"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("request.ID", strconv.Itoa(int(reqID))),
		event.Custom("request.type", rt),
		event.Payload(b),
	))

	return resp
}

func (s7service *s7commService) parseUserInput() {
	s7service.s7packet.ui.Hardware = s7service.Hardware
	s7service.s7packet.ui.SysName = s7service.SysName
	s7service.s7packet.ui.Copyright = s7service.Copyright
	s7service.s7packet.ui.Version = s7service.Version
	s7service.s7packet.ui.ModType = s7service.ModType
	s7service.s7packet.ui.Mod = s7service.Mod
	s7service.s7packet.ui.SerialNum = s7service.SerialNum
	s7service.s7packet.ui.PlantID = s7service.PlantID
	s7service.s7packet.ui.CPUType = s7service.CPUType
	return
}

func errorCheck(err error) bool {
	if err != nil {
		if err.Error() == "EOF" {
			return true
		}
	}

	return false
}

type errHandler struct {
	err error
}

func (errorHandler *errHandler) serializer(buf *bytes.Buffer, i interface{}) {
	if errorHandler.err != nil {
		return
	}
	errorHandler.err = binary.Write(buf, binary.BigEndian, i)
}
