package scripter

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/abtester"
	"github.com/honeytrap/honeytrap/utils/files"
	"github.com/op/go-logging"
	"net"
	"time"
)

var (
	scripters = map[string]func(string, ...func(Scripter) error) (Scripter, error){}
)
var log = logging.MustGetLogger("scripter")

//Register the scripter instance
func Register(key string, fn func(string, ...func(Scripter) error) (Scripter, error)) func(string, ...func(Scripter) error) (Scripter, error) {
	scripters[key] = fn
	return fn
}

//Get a scripter instance
func Get(key string) (func(string, ...func(Scripter) error) (Scripter, error), bool) {
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

//Scripter interface that implements basic scripter methods
type Scripter interface {
	Init(string) error
	//SetGlobalFn(name string, fn func() string) error
	GetConnection(service string, conn net.Conn) ConnectionWrapper
	CanHandle(service string, message string) bool
}

//ConnectionWrapper interface that implements the basic method that a connection should have
type ConnectionWrapper interface {
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
func WithConfig(c toml.Primitive) func(Scripter) error {
	return func(scr Scripter) error {
		return toml.PrimitiveDecode(c, scr)
	}
}

//SetBasicMethods sets methods that can be called by each script, returning basic functionality
func SetBasicMethods(c ScrConn, service string) {
	c.SetStringFunction("getRemoteAddr", func() string { return c.GetConn().RemoteAddr().String() }, service)
	c.SetStringFunction("getLocalAddr", func() string { return c.GetConn().LocalAddr().String() }, service)

	c.SetStringFunction("getDatetime", func() string {
		t := time.Now()
		return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
	}, service)

	c.SetStringFunction("getFileDownload", func() string {
		params, _ := c.GetParameters([]string{"url", "path"}, service)

		if err := files.Download(params["url"], params["path"]); err != nil {
			log.Errorf("error downloading file: %s", err)
			return "no"
		}
		return "yes"
	}, service)

	if ab, ok := c.(ScrAbTester); ok {
		//In the script the function 'getAbTest(key)' can be called, returning a random result for the given key
		c.SetStringFunction("getAbTest", func() string {
			params, _ := c.GetParameters([]string{"key"}, service)

			val, err := ab.GetAbTester().GetForGroup(service, params["key"], -1)
			if err != nil {
				return "_" //No response, _ so lua knows it has no ab-test
			}

			return val
		}, service)
	}

	//In the script the function 'doLog(type, message)' can be called, with type = logging type and message the message
	c.SetVoidFunction("doLog", func() {
		params, _ := c.GetParameters([]string{"logType", "message"}, service)
		logType := params["logType"]
		message := params["message"]

		if logType == "critical" {
			log.Critical(message)
		}
		if logType == "debug" {
			log.Debug(message)
		}
		if logType == "error" {
			log.Error(message)
		}
		if logType == "fatal" {
			log.Fatal(message)
		}
		if logType == "info" {
			log.Info(message)
		}
		if logType == "notice" {
			log.Notice(message)
		}
		if logType == "panic" {
			log.Panic(message)
		}
		if logType == "warning" {
			log.Warning(message)
		}
	}, service)
}
