package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/op/go-logging"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"net"
	"strings"
	"github.com/honeytrap/honeytrap/abtester"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// Create a lua scripter instance that handles the connection to all lua-scripts
// A list where all scripts are stored in is generated
func New(name string, options ...func(scripter.Scripter) error) (scripter.Scripter, error) {
	l := &luaScripter{
		name: name,
	}

	for _, optionFn := range options {
		optionFn(l)
	}

	log.Infof("Using folder: %s", l.Folder)
	l.scripts = map[string]map[string]string{}
	l.connections = map[string]*luaConn{}
	l.abTester, _ = abtester.Namespace("lua")

	if err := l.abTester.LoadFromFile("scripter/abtests.json"); err != nil {
		return nil, err
	}

	return l, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	name string

	Folder string `toml:"folder"`

	//Source of the states, initialized per connection: directory/scriptname
	scripts map[string]map[string]string
	//List of connections keyed by 'ip'
	connections map[string]*luaConn

	abTester abtester.Abtester
}

// Initialize the scripts from a specific service
// The service name is given and the method will loop over all files in the lua-scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) Init(service string) error {
	fileNames, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", l.Folder, l.name, service))
	if err != nil {
		return err
	}

	// TODO: Load basic lua functions from shared context
	l.scripts[service] = map[string]string{}

	for _, f := range fileNames {
		l.scripts[service][f.Name()] = fmt.Sprintf("%s/%s/%s/%s", l.Folder, l.name, service, f.Name())
	}

	return nil
}

// Closes the scripter state
func (l *luaScripter) Close() {
	l.Close()
}

//Return a connection for the given ip-address, if no connection exists yet, create it.
func (l *luaScripter) GetConnection(service string, conn net.Conn) scripter.ConnectionWrapper {
	s := strings.Split(conn.RemoteAddr().String(), ":")
	s = s[:len(s)-1]
	ip := strings.Join(s, ":")
	var sConn *luaConn
	var ok bool

	if sConn, ok = l.connections[ip]; !ok {
		sConn = &luaConn{}
		sConn.conn = conn
		sConn.scripts = map[string]map[string]*lua.LState{}
		sConn.abTester = l.abTester
		l.connections[ip] = sConn
	}

	if !sConn.HasScripts(service) {
		sConn.AddScripts(service, l.scripts[service])
	}

	return &scripter.ConnectionStruct{Service: service, MyConn: sConn}
}

