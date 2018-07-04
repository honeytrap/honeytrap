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

package hadoop_namenode

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

/*-------- DOCKER CONFIGURATION

[service.hadoop_namenode]
type="hadoop_namenode"

[[port]]
port="tcp/50070"
services=["hadoop_namenode"]

-----------------*/

var (
	_ = services.Register("hadoop_namenode", Hadoop)
)

func Hadoop(options ...services.ServicerFunc) services.Servicer {
	s := &hadoopService{
		hadoopServiceConfig: hadoopServiceConfig{
			Version: "2.7.1",
			Os:      "Linux",
		},
	}
	for _, o := range options {
		o(s)
	}
	return s
}

type hadoopServiceConfig struct {
	Version string
	Os      string
}

type hadoopService struct {
	hadoopServiceConfig

	ch pushers.Channel
}

func (s *hadoopService) SetChannel(ch pushers.Channel) {
	s.ch = ch
}

func ShowRequest(reqMethod, reqUri string, s *hadoopService, conn net.Conn) {
	if reqMethod == "GET" {
		if strings.HasPrefix(reqUri, "/jmx?qry=") {
			reqUri := strings.SplitAfter(reqUri, "/jmx?qry=")
			if strings.HasPrefix(reqUri[1], "Hadoop:") {
				trim_hadoop := strings.SplitAfter(reqUri[1], "Hadoop:")
				request := strings.Split(trim_hadoop[1], ",")
				if len(request) == 2 {
					if request[0] == "service=NameNode" && request[1] == "name=NameNodeInfo" {
						conn.Write([]byte(s.showNamenode()))
					} else if request[0] == "service=NameNode" && request[1] == "name=FSNamesystemState" {
						conn.Write([]byte(s.showFSNamesystemState()))
					} else {
						conn.Write([]byte(s.showNothing()))
					}
				} else {
					conn.Write([]byte(s.showNothing()))
				}
			} else {
				conn.Write([]byte(s.showEmpty()))
			}
		} else {
			conn.Write([]byte(s.showWithoutQuerry()))
		}
	}
}

func (s *hadoopService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()
	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	ShowRequest(req.Method, req.RequestURI, s, conn)

	s.ch.Send(event.New(
		services.EventOptions,
		event.Category("hadoop_namenode"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.user-agent", req.UserAgent()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
	))

	return nil
}
