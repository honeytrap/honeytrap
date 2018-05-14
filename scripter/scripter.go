package scripter

import (
	"github.com/BurntSushi/toml"
	"net"
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

//The scripter interface that implements basic scripter methods
type Scripter interface {
	Init(string) error
	//SetGlobalFn(name string, fn func() string) error
	GetConnection(service string, conn net.Conn) ConnectionWrapper
	Close()
}

//The connectionWrapper interface that implements the basic method that a connection should have
type ConnectionWrapper interface {
	Handle(message string) (string, error)
	SetStringFunction(name string, getString func() string) error
	SetFloatFunction(name string, getFloat func() float64) error
	GetParameter(index int) (string, error)
}

func WithConfig(c toml.Primitive) func(Scripter) error {
	return func(scr Scripter) error {
		return toml.PrimitiveDecode(c, scr)
	}
}
