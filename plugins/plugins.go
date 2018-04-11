package plugins

import (
	"fmt"
	"github.com/honeytrap/honeytrap/transforms"
	"os/user"
	"path"
	"plugin"
)

var staticPlugins = make(map[string]transforms.TransformFunc)

// Registers a static plugin.
func Register(name string, fn func() transforms.TransformFunc) int {
	staticPlugins[name] = fn()
	// The return value is unused, but it allows for `var _ = Register("name", handler)`
	return 0
}

// Gets a static or dynamic plugin, giving priority to static ones.
func Get(name string) (transforms.TransformFunc, error) {
	staticPl, ok := staticPlugins[name]
	if ok {
		return staticPl, nil
	}

	/*
		luaPl, ok := readfile(name)
		if ok {
			return lua.New(luaPl), nil
		}
	*/

	// messy, todo: fix/choose path
	// https://stackoverflow.com/a/17617721
	usr, _ := user.Current()
	home := usr.HomeDir
	dynamicPl, err := plugin.Open(path.Join(home, ".honeytrap", name+".so"))
	if err != nil {
		return nil, fmt.Errorf("Couldn't load dynamic plugin: %s", err.Error())
		// return nil, fmt.Errorf("No such plugin")
	}
	sym, err := dynamicPl.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("Couldn't lookup Plugin symbol: %s", err.Error())
	}
	transformConstructor := sym.(func() transforms.TransformFunc)
	return transformConstructor(), nil
}

func MustGet(name string) transforms.TransformFunc {
	out, err := Get(name)
	if err != nil {
		panic(err.Error())
	}
	return out
}
