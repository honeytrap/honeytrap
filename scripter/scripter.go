package scripter

import (
	"github.com/BurntSushi/toml"
)

var (
	scripters = map[string]func(string, ...func(Scripter) error) (Scripter, error){}
)

func Register(key string, fn func(string, ...func(Scripter) error) (Scripter, error)) func(string, ...func(Scripter) error) (Scripter, error) {
	scripters[key] = fn
	return fn
}

func Get(key string) (func(string, ...func(Scripter) error) (Scripter, error), bool) {
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
	InitScripts(string) error
	Handle(message string) (string, error)
	SetGlobalFn(name string, fn func() string) error
	SetStringFunction(name string, getString func() string) error
}

func WithConfig(c toml.Primitive) func(Scripter) error {
	return func(scr Scripter) error {
		return toml.PrimitiveDecode(c, scr)
	}
}
