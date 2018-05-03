package scripter

import (
	"github.com/BurntSushi/toml"
)

var (
	scripters = map[string]func(...func(Scripter) error) (Scripter, error) {}
)

func Register(key string, fn func(...func(Scripter) error) (Scripter, error)) func(...func(Scripter) error) (Scripter, error) {
	scripters[key] = fn
	return fn
}

func Get(key string) (func(...func(Scripter) error) (Scripter, error), bool) {
	if fn, ok := scripters[key]; ok {
		return fn, true
	}

	return nil, false
}

func GetAvailableScripterNames() []string {
	var out []string
	for key := range scripters {
		out = append(out, key)
	}
	return out
}

type Scripter interface {
	InitScripts(service string)
	Handle(service string, message string) (string, error)
	SetVariable(service string, name string, value string) error
}


func WithConfig(c toml.Primitive) func(Scripter) error {
	return func(scr Scripter) error {
		return toml.PrimitiveDecode(c, scr)
	}
}
