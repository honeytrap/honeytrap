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
package telnet

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/rs/xid"

	lua "github.com/yuin/gopher-lua"
)

var (
	_ = services.Register("telnet-lua", TelnetLua)
)

// TelnetLua is a placeholder
func TelnetLua(options ...services.ServicerFunc) services.Servicer {
	s := &telnetLuaService{
		MOTD:   motd,
		Prompt: prompt,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type telnetLuaService struct {
	c pushers.Channel

	File string `toml:"file"`

	Prompt string `toml:"prompt"`
	MOTD   string `toml:"motd"`
}

func (s *telnetLuaService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *telnetLuaService) Handle(ctx context.Context, conn net.Conn) error {
	id := xid.New()

	defer conn.Close()

	var connOptions event.Option = nil

	if ec, ok := conn.(*event.Conn); ok {
		connOptions = ec.Options()
	}

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("telnet-lua"),
		event.Type("connect"),
		connOptions,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("telnet.sessionid", id.String()),
	))

	term := NewTerminal(conn, s.Prompt)

	term.Write([]byte(s.MOTD + "\n"))

	term.SetPrompt("Username: ")
	username, err := term.ReadLine()
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	password, err := term.ReadPassword("Password: ")
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("telnet-lua"),
		event.Type("password-authentication"),
		connOptions,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("telnet.sessionid", id.String()),
		event.Custom("telnet.username", username),
		event.Custom("telnet.password", password),
	))

	term.SetPrompt(s.Prompt)

	L := lua.NewState()

	mt := L.NewTypeMetatable("terminal")
	L.SetGlobal("terminal", mt)

	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(),
		map[string]lua.LGFunction{
			"log_event": func(L *lua.LState) int {
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
					opts = append(opts, event.Custom(fmt.Sprintf("telnet.%s", k), v))
				}

				s.c.Send(event.New(
					services.EventOptions,
					event.Category("telnet-lua"),
					event.Type("session"),
					connOptions,
					event.SourceAddr(term.RemoteAddr()),
					event.DestinationAddr(term.LocalAddr()),
					event.Custom("telnet.sessionid", id),
					event.NewWith(opts...),
				))

				return 0
			},
			"read_line": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				term, ok := ud.Value.(*Terminal)
				if !ok {
					L.ArgError(1, "terminal expected")
					return 1
				}

				line, err := term.ReadLine()
				if err == io.EOF {
					return 0
				} else if err != nil {
					return 0
				}

				var connOptions event.Option = nil

				if ec, ok := term.Conn.(*event.Conn); ok {
					connOptions = ec.Options()
				}

				s.c.Send(event.New(
					services.EventOptions,
					event.Category("telnet-lua"),
					event.Type("session"),
					connOptions,
					event.SourceAddr(term.RemoteAddr()),
					event.DestinationAddr(term.LocalAddr()),
					event.Custom("telnet.sessionid", id),
					event.Custom("telnet.command", line),
				))

				L.Push(lua.LString(line))
				return 1
			},
			"write": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				term, ok := ud.Value.(*Terminal)
				if !ok {
					L.ArgError(1, "terminal expected")
					return 0
				}

				if L.GetTop() != 2 {
					L.ArgError(1, "string expected")
					return 0
				}

				log.Info(L.Get(2).String())
				term.Write([]byte(L.Get(2).String()))
				return 0
			},
			"write_line": func(L *lua.LState) int {
				ud := L.CheckUserData(1)

				term, ok := ud.Value.(*Terminal)
				if !ok {
					L.ArgError(1, "terminal expected")
					return 0
				}

				if L.GetTop() != 2 {
					L.ArgError(1, "string expected")
					return 0
				}

				log.Info(L.Get(2).String())
				term.Write([]byte(L.Get(2).String()))
				term.Write([]byte("\n"))
				return 0
			},
		}))

	if err := L.DoFile(s.File); err != nil {
		return err
	}

	ud := L.NewUserData()
	ud.Value = term
	L.SetMetatable(ud, L.GetTypeMetatable("terminal"))

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

	_ = ret

	return nil
}
