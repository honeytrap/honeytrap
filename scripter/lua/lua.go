package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"github.com/op/go-logging"
	"sync"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

func New(options ...func(scripter.Scripter) error) (scripter.Scripter, error) {
	s := &luaScripter{}

	for _, optionFn := range options {
		optionFn(s)
	}

	log.Infof("Using folder: %s", s.Folder)
	s.scripts = &sync.Map{} // map[string]*lua.LState{}

	return s, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	Folder string `toml:"folder"`

	scripts *sync.Map
}

func (l *luaScripter) InitScripts(service string) {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", l.Folder, service))
	if err != nil {
		log.Errorf(err.Error())
	}

	ls := lua.NewState()
	//Todo: Load basic lua functions

	for _, f := range files {
		ls.DoFile(fmt.Sprintf("%s/%s/%s", l.Folder, service, f.Name()))
	}

	l.scripts.Store(service, ls)
}

// Handle incoming message string
func (l *luaScripter) Handle(service string, message string) (string, error) {
	ls, err := l.loadScript(service)
	if err != nil {
		return message, err
	}

	// Call method to handle the message
	if err := ls.CallByParam(lua.P{
		Fn:      ls.GetGlobal("handle"),
		NRet:    1,
		Protect: true,
	}, lua.LString(message)); err != nil {
		return "", err
	}

	// Get result of the function
	result := ls.Get(-1).String()
	ls.Pop(1)

	return result, nil
}

func (l *luaScripter) SetVariable(service string, name string, value string) error {
	ls, err := l.loadScript(service)
	if err != nil {
		return err
	}

	ls.SetGlobal(name, lua.LString(value))

	return nil
}

func (l *luaScripter) loadScript(service string) (*lua.LState, error) {
	lState, ok := l.scripts.Load(service)
	if !ok {
		return nil, fmt.Errorf("could not retrieve lua state for service %s", service)
	}
	return lState.(*lua.LState), nil
}

// Closes the scripter state
func (l *luaScripter) Close() {
	l.Close()
}
