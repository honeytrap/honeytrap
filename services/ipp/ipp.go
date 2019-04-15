// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ipp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("services/ipp")

var (
	_ = services.Register("ipp", IPP)
)

func IPP(options ...services.ServicerFunc) services.Servicer {
	s := &ippService{
		Config: Config{
			Banner:     "hplj1020",
			StorageDir: "",
			SizeLimit:  104857600,
		},
	}

	for _, o := range options {
		o(s)
	}

	model.val = append(model.val, &valStr{nameWithoutLang, "printer-name", []string{s.PrinterName}})

	return s
}

type Config struct {
	Banner string `toml:"server"`

	StorageDir string `toml:"storage-dir"`

	PrinterName string `toml:"printer-name"`

	SizeLimit int `toml:"size-treshold"`
}

type ippService struct {
	Config

	ch pushers.Channel
}

func (s *ippService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *ippService) Handle(ctx context.Context, conn net.Conn) error {

	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)

	if err == io.EOF {
		return nil
	} else if err != nil {
		log.Error("Bad ipp request: %s", err.Error())
		return err
	}

	if req.Method != "POST" {
		log.Error("Bad ipp request, request method: %s", req.Method)
		return nil
	} else if contentType := req.Header.Get("Content-Type"); contentType != "application/ipp" {
		log.Error("Bad ipp request, wrong content-type")
		return nil
	}

	ippReq, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Error reading request: ", err.Error())
		return err
	}

	if err := req.Body.Close(); err != nil {
		return err
	}

	ippResp, err := ippHandler(ippReq)
	if err != nil {
		return err
	}

	if len(ippResp.data) == 0 {
		// no print data
	} else if s.StorageDir == "" {
		// no storage location
	} else if len(ippResp.data) > s.SizeLimit {
		log.Debug("Data exceeds size limit")
	} else {
		ext := ""

		switch ippResp.format {
		case "application/pdf":
			ext = ".pdf"
		case "image/pwg-raster":
			ext = ".ras"
		case "application/octet-stream":
			ext = ".octet-stream"
		}

		p := path.Join(s.StorageDir, fmt.Sprintf("%s%s", time.Now().Format("ipp-20060102150405"), ext))
		log.Debugf("Data size %v, file %v", len(ippResp.data), p)

		ioutil.WriteFile(p, ippResp.data, 0644)

		ippResp.data = []byte("Print data stored to " + p)
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
			"Server":        []string{s.Banner},
			"Content-Type":  []string{"application/ipp"},
			"Cache-Control": []string{"no-cache"},
			"Pragma":        []string{"no-cache"},
		},
		Close: true,
		Body:  ioutil.NopCloser(rbody), //need io.ReadCloser,
	}

	if err := resp.Write(conn); err != nil {
		log.Error("error writing response: %s", err.Error())
		return err
	}

	return nil
}
