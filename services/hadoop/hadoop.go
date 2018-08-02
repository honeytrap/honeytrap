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
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

var (
	hadoopRequestNameNode = map[string]func(*hadoopService) string{
		"service=NameNode,name=NameNodeInfo":      (*hadoopService).showNameNode,
		"service=NameNode,name=FSNamesystemState": (*hadoopService).showFSNamesystemState,
	}
	hadoopRequestDataNode = map[string]func(*hadoopService) string{
		"service=DataNode,name=DataNodeInfo": (*hadoopService).showDataNode,
	}
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
	Version string `toml:"version"`
	Os      string `toml:"os"`
}

type hadoopService struct {
	hadoopServiceConfig

	ch pushers.Channel
}

func (s *hadoopService) SetChannel(ch pushers.Channel) {
	s.ch = ch
}

func (s *hadoopService) ShowRequest(req *http.Request, hadoopRequest map[string]func(*hadoopService) string) string {
	for i, _ := range req.Form {
		if !strings.Contains(strings.Join(req.Form[i], ""), ":") {
			return s.showEmpty()
		}
		request := strings.Split(strings.Join(req.Form[i], ""), ":")
		switch request[0] {
		case "Hadoop":
			fn, ok := hadoopRequest[request[1]]
			if !strings.Contains(request[1], ",") {
				return s.showEmpty()
			}
			if !ok {
				return s.showNothing()
			}
			return fn(s)
		default:
			return s.showNothing()
		}
	}
	return ""
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

	err = req.ParseForm()
	if err != nil {
		return err
	}

	port := conn.LocalAddr().(*net.TCPAddr).Port

	if port == 50070 {
		return s.HandleNameNode(conn, req)
	} else if port == 50075 {
		return s.HandleDataNode(conn, req)
	}

	return nil
}
