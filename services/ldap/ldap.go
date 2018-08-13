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
	"context"
	"crypto/tls"
	"errors"
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
credentials=["admin:admin", "root:root"]
## rootDSE values, empty values can be omitted
naming-contexts=[ "dc=example,dc=com", "dc=ad,dc=myserver,dc=com" ]
supported-ldap-version=[ "3" ]
#supported-extension=[ "1.3.6.1.4.1.1466.20037" ]
vendor-name=[ "HT Directory Server" ]
vendor-version=[ "0.1.0.0" ]
description=[ "Directory Server" ]
objectclass=[ "dcObject", "organization" ]

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

	store, err := getStorage()
	if err != nil {
		log.Errorf("LDAP: Could not initialize storage. %s", err.Error())
	}

	cert, err := store.Certificate()
	if err != nil {
		log.Errorf("TLS: %s", err.Error())
	}

	s := &ldapService{
		Server: Server{
			Handlers: make([]requestHandler, 0, 4),

			Credentials: []string{"root:root"},

			tlsConfig: &tls.Config{
				Certificates:       []tls.Certificate{*cert},
				InsecureSkipVerify: true,
			},

			DSE: &DSE{
				SupportedLDAPVersion: []string{"2", "3"},
				SupportedExtension:   []string{"1.3.6.1.4.1.1466.20037"},
			},
		},
	}

	for _, o := range options {
		if err := o(s); err != nil {
			log.Warning(err.Error())
		}
	}

	// Set request handlers
	s.setHandlers()

	return s
}

type ldapService struct {
	Server

	*Conn

	wantTLS bool

	c pushers.Channel
}

type eventLog map[string]interface{}

func (s *ldapService) setHandlers() {

	s.Handlers = append(s.Handlers,
		&extFuncHandler{
			tlsFunc: func() error {
				if s.isTLS {
					return errors.New("TLS already established")
				}

				if s.tlsConfig != nil {
					s.wantTLS = true
					return nil
				}
				return errors.New("TLS not available")
			},
		})

	s.Handlers = append(s.Handlers,
		&bindFuncHandler{
			bindFunc: func(binddn string, bindpw []byte) bool {

				var cred strings.Builder // build "name:password" string
				_, err := cred.WriteString(binddn)
				_, err = cred.WriteRune(':') // separator
				_, err = cred.Write(bindpw)
				if err != nil {
					log.Debug("ldap.bind: couldn't construct bind name")
					return false
				}

				// anonymous bind is ok
				if cred.Len() == 1 { // empty credentials (":")
					s.login = ""
					return true
				}

				for _, u := range s.Credentials {
					if u == cred.String() {
						s.login = binddn
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

				// if not authenticated send only rootDSE else nothing
				if req.FilterAttr == "" && req.FilterValue == "*" && !s.isLogin() {
					ret = append(ret, s.DSE.Get())
					return ret
				}

				// produce a single search result that matches whatever
				// they are searching for
				if req.FilterAttr == "uid" || req.FilterAttr == "givenName" {
					ret = append(ret, &SearchResultEntry{
						DN: "cn=" + req.FilterValue + "," + req.BaseDN,
						Attrs: AttributeMap{
							"sn":            []string{req.FilterValue},
							"cn":            []string{req.FilterValue},
							"uid":           []string{req.FilterValue},
							"givenName":     []string{req.FilterValue},
							"homeDirectory": []string{"/home/" + req.FilterValue},
							"objectClass": []string{
								"top",
								"posixAccount",
								"inetOrgPerson",
							},
						},
					})
					return ret
				}
				return nil
			},
		})

	// CatchAll should be the last handler
	s.Handlers = append(s.Handlers,
		&CatchAll{
			isLogin: func() bool {
				return s.isLogin()
			},
		},
	)
}

func (s *ldapService) SetChannel(c pushers.Channel) {

	s.c = c
}

func (s *ldapService) Handle(ctx context.Context, conn net.Conn) error {
	s.wantTLS = false

	s.login = "" // set the anonymous authstate

	s.Conn = NewConn(conn)

	for {

		p, err := ber.ReadPacket(s.ConnReader)
		if err != nil {
			return err
		}

		//ber.PrintPacket(p)

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
					if _, err := s.con.Write(part.Bytes()); err != nil {
						return err
					}
				}
				// request is handled
				break
			}
		}

		if s.wantTLS {
			s.wantTLS = false
			if err := s.StartTLS(s.tlsConfig); err != nil {
				return err
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
	return nil
}
