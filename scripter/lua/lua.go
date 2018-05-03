package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/yuin/gopher-lua"
	"log"
)

var (
	_ = scripter.Register("lua", New)
)

func New(options ...func(scripter.Scripter) error) (scripter.Scripter, error) {
	s := &luaScripter{}

	for _, optionFn := range options {
		optionFn(s)
	}

	log.Printf("Using folder: %s", s.Folder)

	return s, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	*lua.LState
	Folder string `toml:"folder"`
}

func (L *luaScripter) LoadScripts(script string) error {
	// Load scripter file
	if err := L.DoFile(script); err != nil {
		return fmt.Errorf("error loading file: %s", err)
	}

	return nil
}

// Handle incoming message string
func (L *luaScripter) Handle(message string) (string, error) {
	// If scripter is not initialized, return default string
	if L == nil {
		return message, nil
	}

	// Call method to handle the message
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("handle"),
		NRet:    1,
		Protect: true,
	}, lua.LString(message)); err != nil {
		return "", err
	}

	// Get result of the function
	result := L.Get(-1).String()
	L.Pop(1)

	return result, nil
}

// Closes the scripter state
func (L *luaScripter) Close() {
	L.Close()
}
