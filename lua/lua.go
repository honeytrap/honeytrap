package lua

import (
	"github.com/yuin/gopher-lua"
	"log"
)

func Handle(message string) (string, error) {
	L := lua.NewState()
	defer L.Close()

	// Load lua file
	if err := L.DoFile("lua-scripts/ssh.lua"); err != nil {
		log.Fatalf("Error loading file: %s", err)
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
