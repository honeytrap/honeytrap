package scripter

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/abtester"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"net"
)

var (
	scripters = map[string]func(string, ...ScripterFunc) (Scripter, error){}
)
var log = logging.MustGetLogger("scripter")

//Register the scripter instance
func Register(key string, fn func(string, ...ScripterFunc) (Scripter, error)) func(string, ...ScripterFunc) (Scripter, error) {
	scripters[key] = fn
	return fn
}

type ScripterFunc func(Scripter) error

//Get a scripter instance
func Get(key string) (func(string, ...ScripterFunc) (Scripter, error), bool) {
	if fn, ok := scripters[key]; ok {
		return fn, true
	}

	return nil, false
}

//GetAvailableScripterNames gets all scripters that are registered
func GetAvailableScripterNames() []string {
	var out []string
	for key := range scripters {
		out = append(out, key)
	}
	return out
}

func WithChannel(eb pushers.Channel) ScripterFunc {
	return func(s Scripter) error {
		s.SetChannel(eb)
		return nil
	}
}

//Scripter interface that implements basic scripter methods
type Scripter interface {
	Init(string) error
	GetConnection(service string, conn net.Conn) ConnectionWrapper
	CanHandle(service string, message string) bool
	SetChannel(c pushers.Channel)
	GetChannel() pushers.Channel
}

//ConnectionWrapper interface that implements the basic method that a connection should have
type ConnectionWrapper interface {
	GetScrConn() ScrConn
	Handle(message string) (string, error)
	SetStringFunction(name string, getString func() string) error
	SetFloatFunction(name string, getFloat func() float64) error
	SetVoidFunction(name string, doVoid func()) error
	GetParameters(params []string) (map[string]string, error)
}

//ScrConn wraps a connection and exposes methods to interact with the connection and scripter
type ScrConn interface {
	GetConn() net.Conn
	SetStringFunction(name string, getString func() string, service string) error
	SetFloatFunction(name string, getFloat func() float64, service string) error
	SetVoidFunction(name string, doVoid func(), service string) error
	GetParameters(params []string, service string) (map[string]string, error)
	HasScripts(service string) bool
	AddScripts(service string, scripts map[string]string)
	Handle(service string, message string) (*Result, error)
	GetConnectionBuffer() *bytes.Buffer
}

//Result struct which allows the result to be a string, an empty string and a nil value
//The nil value can be used to indicate that lua has no value to return
type Result struct {
	Content string
}

//ScrAbTester exposes methods to interact with the AbTester
type ScrAbTester interface {
	GetAbTester() abtester.Abtester
}

//WithConfig returns a function to attach the config to the scripter
func WithConfig(c toml.Primitive) ScripterFunc {
	return func(scr Scripter) error {
		return toml.PrimitiveDecode(c, scr)
	}
}
