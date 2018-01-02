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
package ipp

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("services")

var (
	_ = services.Register("ipp", IPP)
)

func IPP(options ...services.ServicerFunc) services.Servicer {
	s := &ippService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type Config struct {
	HttpServer string `toml:"server"`

	SizeLimit int `toml:"size-treshold"`
}

type ippService struct {
	Config

	ch pushers.Channel
}

func (s *ippService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *ippService) Handle(conn net.Conn) error {

	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)

	if err == io.EOF {
		log.Debug("IPP: Bad ipp request")
		return nil
	} else if err != nil {
		return err
	}
	if req.Method != "POST" || req.Header.Get("Content-Type") != "application/ipp" {
		log.Debug("IPP: Bad ipp request")
		return nil
	}

	//TODO: This could exhaust memory!
	ippReq, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Debug("IPP: error reading IPP request!")
		return err
	}
	if err := req.Body.Close(); err != nil {
		return err
	}

	ippResp, err := IPPHandler(ippReq)
	if err != nil {
		return err
	}

	//Check for print
	if ippResp.data != nil {
		log.Debug("IPP: Print received")

		dir := "/tmp/"
		ext := ""

		switch ippResp.format {
		case "application/pdf":
			ext = ".pdf"
		case "image/pwg-raster":
			ext = ".ras"
		case "application/octet-stream":
			ext = ".raw"
		}

		if len(ippResp.data) > s.SizeLimit {
			fname := dir + ippResp.jobname + ext
			ioutil.WriteFile(fname, ippResp.data, 0600)
			ippResp.data = []byte("Print data stored as " + fname)
		}

	}

	s.ch.Send(event.New(
		services.EventOptions,
		event.Category("ipp"),
		event.Type("request"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.url", req.URL.String()),
		event.Custom("ipp.uri", ippResp.uri),
		event.Custom("ipp.user", ippResp.username),
		event.Custom("ipp.job-name", string(ippResp.jobname)),
		event.Custom("ipp.data", string(ippResp.data)),
	))

	rbody := ippResp.encode()

	resp := http.Response{
		StatusCode:    http.StatusOK,
		Status:        http.StatusText(http.StatusOK),
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		Request:       req,
		ContentLength: int64(rbody.Len()),
		Header: map[string][]string{
			"Server":        []string{s.HttpServer},
			"Content-Type":  []string{"application/ipp"},
			"Cache-Control": []string{"no-cache"},
			"Pragma":        []string{"no-cache"},
		},
		Close: true,
		Body:  ioutil.NopCloser(rbody), //need io.ReadCloser,
	}

	if err := resp.Write(conn); err != nil {
		log.Debug("IPP: error writing respons!")
		return err
	}

	return nil
}
