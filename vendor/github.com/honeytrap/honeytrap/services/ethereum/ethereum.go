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
package ethereum

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"io/ioutil"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var (
	_ = services.Register("ethereum", Ethereum)
)

var log = logging.MustGetLogger("honeytrap/services/ethereum")

// Ethereum is a placeholder
func Ethereum(options ...services.ServicerFunc) services.Servicer {
	s := &ethereumService{
		ethereumServiceConfig: ethereumServiceConfig{},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type ethereumServiceConfig struct {
	Name        string `toml:"name"`
	ClusterName string `toml:"cluster_name"`
	ClusterUUID string `toml:"cluster_uuid"`
}

type ethereumService struct {
	ethereumServiceConfig

	c pushers.Channel
}

func (s *ethereumService) SetChannel(c pushers.Channel) {
	s.c = c
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

var ethereumMethods = map[string]func(map[string]interface{}) map[string]interface{}{
	"eth_getBalance": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  "0x0234c8a3397aab58", // 158972490234375000
		}
	},
	"net_version": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  "1",
		}
	},
	"miner_setEtherbase": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  true,
		}
	},
	"eth_mining": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  true,
		}
	},
	"eth_coinbase": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  []string{"0x407d73d8a49eeb85d32cf465507dd71d507100c1"},
		}
	},
	"eth_accounts": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  []string{"0x407d73d8a49eeb85d32cf465507dd71d507100c1"},
		}
	},
	"eth_blockNumber": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  "0x4b7",
		}
	},
	"web3_clientVersion": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  "Mist/v0.9.3/darwin/go1.4.1",
		}
	},
	"eth_getBlockByNumber": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result": map[string]interface{}{
				"number":           "0x1b4",
				"hash":             "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
				"parentHash":       "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
				"nonce":            "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
				"sha3Uncles":       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
				"logsBloom":        "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
				"transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"stateRoot":        "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
				"miner":            "0x4e65fda2159562a496f9f3522f89122a3088497a",
				"difficulty":       "0x027f07",
				"totalDifficulty":  "0x027f07",
				"extraData":        "0x0000000000000000000000000000000000000000000000000000000000000000",
				"size":             "0x027f07",
				"gasLimit":         "0x9f759",
				"minGasPrice":      "0x9f759",
				"gasUsed":          "0x9f759",
				"timestamp":        "0x54e34e8e",
				"transactions":     []interface{}{},
				"uncles":           []interface{}{},
			},
		}
	},
	"eth_sendTransaction": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
		}
	},
	"personal_unlockAccount": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  true,
		}
	},
	"eth_getTransactionCount": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result":  "0x1",
		}
	},
	"rpc_modules": func(m map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"id":      m["id"],
			"jsonrpc": m["jsonrpc"],
			"result": map[string]string{
				"admin":    "1.0",
				"db":       "1.0",
				"debug":    "1.0",
				"eth":      "1.0",
				"miner":    "1.0",
				"net":      "1.0",
				"personal": "1.0",
				"shh":      "1.0",
				"txpool":   "1.0",
				"web3":     "1.0",
			},
		}
	},
}

// Todo: implement CanHandle

func (s *ethereumService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	br := bufio.NewReader(conn)

	req, err := http.ReadRequest(br)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	jsonRequest := map[string]interface{}{}
	if err := json.Unmarshal(data, &jsonRequest); err != nil {
		return err
	}

	method := ""

	if s, ok := jsonRequest["method"].(string); ok {
		method = s
	}

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("ethereum"),
		event.Type(method),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.user-agent", req.UserAgent()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
		event.Custom("ethereum.id", jsonRequest["id"]),
		event.Custom("ethereum.method", jsonRequest["method"]),
		event.Custom("ethereum.jsonrpc", jsonRequest["jsonrpc"]),
		event.Payload(data),
		Headers(req.Header),
	))

	buff := bytes.Buffer{}

	fn, ok := ethereumMethods[method]
	if ok {
		v := fn(jsonRequest)

		if err := json.NewEncoder(&buff).Encode(v); err != nil {
			return err
		}
	} else {
		log.Errorf("Method %s not supported", method)
	}

	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Request:    req,
		Header:     http.Header{},
	}

	resp.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp.Header.Add("Content-Length", fmt.Sprintf("%d", buff.Len()))

	resp.Body = ioutil.NopCloser(&buff)

	return resp.Write(conn)
}
