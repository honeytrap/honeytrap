package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/op/go-logging"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"errors"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// Create a lua scripter instance that handles the connection to all lua-scripts
// A list where all scripts are stored in is generated
func New(name string, options ...func(scripter.Scripter) error) (scripter.Scripter, error) {
	s := &luaScripter{
		name: name,
	}

	for _, optionFn := range options {
		optionFn(s)
	}

	log.Infof("Using folder: %s", s.Folder)
	s.scripts = map[string]map[string]*lua.LState{}

	return s, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	name string

	Folder string `toml:"folder"`

	scripts map[string]map[string]*lua.LState
}

// Initialize the scripts from a specific service
// The service name is given and the method will loop over all files in the lua-scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) InitScripts(service string) error {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", l.Folder, l.name, service))
	if err != nil {
		return err
	}

	// TODO: Load basic lua functions from shared context
	l.scripts[l.name] = map[string]*lua.LState {}

	for _, f := range files {
		ls := lua.NewState()
		ls.DoFile(fmt.Sprintf("%s/%s/%s/%s", l.Folder, l.name, service, f.Name()))
		if err != nil {
			return err
		}

		l.scripts[l.name][f.Name()] = ls
	}

	return nil
}

func (l *luaScripter) SetGlobalFn(name string, fn func() string) error {
	return l.SetStringFunction(name, fn)
}

// Handle incoming message string
// Get all scripts for a given service and pass the string to each script
func (l *luaScripter) Handle(message string) (string, error) {
	result := message
	var err error

	// TODO: Figure out the correct way to call all handle methods
	for _, ls := range l.scripts[l.name] {
		result, err = handleScript(ls, result)
		if err != nil {
			return "", err
		}
	}

	return result, nil
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

// Set a function that is available in all scripts for a service
func (l *luaScripter) SetStringFunction(name string, getString func() string) error {
	for _, ls := range l.scripts[l.name] {
		ls.Register(name, func(state *lua.LState) int {
			state.Push(lua.LString(getString()))
			return 1
		})
	}

	return nil
}

// Closes the scripter state
func (l *luaScripter) Close() {
	l.Close()
}
