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
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/rs/xid"

	"io/ioutil"

	"encoding/json"

	"strings"

	"bytes"

	lua "github.com/yuin/gopher-lua"
)

var (
	_ = Register("http-lua", HTTPLua)
)

// Http is a placeholder
func HTTPLua(options ...ServicerFunc) Servicer {
	s := &httpLuaService{
		httpLuaServiceConfig: httpLuaServiceConfig{},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type httpLuaServiceConfig struct {
	File string `toml:"file"`
}

type httpLuaService struct {
	httpLuaServiceConfig

	c pushers.Channel
}

func (s *httpLuaService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *httpLuaService) Handle(ctx context.Context, conn net.Conn) error {
	L := lua.NewState()

	mtreq := L.NewTypeMetatable("http_request")
	L.SetGlobal("http_request", mtreq)

	L.SetField(mtreq, "__index", L.SetFuncs(L.NewTable(),
		map[string]lua.LGFunction{
			"log_event": func(L *lua.LState) int {
				/*
					ud := L.CheckUserData(1)

					term, ok := ud.Value.(*Terminal)
					if !ok {
						L.ArgError(1, "terminal expected")
						return 1
					}

					var connOptions event.Option = nil

					if ec, ok := term.Conn.(*event.Conn); ok {
						connOptions = ec.Options()
					}

					opts := []event.Option{}

					if L.GetTop() != 2 {
						L.ArgError(1, "table expected")
						return 0
					}

					params, ok := FromLUA(L.Get(2)).(map[string]interface{})
					if !ok {
						log.Errorf("Unexpected type: %#+v", FromLUA(L.Get(2)))
						return 0
					}

					for k, v := range params {
						opts = append(opts, event.Custom(fmt.Sprintf("http.%s", k), v))
					}

					s.c.Send(event.New(
						services.EventOptions,
						event.Category("http-lua"),
						event.Type("session"),
						connOptions,
						event.SourceAddr(term.RemoteAddr()),
						event.DestinationAddr(term.LocalAddr()),
						event.Custom("http.sessionid", id),
						event.NewWith(opts...),
					))
				*/

				return 0
			},
			"method": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				req, ok := ud.Value.(*http.Request)
				if !ok {
					L.ArgError(1, "http_request expected")
					return 1
				}

				L.Push(lua.LString(string(req.Method)))
				return 1
			},
			"url": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				req, ok := ud.Value.(*http.Request)
				if !ok {
					L.ArgError(1, "http_request expected")
					return 1
				}

				L.Push(ToLUA(L, map[string]interface{}{
					"path":  req.URL.Path,
					"host":  req.URL.Host,
					"query": req.URL.RawQuery,
				}))
				return 1
			},
			"body": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				req, ok := ud.Value.(*http.Request)
				if !ok {
					L.ArgError(1, "http_request expected")
					return 1
				}

				defer req.Body.Close()

				body, err := ioutil.ReadAll(req.Body)
				if err == io.EOF {
					return 0
				} else if err != nil {
					return 0
				}

				fmt.Println(string(body))
				L.Push(lua.LString(string(body)))
				return 1
			},
			"body_json": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				req, ok := ud.Value.(*http.Request)
				if !ok {
					L.ArgError(1, "http_request expected")
					return 1
				}

				defer req.Body.Close()

				var v interface{}

				if err := json.NewDecoder(req.Body).Decode(&v); err == io.EOF {
					return 0
				} else if err != nil {
					log.Error("Could not parse json body: %s", err.Error())
					return 0
				}

				L.Push(ToLUA(L, v))
				return 1
			},
		}))

	mtresp := L.NewTypeMetatable("http_response")
	L.SetGlobal("http_response", mtresp)

	L.SetField(mtresp, "new", L.NewFunction(func(L *lua.LState) int {
		ud := L.CheckUserData(1)

		req, ok := ud.Value.(*http.Request)
		if !ok {
			L.ArgError(1, "http_request expected")
			return 1
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

		ret := L.NewUserData()
		ret.Value = &resp
		L.SetMetatable(ret, L.GetTypeMetatable("http_response"))
		L.Push(ret)
		return 1
	}))

	L.SetField(mtresp, "__index", L.SetFuncs(L.NewTable(),
		map[string]lua.LGFunction{
			"log_event": func(L *lua.LState) int {
				/*
					ud := L.CheckUserData(1)

					term, ok := ud.Value.(*Terminal)
					if !ok {
						L.ArgError(1, "terminal expected")
						return 1
					}

					var connOptions event.Option = nil

					if ec, ok := term.Conn.(*event.Conn); ok {
						connOptions = ec.Options()
					}

					opts := []event.Option{}

					if L.GetTop() != 2 {
						L.ArgError(1, "table expected")
						return 0
					}

					params, ok := FromLUA(L.Get(2)).(map[string]interface{})
					if !ok {
						log.Errorf("Unexpected type: %#+v", FromLUA(L.Get(2)))
						return 0
					}

					for k, v := range params {
						opts = append(opts, event.Custom(fmt.Sprintf("http.%s", k), v))
					}

					s.c.Send(event.New(
						services.EventOptions,
						event.Category("http-lua"),
						event.Type("session"),
						connOptions,
						event.SourceAddr(term.RemoteAddr()),
						event.DestinationAddr(term.LocalAddr()),
						event.Custom("http.sessionid", id),
						event.NewWith(opts...),
					))
				*/

				return 0
			},
			"set_statuscode": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				resp, ok := ud.Value.(*http.Response)
				if !ok {
					L.ArgError(1, "http_response expected")
					return 1
				}

				if L.GetTop() != 2 {
					L.ArgError(1, "status_code expected")
					return 0
				}

				code := L.CheckInt(2)

				resp.StatusCode = code
				resp.Status = http.StatusText(code)
				return 0
			},
			"set_header": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				resp, ok := ud.Value.(*http.Response)
				if !ok {
					L.ArgError(1, "http_response expected")
					return 1
				}

				if L.GetTop() != 3 {
					L.ArgError(1, "body expected")
					return 0
				}

				k := L.CheckString(2)
				v := L.CheckString(3)

				resp.Header.Set(k, v)
				return 0
			},
			"body": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				resp, ok := ud.Value.(*http.Response)
				if !ok {
					L.ArgError(1, "http_response expected")
					return 1
				}

				if L.GetTop() != 2 {
					L.ArgError(1, "body expected")
					return 0
				}

				body := L.CheckString(2)

				log.Debug(body)
				resp.Body = ioutil.NopCloser(strings.NewReader(body))
				resp.ContentLength = int64(len(body))
				return 0
			},
			/*
				"write_headers": func(L *lua.LState) int {
					ud := L.CheckUserData(1)

					resp, ok := ud.Value.(*http.Response)
					if !ok {
						L.ArgError(1, "http_response expected")
						return 1
					}

					if L.GetTop() != 2 {
						L.ArgError(1, "content expected")
						return 0
					}

					resp.TransferEncoding="chunked"
					return 0
				},
				"write_line": func(L *lua.LState) int {
					ud := L.CheckUserData(1)

					resp, ok := ud.Value.(*http.Response)
					if !ok {
						L.ArgError(1, "http_response expected")
						return 1
					}

					if L.GetTop() != 2 {
						L.ArgError(1, "content expected")
						return 0
					}

					content := L.CheckString(2)
					resp.
						fmt.Fprintln(resp.Body, content)
					return 0
				},
				"write": func(L *lua.LState) int {
					ud := L.CheckUserData(1)

					resp, ok := ud.Value.(*http.Response)
					if !ok {
						L.ArgError(1, "http_response expected")
						return 1
					}

					if L.GetTop() != 2 {
						L.ArgError(1, "content expected")
						return 0
					}

					content := L.CheckString(2)
					fmt.Fprint(resp.Body, content)
					return 0
				},
			*/
		}))

	if err := L.DoFile(s.File); err != nil {
		return err
	}

	id := xid.New()
	_ = id

	var connOptions event.Option = nil

	if ec, ok := conn.(*event.Conn); ok {
		connOptions = ec.Options()
	}

	br := bufio.NewReader(conn)

	for {
		req, err := http.ReadRequest(br)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(req.Body)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(body))

		s.c.Send(event.New(
			EventOptions,
			connOptions,
			event.Category("http"),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("http.sessionid", id.String()),
			event.Custom("http.method", req.Method),
			event.Custom("http.proto", req.Proto),
			event.Custom("http.host", req.Host),
			event.Custom("http.url", req.URL.String()),
			event.Payload(body),
			Headers(req.Header),
			Cookies(req.Cookies()),
		))

		ud := L.NewUserData()
		ud.Value = req
		L.SetMetatable(ud, L.GetTypeMetatable("http_request"))

		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("handle"),
			NRet:    1,
			Protect: true,
		}, ud); err != nil {
			log.Error("Error calling lua method: %s", err.Error())
			return err
		}

		ret := L.Get(-1)
		L.Pop(1)

		if ret == nil {
			// return 500
		}

		udret := ret.(*lua.LUserData)

		resp, ok := udret.Value.(*http.Response)
		if !ok {
			return fmt.Errorf("http_response expected")
		}

		if err := resp.Write(conn); err != nil {
			return err
		}
	}
}
