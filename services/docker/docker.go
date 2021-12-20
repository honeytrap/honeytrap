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
//
//
//  Documentation: https://docs.docker.com/engine/api/

package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/rs/xid"
)

var versionPattern = regexp.MustCompile(`[/.*/]?version`)
var infoPattern = regexp.MustCompile(`[/.*/]?info`)
var getContainersPattern = regexp.MustCompile(`/.*/containers/json`)
var createContainersPattern = regexp.MustCompile(`/.*/containers/create`)
var killContainersPattern = regexp.MustCompile(`/.*/containers/.*/kill`)
var startContainersPattern = regexp.MustCompile(`/.*/containers/.*/start`)
var waitContainersPattern = regexp.MustCompile(`/.*/containers/.*/wait`)
var attachContainersPattern = regexp.MustCompile(`/.*/containers/.*/attach`)
var createImagesPattern = regexp.MustCompile(`/.*/images/create`)
var getImagesPattern = regexp.MustCompile(`/.*/images/json`)

var (
	_ = services.Register("docker", Docker)
)

// Docker is a placeholder
func Docker(options ...services.ServicerFunc) services.Servicer {
	s := &dockerService{
		dockerServiceConfig: dockerServiceConfig{
			Server: "Docker/19.03.13 (linux)",
		},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type dockerServiceConfig struct {
	Server string `toml:"server"`
}

type dockerService struct {
	dockerServiceConfig

	c pushers.Channel
}

func (s *dockerService) CanHandle(payload []byte) bool {

	if bytes.HasPrefix(payload, []byte("GET")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("HEAD")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("POST")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("PUT")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("DELETE")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("PATCH")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("TRACE")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("CONNECT")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("OPTIONS")) {
		return true
	}

	return false
}

func (s *dockerService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *dockerService) Handle(ctx context.Context, conn net.Conn) error {
	id := xid.New()

	defer conn.Close()

	for {

		br := bufio.NewReader(conn)

		req, err := http.ReadRequest(br)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		body := make([]byte, 1024)

		n, err := req.Body.Read(body)
		if err != nil && err != io.EOF {
			return err
		}

		body = body[:n]

		io.Copy(ioutil.Discard, req.Body)

		var connOptions event.Option = nil

		if ec, ok := conn.(*event.Conn); ok {
			connOptions = ec.Options()
		}

		s.c.Send(event.New(
			services.EventOptions,
			connOptions,
			event.Category("docker"),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("http.sessionid", id.String()),
			event.Custom("http.method", req.Method),
			event.Custom("http.proto", req.Proto),
			event.Custom("http.host", req.Host),
			event.Custom("http.url", req.URL.String()),
			event.Payload(body),
			services.Headers(req.Header),
			services.Cookies(req.Cookies()),
		))

		buff := bytes.Buffer{}
		var status int
		var contentType = "application/json; charset=UTF-8"
		var hijack = false

		if versionPattern.MatchString(req.URL.Path) {
			if err := json.NewEncoder(&buff).Encode(versionResp); err != nil {
				return err
			}

			status = 200

		} else if infoPattern.MatchString(req.URL.Path) {

			if err := json.NewEncoder(&buff).Encode(infoResp); err != nil {
				return err
			}

			status = 200
		} else if getContainersPattern.MatchString(req.URL.Path) {
			if err := json.NewEncoder(&buff).Encode([]string{}); err != nil {
				return err
			}

			status = 200
		} else if createContainersPattern.MatchString(req.URL.Path) {

			if err := json.NewEncoder(&buff).Encode(map[string]interface{}{
				"Id":       "e90e34656806",
				"Warnings": []string{},
			}); err != nil {
				return err
			}

			status = 201

		} else if killContainersPattern.MatchString(req.URL.Path) {
			status = 204
		} else if startContainersPattern.MatchString(req.URL.Path) {
			status = 204

		} else if waitContainersPattern.MatchString(req.URL.Path) {
			// wait a little while to make it look like the container is up
			time.Sleep(time.Duration(2))

		} else if attachContainersPattern.MatchString(req.URL.Path) {
			status = 101
			contentType = "application/vnd.docker.raw-stream"

			hijack = true

		} else if getImagesPattern.MatchString(req.URL.Path) {
			if err := json.NewEncoder(&buff).Encode([]string{}); err != nil {
				return err
			}

			status = 200

		} else if createImagesPattern.MatchString(req.URL.Path) {

			m, _ := url.ParseQuery(req.URL.RawQuery)

			var image = m["fromImage"][0]
			var tag string

			if _, ok := m["tag"]; ok {
				tag = m["tag"][0]
			} else {
				tag = "latest"
			}

			if err := json.NewEncoder(&buff).Encode(imageCreateResp(image, tag)[0]); err != nil {
				return err
			}

			buff.WriteString("\n")

			status = 200
		} else if req.URL.Path == "/_ping" {

			status = http.StatusOK
			buff.WriteString("OK")

		} else {

			if err := json.NewEncoder(&buff).Encode(map[string]interface{}{
				"message": "page not found",
			}); err != nil {
				return err
			}
			status = 400
		}

		resp := http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Request:    req,
			Header:     http.Header{},
		}

		if !hijack {
			resp.Header.Add("Content-Length", fmt.Sprintf("%d", buff.Len()))
			resp.Header.Add("Server", s.Server)
			resp.Header.Add("Api-Version", "1.40")
			resp.Header.Add("Docker-Experimental", "false")
			resp.Header.Add("Ostype", "linux")

		} else {
			resp.Header.Add("Connection", "Upgrade")
			resp.Header.Add("Upgrade", "tcp")

		}
		resp.Header.Add("Content-Type", contentType)

		resp.Body = ioutil.NopCloser(&buff)

		if err := resp.Write(conn); err != nil {
			return err
		}

		if hijack {
			// This response is made up and docker never does this, but here
			// until we make it interactive.

			conn.Write([]byte{1, 0, 0, 0, 0, 0, 0, 20})
			conn.Write([]byte("Container started..\n"))
		}

		return nil
	}
}

func imageCreateResp(image string, tag string) []interface{} {
	var resp = []interface{}{
		map[string]interface{}{
			"status": fmt.Sprintf("Pulling from %s:%s", image, tag),
			"id":     "latest",
		},

		map[string]interface{}{
			"status": "Digest: sha256:c95a8e48bf88e9849f3e0f723d9f49fa12c5a00cfc6e60d2bc99d87555295e4c",
		},

		map[string]interface{}{
			"status": fmt.Sprintf("Status: Image is up to date for %s:%s", image, tag),
		},
	}

	return resp

}

var infoResp = map[string]interface{}{
	"ID":                "22PP:LAT7:H243:YGMT:K3FM:N23K:T5IP:A5EJ:HIXW:3ETS:CUXA:BGNP",
	"Containers":        0,
	"ContainersRunning": 0,
	"ContainersPaused":  0,
	"ContainersStopped": 8,
	"Images":            45,
	"Driver":            "overlay2",
	"DriverStatus": []interface{}{
		[]interface{}{
			"Backing Filesystem",
			"extfs",
		},
		[]interface{}{
			"Supports d_type",
			"true",
		},
		[]interface{}{
			"Native Overlay Diff",
			"true",
		},
	},
	"SystemStatus": nil,
	"Plugins": map[string]interface{}{
		"Volume": []interface{}{
			"local",
		},
		"Network": []interface{}{
			"bridge",
			"host",
			"ipvlan",
			"macvlan",
			"null",
			"overlay",
		},
		"Authorization": nil,
		"Log": []interface{}{
			"awslogs",
			"fluentd",
			"gcplogs",
			"gelf",
			"journald",
			"json-file",
			"local",
			"logentries",
			"splunk",
			"syslog",
		},
	},
	"MemoryLimit":        true,
	"SwapLimit":          true,
	"KernelMemory":       true,
	"KernelMemoryTCP":    true,
	"CpuCfsPeriod":       true,
	"CpuCfsQuota":        true,
	"CPUShares":          true,
	"CPUSet":             true,
	"PidsLimit":          true,
	"IPv4Forwarding":     true,
	"BridgeNfIptables":   true,
	"BridgeNfIp6tables":  true,
	"Debug":              false,
	"NFd":                21,
	"OomKillDisable":     true,
	"NGoroutines":        35,
	"SystemTime":         "2021-01-02T16:33:57.921068574Z",
	"LoggingDriver":      "json-file",
	"CgroupDriver":       "cgroupfs",
	"NEventsListener":    0,
	"KernelVersion":      "5.8.0-31-generic",
	"OperatingSystem":    "Ubuntu 20.10",
	"OSType":             "linux",
	"Architecture":       "x86_64",
	"IndexServerAddress": "https://index.docker.io/v1/",
	"RegistryConfig": map[string]interface{}{
		"AllowNondistributableArtifactsCIDRs":     []interface{}{},
		"AllowNondistributableArtifactsHostnames": []interface{}{},
		"InsecureRegistryCIDRs": []interface{}{
			"127.0.0.0/8",
		},
		"IndexConfigs": map[string]interface{}{
			"docker.io": map[string]interface{}{
				"Name":     "docker.io",
				"Mirrors":  []interface{}{},
				"Secure":   true,
				"Official": true,
			},
		},
		"Mirrors": []interface{}{},
	},
	"NCPU":              8,
	"MemTotal":          16348065792,
	"GenericResources":  nil,
	"DockerRootDir":     "/var/lib/docker",
	"HttpProxy":         "",
	"HttpsProxy":        "",
	"NoProxy":           "",
	"Name":              "docker",
	"Labels":            []interface{}{},
	"ExperimentalBuild": false,
	"ServerVersion":     "19.03.13",
	"ClusterStore":      "",
	"ClusterAdvertise":  "",
	"Runtimes": map[string]interface{}{
		"runc": map[string]interface{}{
			"path": "runc",
		},
	},
	"DefaultRuntime": "runc",
	"Swarm": map[string]interface{}{
		"NodeID":           "",
		"NodeAddr":         "",
		"LocalNodeState":   "inactive",
		"ControlAvailable": false,
		"Error":            "",
		"RemoteManagers":   nil,
	},
	"LiveRestoreEnabled": false,
	"Isolation":          "",
	"InitBinary":         "docker-init",
	"ContainerdCommit": map[string]interface{}{
		"ID":       "",
		"Expected": "",
	},
	"RuncCommit": map[string]interface{}{
		"ID":       "",
		"Expected": "",
	},
	"InitCommit": map[string]interface{}{
		"ID":       "",
		"Expected": "",
	},
	"SecurityOptions": []interface{}{
		"name=apparmor",
		"name=seccomp,profile=default",
	},
	"Warnings": nil,
}

var versionResp = map[string]interface{}{
	"Platform": map[string]interface{}{
		"Name": "",
	},
	"Components": []interface{}{
		map[string]interface{}{
			"Name":    "Engine",
			"Version": "19.03.13",
			"Details": map[string]interface{}{
				"ApiVersion":    "1.40",
				"Arch":          "amd64",
				"BuildTime":     "2020-10-14T13:25:32.000000000+00:00",
				"Experimental":  "false",
				"GitCommit":     "4484c46",
				"GoVersion":     "go1.13.8",
				"KernelVersion": "5.8.0-31-generic",
				"MinAPIVersion": "1.12",
				"Os":            "linux",
			},
		},
		map[string]interface{}{
			"Name":    "containerd",
			"Version": "1.3.7-0ubuntu3",
			"Details": map[string]interface{}{
				"GitCommit": "",
			},
		},
		map[string]interface{}{
			"Name":    "runc",
			"Version": "spec: 1.0.1-dev",
			"Details": map[string]interface{}{
				"GitCommit": "",
			},
		},
		map[string]interface{}{
			"Name":    "docker-init",
			"Version": "0.18.0",
			"Details": map[string]interface{}{
				"GitCommit": "",
			},
		},
	},
	"Version":       "19.03.13",
	"ApiVersion":    "1.40",
	"MinAPIVersion": "1.12",
	"GitCommit":     "4484c46",
	"GoVersion":     "go1.13.8",
	"Os":            "linux",
	"Arch":          "amd64",
	"KernelVersion": "5.8.0-31-generic",
	"BuildTime":     "2020-10-14T13:25:32.000000000+00:00",
}
