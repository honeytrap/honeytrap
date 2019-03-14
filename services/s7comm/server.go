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

package s7comm

import (
	"context"
	"net"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/services/s7comm/com"
	logging "github.com/op/go-logging"
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
	c pushers.Channel
	S S7Packet
}

func (s *s7commService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *s7commService) Handle(ctx context.Context, conn net.Conn) error {
	s.parseUserInput()
	var cotp, s7 bool = false, false
	var err error
	for {
		b := make([]byte, 4096)
		bl, err := conn.Read(b)
		b = b[:bl]
		if errCk(err) {
			break
		}
		if len(b) < 1 {
			break
		}
		if !cotp {

			response := s.S.C.connect(b)
			if response != nil {
				len, err := conn.Write(response)
				if err != nil || len < 1 {
					break
				}
				cotp = true
			}
		}
		if cotp {
			P, isS7 := unpackS7(b)

			if isS7 && !s7 {

				if P.S7.Parameter.SetupCom.Function == com.S7ConReq {
					response := S7ConResp(P)
					len, err := conn.Write(response)
					if err != nil || len < 1 {
						break
					}
					s7 = true
				}
			}
			if isS7 && s7 {
				if P.S7.Data.SZLID == 0x0011 {
					len, err := conn.Write(s.S.primRes(P))
					if err != nil || len < 1 {
						break
					}
				}
				if P.S7.Data.SZLID == 0x001c {
					len, err := conn.Write(s.S.secRes(P))
					if err != nil || len < 1 {
						break
					}
					err = conn.Close()
					if err != nil {
						break
					}
				}
			}
		}
	}
	return err
}

func errCk(err error) bool {
	if err != nil {
		if err.Error() == "EOF" {
			return true
		}
	}

	return false
}

func (s *s7commService) parseUserInput() {
	s.S.ui.Hardware = s.Hardware
	s.S.ui.SysName = s.SysName
	s.S.ui.Copyright = s.Copyright
	s.S.ui.Version = s.Version
	s.S.ui.ModType = s.ModType
	s.S.ui.Mod = s.Mod
	s.S.ui.SerialNum = s.SerialNum
	s.S.ui.PlantID = s.PlantID
	s.S.ui.CPUType = s.CPUType
	return
}
