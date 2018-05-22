package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/abtester"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/op/go-logging"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"net"
	"strings"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// New creates a lua scripter instance that handles the connection to all lua-scripts
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
	l.canHandleStates = map[string]map[string]*lua.LState{}
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
	//Lua states to check whether the connection can be handled with the script
	canHandleStates map[string]map[string]*lua.LState

	abTester abtester.Abtester
}

// Init initializes the scripts from a specific service
// The service name is given and the method will loop over all files in the lua-scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) Init(service string) error {
	fileNames, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", l.Folder, l.name, service))
	if err != nil {
		return err
	}

	// TODO: Load basic lua functions from shared context
	l.scripts[service] = map[string]string{}
	l.canHandleStates[service] = map[string]*lua.LState{}

	for _, f := range fileNames {
		sf := fmt.Sprintf("%s/%s/%s/%s", l.Folder, l.name, service, f.Name())
		l.scripts[service][f.Name()] = sf

		ls := lua.NewState()
		if err := ls.DoFile(sf); err != nil {
			return err
		}
		l.canHandleStates[service][f.Name()] = ls
	}

	return nil
}

//GetConnection returns a connection for the given ip-address, if no connection exists yet, create it.
func (l *luaScripter) GetConnection(service string, conn net.Conn) scripter.ConnectionWrapper {
	ip := getConnIP(conn)

	sConn, ok := l.connections[ip]
	if !ok {
		sConn = &luaConn{conn: conn, scripts: map[string]map[string]*lua.LState{}, abTester: l.abTester}
		l.connections[ip] = sConn
	}

	if !sConn.HasScripts(service) {
		sConn.AddScripts(service, l.scripts[service])
	}

	return &scripter.ConnectionStruct{Service: service, Conn: sConn}
}

// CanHandle checks whether scripter can handle incoming connection for the peeked message
// Returns true if there is one script able to handle the connection
func (l *luaScripter) CanHandle(service string, message string) bool {
	for _, ls := range l.canHandleStates[service] {
		canHandle, err := callCanHandle(ls, message)
		if err != nil {
			log.Errorf("%s", err)
		} else if canHandle {
			return true
		}
	}

	return false
}

func getConnIP(conn net.Conn) string {
	s := strings.Split(conn.RemoteAddr().String(), ":")
	s = s[:len(s)-1]
	return strings.Join(s, ":")
}
