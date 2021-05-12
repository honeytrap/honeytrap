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
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/services"
)

/*Example config:

[service.hadoop_namenode]
type="hadoop_namenode"
version="2.7.1"
os="Linux"

[[port]]
port="tcp/50070"
services=["hadoop_namenode"]

*/

var (
	_ = services.Register("hadoop_namenode", Hadoop)
)

func (s *hadoopService) HandleNameNode(conn net.Conn, req *http.Request) error {
	hadoopRequest := hadoopRequestNameNode
	s.ch.Send(event.New(
		services.EventOptions,
		event.Service("Hadoop NameNode"),
		event.Category("hadoop_namenode"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.user-agent", req.UserAgent()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
		event.Custom("http.request", req.RequestURI),
	))

	if req.URL.Path != "/jmx" {
		resp := http.Response{
			StatusCode: http.StatusNotFound,
			Status:     http.StatusText(http.StatusNotFound),
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Request:    req,
			Header: map[string][]string{
				"Cache-Control":  []string{"must-revalidate,no-cache,no-store"},
				"Date":           []string{time.Now().Format(http.TimeFormat)},
				"Pragma":         []string{"no-cache"},
				"Content-Type":   []string{"text/html; charset=iso-8859-1"},
				"Content-length": []string{fmt.Sprintf("%d", len(s.htmlErrorPage(req.URL.Path)))},
				"Server":         []string{"Jetty(6.1.26)"},
			},
		}
		resp.Body = ioutil.NopCloser(strings.NewReader(s.htmlErrorPage(req.URL.Path)))
		return resp.Write(conn)
	}

	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Request:    req,
		Header: map[string][]string{
			"Cache-Control":                []string{"no-cache"},
			"Expires":                      []string{time.Now().Format(http.TimeFormat)},
			"Date":                         []string{time.Now().Format(http.TimeFormat)},
			"Pragma":                       []string{"no-cache"},
			"Content-Type":                 []string{"application/json; charset=utf-8"},
			"Access-Control-Allow-Methods": []string{req.Method},
			"Access-Control-Allow-Origin":  []string{"*"},
			"Transfer-Encoding":            []string{"chunked"},
			"Server":                       []string{"Jetty(6.1.26)"},
		},
	}

	resp.Body = ioutil.NopCloser(strings.NewReader(s.ShowRequest(req, hadoopRequest)))

	return resp.Write(conn)
}
