package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// Create a lua scripter instance that handles the connection to all lua-scripts
// A list where all scripts are stored in is generated
func New(options ...func(scripter.Scripter) error) (scripter.Scripter, error) {
	s := &luaScripter{}

	for _, optionFn := range options {
		optionFn(s)
	}

	log.Infof("Using folder: %s", s.Folder)
	s.scripts = map[string]map[string]*lua.LState{} // map[string]*lua.LState{}

	return s, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	Folder string `toml:"folder"`

	scripts map[string]map[string]*lua.LState
}

// Initialize the scripts from a specific service
// The service name is given and the method will loop over all files in the lua-scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) InitScripts(service string) {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", l.Folder, service))
	if err != nil {
		log.Errorf(err.Error())
	}

	//Todo: Load basic lua functions
	l.scripts[service] = map[string]*lua.LState{}

	for _, f := range files {
		ls := lua.NewState()
		ls.DoFile(fmt.Sprintf("%s/%s/%s", l.Folder, service, f.Name()))

		l.scripts[service][f.Name()] = ls
	}
}

// Handle incoming message string
// Get all scripts for a given service and pass the string to each script
func (l *luaScripter) Handle(service string, message string) (string, error) {
	result := message
	var retError error

	for _, v := range l.scripts[service] {
		result, retError = handleScript(*v, result)
	}

	return result, retError
}

// Run the given script on a given message
// Return the value that come out of function(message)
func handleScript(script lua.LState, message string) (string, error) {
	// Call method to handle the message
	if err := script.CallByParam(lua.P{
		Fn:      script.GetGlobal("handle"),
		NRet:    1,
		Protect: true,
	}, lua.LString(message)); err != nil {
		return "", err
	}

	// Get result of the function
	result := script.Get(-1).String()
	script.Pop(1)

	return result, nil
}

// Set a variable that is available in all scripts for a given service
func (l *luaScripter) SetVariable(service string, name string, value string) error {
	for _, v := range l.scripts[service] {
		v.SetGlobal(name, lua.LString(value))
	}
	return nil
}

// Closes the scripter state
func (l *luaScripter) Close() {
	l.Close()
}
