/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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
package ldap

import (
	"bufio"
	"context"
	"net"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

/*

[service.ldap]
type="ldap"
credentials=[ "user:password", "admin:admin" ]

[[port]]
port="tcp/389"
services=[ "ldap" ]

[[port]]
port="udp/389"
services=[ "ldap" ]

*/

var (
	_   = services.Register("ldap", LDAP)
	log = logging.MustGetLogger("services/ldap")
)

// LDAP service setup
func LDAP(options ...services.ServicerFunc) services.Servicer {

	s := &ldapService{
		Server: Server{
			Handlers:    make([]requestHandler, 0, 4),
			Credentials: []string{"root:root"},
			anon:        true,
		},
	}

	// Set requestHandlers
	s.setHandlers()

	for _, o := range options {
		if err := o(s); err != nil {
			log.Warning(err.Error())
		}
	}

	return s
}

type ldapService struct {
	Server

	c pushers.Channel
}

//Server ldap server data
type Server struct {
	Handlers []requestHandler

	Credentials []string `toml:"credentials"`

	anon bool // anonymous authenticated, false: user is logged in
}

type eventLog map[string]interface{}

func (s *ldapService) setHandlers() {

	s.Handlers = append(s.Handlers,
		&bindFuncHandler{
			bindFunc: func(binddn string, bindpw []byte) bool {

				// check for anonymous authentication
				if binddn == "" {
					s.anon = true // set the anonymous auth flag
					return true
				}
				var cred strings.Builder // build "name:password" string
				_, err := cred.WriteString(binddn)
				_, err = cred.WriteRune(':') // separator
				_, err = cred.Write(bindpw)
				if err != nil {
					log.Debug("ldap.bind: couldn't construct bind name")
					return false
				}

				for _, u := range s.Credentials {
					if u == cred.String() {
						s.anon = false
						return true
					}
				}
				return false
			},
		})

	s.Handlers = append(s.Handlers,
		&searchFuncHandler{
			searchFunc: func(req *SearchRequest) []*SearchResultEntry {

				ret := make([]*SearchResultEntry, 0, 1)

				// if anonymous auth send only rootDSE
				if s.anon {
				}

				// produce a single search result that matches whatever
				// they are searching for
				if req.FilterAttr == "uid" {
					ret = append(ret, &SearchResultEntry{
						DN: "cn=" + req.FilterValue + "," + req.BaseDN,
						Attrs: map[string]interface{}{
							"sn":            req.FilterValue,
							"cn":            req.FilterValue,
							"uid":           req.FilterValue,
							"homeDirectory": "/home/" + req.FilterValue,
							"objectClass": []string{
								"top",
								"posixAccount",
								"inetOrgPerson",
							},
						},
					})
				}
				return ret
			},
		},
	)

	// CatchAll should be the last handler
	s.Handlers = append(s.Handlers,
		&CatchAll{
			catchallFunc: func() bool {
				return s.anon
			},
		})
}

func (s *ldapService) SetChannel(c pushers.Channel) {

	s.c = c
}

func (s *ldapService) SetDataDir(string) {}

func (s *ldapService) Handle(ctx context.Context, conn net.Conn) error {

	s.anon = true // start with anonymous authstate

	br := bufio.NewReader(conn)

	for {

		p, err := ber.ReadPacket(br)
		if err != nil {
			return err
		}

		// check if packet is readable
		id, err := messageID(p)
		if err != nil {
			return err
		}

		// Storage for events
		elog := make(eventLog)

		elog["ldap.message-id"] = id

		// check if this is an unbind, if so we can close immediately

		if isUnbindRequest(p) {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ldap"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ldap.message-id", id),
				event.Custom("ldap.request-type", "unbind"),
			))

			// We don't have to return anything, so we just close the connection
			return nil
		}

		// Handle request and create a response packet(ASN.1 BER)
		for _, h := range s.Handlers {
			plist := h.handle(p, elog)

			if len(plist) > 0 {
				for _, part := range plist {
					if _, err := conn.Write(part.Bytes()); err != nil {
						return err
					}
				}
				// Handled the request, break out of the handling loop
				break
			}
		}

		// Send Message Data
		s.c.Send(event.New(
			services.EventOptions,
			event.Category("ldap"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.CopyFrom(elog),
		))

	}
}
