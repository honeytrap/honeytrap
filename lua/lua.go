package lua

import (
	"github.com/yuin/gopher-lua"
	"fmt"
)

// The lua state to which lua functions are attached
type Lua struct {
	*lua.LState
}

// Default empty lua state for initialization
var Default = Lua{}

// Return a new lua state
func New() *Lua {
	return &Lua{lua.NewState()}
}

func (L *Lua) LoadScripts() error {
	// Load lua file
	if err := L.DoFile("lua-scripts/ssh.lua"); err != nil {
		return fmt.Errorf("error loading file: %s", err)
	}

	return nil
}

// Handle incoming message string
func (L *Lua) Handle(message string) (string, error) {
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

// Closes the lua state
func (L *Lua) Close() {
	L.Close()
}
