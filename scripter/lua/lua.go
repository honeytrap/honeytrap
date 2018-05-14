package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/op/go-logging"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"errors"
	"net"
	"strings"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// Create a lua scripter instance that handles the connection to all lua-scripts
// A list where all scripts are stored in is generated
func New(name string, options ...func(scripter.Scripter) error) (scripter.Scripter, error) {
	l := &luaScripter{
		name: name,
	}

	for _, optionFn := range options {
		optionFn(l)
	}

	log.Infof("Using folder: %s", l.Folder)
	l.scripts = map[string]map[string]string{}
	l.connections = map[string]scripterConn{}

	return l, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	name string

	Folder string `toml:"folder"`

	//Source of the states, initialized per connection: directory/scriptname
	scripts map[string]map[string]string
	//List of connections keyed by 'ip'
	connections map[string]scripterConn
}

// Initialize the scripts from a specific service
// The service name is given and the method will loop over all files in the lua-scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) Init(service string) error {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", l.Folder, l.name, service))
	if err != nil {
		return err
	}

	// TODO: Load basic lua functions from shared context
	l.scripts[service] = map[string]string{}

	for _, f := range files {
		l.scripts[service][f.Name()] = fmt.Sprintf("%s/%s/%s/%s", l.Folder, l.name, service, f.Name())
	}

	return nil
}

// Closes the scripter state
func (l *luaScripter) Close() {
	l.Close()
}

//Return a connection for the given ip-address, if no connection exists yet, create it.
func (l *luaScripter) GetConnection(service string, conn net.Conn) scripter.ConnectionWrapper {
	s := strings.Split(conn.RemoteAddr().String(), ":")
	s = s[:len(s)-1]
	ip := strings.Join(s, ":")
	var sConn scripterConn
	var ok bool

	if sConn, ok = l.connections[ip]; !ok {
		sConn = scripterConn{}
		sConn.conn = conn
		sConn.scripts = map[string]map[string]*lua.LState{}
		l.connections[ip] = sConn
	}

	if !sConn.hasScripts(service) {
		sConn.addScripts(service, l.scripts[service])
	}

	return &ConnectionStruct{service, sConn}
}

//func (l *luaScripter) SetGlobalFn(name string, fn func() string) error {
//	//for _, script := range l.scripts {
//	//	return l.SetStringFunction(name, fn)
//	//}
//}

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



// Connection Wrapper struct
type ConnectionStruct struct {
	service string
	conn scripterConn
}

// Handle incoming message string
// Get all scripts for a given service and pass the string to each script
func (w *ConnectionStruct) Handle(message string) (string, error) {
	result := message
	var err error

	// TODO: Figure out the correct way to call all handle methods
	for _, script := range w.conn.scripts[w.service] {
		result, err = handleScript(script, result)
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

//Set a string function for a connection
func (w *ConnectionStruct) SetStringFunction(name string, getString func() string) error {
	return w.conn.SetStringFunction(name, getString, w.service)
}

//Set a string function for a connection
func (w *ConnectionStruct) SetFloatFunction(name string, getFloat func() float64) error {
	return w.conn.SetFloatFunction(name, getFloat, w.service)
}

//Get a parameter from a connection
func (w *ConnectionStruct) GetParameter(index int) (string, error) {
	return w.conn.GetParameter(index, w.service)
}



// Scripter Connection struct
type scripterConn struct {
	conn net.Conn

	//List of lua scripts running for this connection: directory/scriptname
	scripts map[string]map[string]*lua.LState
}

// Set a function that is available in all scripts for a service
func (c *scripterConn) SetStringFunction(name string, getString func() string, service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			state.Push(lua.LString(getString()))
			return 1
		})
	}

	return nil
}

// Set a function that is available in all scripts for a service
func (c *scripterConn) SetFloatFunction(name string, getFloat func() float64, service string) error {
	for _, script := range c.scripts[service] {
		script.Register(name, func(state *lua.LState) int {
			state.Push(lua.LNumber(getFloat()))
			return 1
		})
	}

	return nil
}

// Get the stack parameter from lua to be used in Go functions
func (c *scripterConn) GetParameter(index int, service string) (string, error) {
	for _, script := range c.scripts[service] {
		if script.GetTop() >= 2 {
			if parameter := script.CheckString(script.GetTop() + index); parameter != "" {
				return parameter, nil
			}
		}
	}

	return "", fmt.Errorf("%s", "Could not find parameter")
}

//Returns if the scripts for a given service are loaded already
func (c *scripterConn) hasScripts(service string) bool {
	_, ok := c.scripts[service]
	return ok
}

//Add scripts to a connection for a given service
func (c *scripterConn) addScripts(service string, scripts map[string]string) {
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
}