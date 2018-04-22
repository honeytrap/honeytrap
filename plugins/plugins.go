package plugins

import (
	"fmt"
	"os"
	"path"
	"plugin"
)

func Get(name, symName, folder string) (sym interface{}, found bool, e error) {
	filename := path.Join(folder, name+".so")
	if _, err := os.Stat(filename); err != nil {
		e = fmt.Errorf("Couldn't find dynamic plugin: %s", err.Error())
		return
	}
	found = true
	dynamicPl, err := plugin.Open(filename)
	if err != nil {
		e = fmt.Errorf("Couldn't load dynamic plugin: %s", err.Error())
		return
	}
	sym, err = dynamicPl.Lookup(symName)
	if err != nil {
		e = fmt.Errorf("Couldn't lookup symbol \"%s\": %s", symName, err.Error())
		return
	}
	return
}

func MustGet(name, symName, folder string) interface{} {
	out, found, err := Get(name, symName, folder)
	if !found {
		panic(fmt.Errorf("Plugin %s not found", name))
	}
	if err != nil {
		panic(err.Error())
	}
	return out
}
