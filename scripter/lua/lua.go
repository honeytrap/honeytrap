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
	"github.com/honeytrap/honeytrap/pushers"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// New creates a lua scripter instance that handles the connection to all scripts
// A list where all scripts are stored in is generated
func New(name string, options ...scripter.ScripterFunc) (scripter.Scripter, error) {
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

	ab abtester.AbTester

	c pushers.Channel
}

// SetChannel sets the channel over which messages to the log and elasticsearch can be set
func (l *luaScripter) SetChannel(c pushers.Channel) {
	l.c = c
}

// GetChannel gets the channel over which messages to the log and elasticsearch can be set
func (l *luaScripter) GetChannel() pushers.Channel {
	return l.c
}

//Set the abTester from which differential responses can be retrieved
func (l *luaScripter) SetAbTester(ab abtester.AbTester) {
	l.ab = ab
}

// Init initializes the scripts from a specific service
// The service name is given and the method will loop over all files in the scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) Init(service string) error {
	fileNames, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", l.Folder, l.name, service))
	if err != nil {
		return err
	}

	// TODO: Load basic lua functions from shared context
	l.connections = map[string]*luaConn{}
	l.scripts[service] = map[string]string{}
	l.canHandleStates[service] = map[string]*lua.LState{}

	for _, f := range fileNames {
		if f.IsDir() {
			continue
		}

		sf := fmt.Sprintf("%s/%s/%s/%s", l.Folder, l.name, service, f.Name())
		l.scripts[service][f.Name()] = sf

		ls := lua.NewState()
		ls.DoString(fmt.Sprintf("package.path = './%s/lua/?.lua;' .. package.path", l.Folder))
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
		sConn = &luaConn{
			conn: conn,
			scripts: map[string]map[string]*lua.LState{},
			abTester: l.ab,
		}
		l.connections[ip] = sConn
	} else {
		sConn.conn = conn
	}

	if !sConn.HasScripts(service) {
		sConn.AddScripts(service, l.scripts[service], l.Folder)
		scripter.SetBasicMethods(l, sConn, service)
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

// GetScripts return the scripts for this scripter
func (l *luaScripter) GetScripts() map[string]map[string]string {
	return l.scripts
}

// GetScriptFolder return the folder where the scripts are located for this scripter
func (l *luaScripter) GetScriptFolder() string {
	return fmt.Sprintf("%s/%s", l.Folder, l.name)
}

// getConnIP retrieves the IP from a connection's remote address
func getConnIP(conn net.Conn) string {
	s := strings.Split(conn.RemoteAddr().String(), ":")
	s = s[:len(s)-1]
	return strings.Join(s, ":")
}