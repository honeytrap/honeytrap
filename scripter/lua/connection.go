package lua

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"net"
	"github.com/honeytrap/honeytrap/abtester"
	"errors"
	"github.com/honeytrap/honeytrap/scripter"
)

// Scripter Connection struct
type luaConn struct {
	conn net.Conn

	//List of lua scripts running for this connection: directory/scriptname
	scripts map[string]map[string]*lua.LState

	abTester abtester.Abtester
}

func (c *luaConn) GetConn() net.Conn {
	return c.conn
}

func (c *luaConn) GetAbTester() abtester.Abtester {
	return c.abTester
}

// Set a function that is available in all scripts for a service
func (c *luaConn) SetStringFunction(name string, getString func() string, service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			state.Push(lua.LString(getString()))
			return 1
		})
	}

	return nil
}

// Set a function that is available in all scripts for a service
func (c *luaConn) SetFloatFunction(name string, getFloat func() float64, service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			state.Push(lua.LNumber(getFloat()))
			return 1
		})
	}

	return nil
}

// Set a function that is available in all scripts for a service
func (c *luaConn) SetVoidFunction(name string, doVoid func(), service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			doVoid()
			return 0
		})
	}

	return nil
}

// Get the stack parameter from lua to be used in Go functions
func (c *luaConn) GetParameter(index int, service string) (string, error) {
	for _, script := range c.scripts[service] {
		if script.GetTop() >= 1 {
			if parameter := script.CheckString(script.GetTop() + index); parameter != "" {
				return parameter, nil
			}
		}
	}

	return "", fmt.Errorf("%s", "Could not find parameter")
}

//Returns if the scripts for a given service are loaded already
func (c *luaConn) HasScripts(service string) bool {
	_, ok := c.scripts[service]
	return ok
}

//Add scripts to a connection for a given service
func (c *luaConn) AddScripts(service string, scripts map[string]string) {
	_, ok := c.scripts[service]; if !ok {
		c.scripts[service] = map[string]*lua.LState{}
	}

	for name, script := range scripts {
		ls := lua.NewState()
		if err := ls.DoFile(script); err != nil {
			log.Errorf("Unable to load lua script: %s", err)
			continue
		}
		c.scripts[service][name] = ls
	}

	scripter.SetBasicMethods(c, service)
}

// Run the given script on a given message
// Return the value that come out of function(message)
func handleScript(ls *lua.LState, message string) (string, error) {
	// Call method to handle the message
	if err := ls.CallByParam(lua.P{
		Fn:      ls.GetGlobal("handle"),
		NRet:    1,
		Protect: true,
	}, lua.LString(message)); err != nil {
		return "", errors.New(fmt.Sprintf("error calling handle method:%s", err))
	}

	// Get result of the function
	result := ls.Get(-1).String()
	ls.Pop(1)

	return result, nil
}

func (c *luaConn) HandleScripts(service string, message string) (string, error) {
	var result string
	var err error

	for _, script := range c.scripts[service] {
		result, err = handleScript(script, result)
		if err != nil {
			return "", err
		}
	}

	return result, err
}