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
package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

/* Example config:

[service.cwmp]
type="cwmp"
port="tcp/3890"

*/

var (
	_ = Register("cwmp", CWMP)
)

func CWMP(options ...ServicerFunc) Servicer {
	s := &cwmpService{}

	for _, o := range options {
		o(s)
	}

	return s
}

type cwmpService struct {
	c pushers.Channel
}

func (s *cwmpService) CanHandle(payload []byte) bool {
	if bytes.HasPrefix(payload, []byte("GET")) {
		return true
	}
	if bytes.HasPrefix(payload, []byte("POST")) {
		return bytes.Contains(payload, []byte("<")) &&
			(bytes.Contains(payload, []byte("SOAP")) || bytes.Contains(payload, []byte("soap"))) &&
			bytes.Contains(payload, []byte("xml"))
	}
	return false
}

func (s *cwmpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *cwmpService) SetDataDir(string) {}

type functionCall struct {
	method       string
	argumentsXML string
}

// Autogenerated with the help of gnewton/chidley.
type body struct {
	Method *method `xml:",any,omitempty"`
}

type envelope struct {
	XMLName xml.Name `xml:"Envelope,omitempty"`
	Body    *body    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body,omitempty"`
}

type method struct {
	XMLName xml.Name
	Attrs   string `xml:"xmlns u,attr"`
	Body    string `xml:",innerxml"`
}

func parseXML(data []byte) (msg functionCall, err error) {
	var root envelope
	err = xml.Unmarshal(data, &root)
	if err != nil {
		return functionCall{}, err
	}
	defer func() {
		if r := recover(); r != nil {
			/* If elements are missing in the XML, the property access will fail.
			 * This block catches these errors, and returns a plain error instead.
			 */
			msg = functionCall{}
			err = fmt.Errorf("XML panic")
		}
	}()
	return functionCall{
		method:       root.Body.Method.XMLName.Local,
		argumentsXML: root.Body.Method.Body,
	}, nil
}

func (s *cwmpService) Handle(ctx context.Context, conn net.Conn) error {
	for {
		br := bufio.NewReader(conn)

		req, err := http.ReadRequest(br)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		var requestBody []byte
		if req.Method == "POST" {
			requestBody, err = ioutil.ReadAll(req.Body)
			if err != nil {
				return err
			}
		}

		if len(requestBody) > 0 {
			msg, err := parseXML([]byte(requestBody))
			// Send the event even if an error occurs; error out only later
			s.c.Send(event.New(
				EventOptions,
				event.Category("cwmp"),
				event.Type("request"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("http.method", req.Method),
				event.Custom("http.proto", req.Proto),
				event.Custom("http.host", req.Host),
				event.Custom("http.url", req.URL.String()),
				event.Custom("http.body", string(requestBody)),
				event.Custom("cwmp.method", msg.method),
				event.Custom("cwmp.argumentsXML", msg.argumentsXML),
				Headers(req.Header),
			))
			if err != nil {
				return err
			}
		}

		resp := http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Request:    req,
			Header: http.Header{
				"Content-Type": []string{"application/xml; charset=utf-8"},
			},
		}

		err = resp.Write(conn)
		if err != nil {
			return err
		}
	}
}
