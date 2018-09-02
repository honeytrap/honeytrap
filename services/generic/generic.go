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
package generic

import (
	"context"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/utils"
	"github.com/op/go-logging"
	"net"
	"fmt"
)

var (
	_   = services.Register("generic", Generic)
	log = logging.MustGetLogger("services/generic")
)

func Generic(options ...services.ServicerFunc) services.Servicer {
	s := &genericService{}

	for _, o := range options {
		o(s)
	}

	return s
}

type genericService struct {
	scr scripter.Scripter
	c   pushers.Channel
}

func (s *genericService) CanHandle(payload []byte) bool {
	return s.scr.CanHandle("generic", string(payload))
}

func (s *genericService) SetScripter(scr scripter.Scripter) {
	s.scr = scr
}

// SetChannel sets the channel for the service
func (s *genericService) SetChannel(c pushers.Channel) {
	s.c = c
}

// Handle handles the request
func (s *genericService) Handle(ctx context.Context, conn net.Conn) error {
	buffer := make([]byte, 4096)
	pConn := utils.PeekConnection(conn)

	// Add the go methods that have to be exposed to the scripts
	if s.scr == nil {
		return fmt.Errorf("%s","undefined scripter")
	}
	connW := s.scr.GetConnection("generic", pConn)

	s.setMethods(connW)

	for {
		//Handle incoming message with the scripter
		n, _ := pConn.Peek(buffer)
		response, err := connW.Handle(string(buffer[:n]))
		if err != nil {
			return err
		} else if response == "_return" {
			// Return called from script
			return nil
		}

		//Write message to the connection
		if _, err := conn.Write([]byte(response)); err != nil {
			return err
		}
	}
	return nil
}
