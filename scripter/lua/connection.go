package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/abtester"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/yuin/gopher-lua"
	"net"
	"bytes"
)

// Scripter Connection struct
type luaConn struct {
	conn net.Conn

	//List of lua scripts running for this connection: directory/scriptname
	scripts map[string]map[string]*lua.LState

	abTester abtester.AbTester

	connectionBuffer bytes.Buffer
}

//GetConn returns the connection for the srcConn
func (c *luaConn) GetConn() net.Conn {
	return c.conn
}

func (c *luaConn) GetAbTester() abtester.AbTester {
	return c.abTester
}

//SetStringFunction sets a function that is available in all scripts for a service
func (c *luaConn) SetStringFunction(name string, getString func() string, service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			state.Push(lua.LString(getString()))
			return 1
		})
	}

	return nil
}

//SetFloatFunction sets a function that is available in all scripts for a service
func (c *luaConn) SetFloatFunction(name string, getFloat func() float64, service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			state.Push(lua.LNumber(getFloat()))
			return 1
		})
	}

	return nil
}

//SetVoidFunction sets a function that is available in all scripts for a service
func (c *luaConn) SetVoidFunction(name string, doVoid func(), service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			doVoid()
			return 0
		})
	}

	return nil
}

//GetParameters gets the stack parameters from lua to be used in Go functions
func (c *luaConn) GetParameters(params []string, service string) (map[string]string, error) {
	for _, script := range c.scripts[service] {
		if script.GetTop() >= len(params) {
			m := make(map[string]string)
			for index, param := range params {
				m[param] = script.CheckString(script.GetTop() - len(params) + (index + 1))
			}
			return m, nil
		}
	}

	return nil, fmt.Errorf("%s", "Could not find parameters")
}

//HasScripts returns whether the scripts for a given service are loaded already
func (c *luaConn) HasScripts(service string) bool {
	_, ok := c.scripts[service]
	return ok
}

//AddScripts adds scripts to a connection for a given service
func (c *luaConn) AddScripts(service string, scripts map[string]string, folder string) {
	if _, ok := c.scripts[service]; !ok {
		c.scripts[service] = map[string]*lua.LState{}
	}

	for name, script := range scripts {
		ls := lua.NewState()
		ls.DoString(fmt.Sprintf("package.path = './%s/lua/?.lua;' .. package.path", folder))
		if err := ls.DoFile(script); err != nil {
			log.Errorf("Unable to load lua script: %s", err)
			continue
		}
		c.scripts[service][name] = ls
	}
}

// GetConnectionBuffer returns the buffer of the connection
func (c *luaConn) GetConnectionBuffer() *bytes.Buffer {
	return &c.connectionBuffer
}

//Call canHandle Method in Lua state
func callCanHandle(ls *lua.LState, message string) (bool, error) {
	// Call method to check canHandle on the message
	if err := ls.CallByParam(lua.P{
		Fn:      ls.GetGlobal("canHandle"),
		NRet:    1,
		Protect: true,
	}, lua.LString(message)); err != nil {
		return false, fmt.Errorf("error calling canHandle method: %s", err)
	}

	result := ls.ToBool(-1)
	ls.Pop(1)

	return result, nil
}

// Run the given script on a given message
// Return the value that come out of function(message)
func callHandle(ls *lua.LState, message string) (*scripter.Result, error) {
	// Call method to handle the message
	if err := ls.CallByParam(lua.P{
		Fn:      ls.GetGlobal("handle"),
		NRet:    1,
		Protect: true,
	}, lua.LString(message)); err != nil {
		return nil, fmt.Errorf("error calling handle method: %s", err)
	}

	// Get result of the function
	result := &scripter.Result{
		Content: ls.ToString(-1),
	}

	ls.Pop(1)

	return result, nil
}

// Handle calls the handle method on the lua state with the message as the argument
func (c *luaConn) Handle(service string, message string) (*scripter.Result, error) {
	for _, script := range c.scripts[service] {
		canHandle, err := callCanHandle(script, message)
		if err != nil {
			return nil, err
		}
		if !canHandle {
			continue
		}

		result, err := callHandle(script, message)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return nil, nil
}