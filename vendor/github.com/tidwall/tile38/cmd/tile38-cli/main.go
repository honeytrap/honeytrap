package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/peterh/liner"
	"github.com/tidwall/resp"
	"github.com/tidwall/tile38/client"
	"github.com/tidwall/tile38/core"
)

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

var (
	historyFile = filepath.Join(userHomeDir(), ".liner_example_history")
)

type connError struct {
	OK  bool   `json:"ok"`
	Err string `json:"err"`
}

var (
	hostname   = "127.0.0.1"
	output     = "json"
	port       = 9851
	oneCommand string
	tokml      bool
	raw        bool
	noprompt   bool
	tty        bool
)

func showHelp() bool {

	gitsha := ""
	if core.GitSHA == "" || core.GitSHA == "0000000" {
		gitsha = ""
	} else {
		gitsha = " (git:" + core.GitSHA + ")"
	}
	fmt.Fprintf(os.Stdout, "tile38-cli %s%s\n\n", core.Version, gitsha)
	fmt.Fprintf(os.Stdout, "Usage: tile38-cli [OPTIONS] [cmd [arg [arg ...]]]\n")
	fmt.Fprintf(os.Stdout, " --raw              Use raw formatting for replies (default when STDOUT is not a tty)\n")
	fmt.Fprintf(os.Stdout, " --noprompt         Do not display a prompt\n")
	fmt.Fprintf(os.Stdout, " --tty              Force TTY\n")
	fmt.Fprintf(os.Stdout, " --resp             Use RESP output formatting (default is JSON output)\n")
	fmt.Fprintf(os.Stdout, " -h <hostname>      Server hostname (default: %s)\n", hostname)
	fmt.Fprintf(os.Stdout, " -p <port>          Server port (default: %d)\n", port)
	fmt.Fprintf(os.Stdout, "\n")
	return false
}

func parseArgs() bool {
	defer func() {
		if v := recover(); v != nil {
			if v, ok := v.(string); ok && v == "bad arg" {
				showHelp()
			}
		}
	}()

	args := os.Args[1:]
	readArg := func(arg string) string {
		if len(args) == 0 {
			panic("bad arg")
		}
		var narg = args[0]
		args = args[1:]
		return narg
	}
	badArg := func(arg string) bool {
		fmt.Fprintf(os.Stderr, "Unrecognized option or bad number of args for: '%s'\n", arg)
		return false
	}

	for len(args) > 0 {
		arg := readArg("")
		if arg == "--help" || arg == "-?" {
			return showHelp()
		}
		if !strings.HasPrefix(arg, "-") {
			args = append([]string{arg}, args...)
			break
		}
		switch arg {
		default:
			return badArg(arg)
		case "-kml":
			tokml = true
		case "--raw":
			raw = true
		case "--tty":
			tty = true
		case "--noprompt":
			noprompt = true
		case "--resp":
			output = "resp"
		case "-h":
			hostname = readArg(arg)
		case "-p":
			n, err := strconv.ParseUint(readArg(arg), 10, 16)
			if err != nil {
				return badArg(arg)
			}
			port = int(n)
		}
	}
	oneCommand = strings.Join(args, " ")
	return true
}

func refusedErrorString(addr string) string {
	return fmt.Sprintf("Could not connect to Tile38 at %s: Connection refused", addr)
}

var groupsM = make(map[string][]string)

func main() {
	if !parseArgs() {
		return
	}

	if !raw && !tty && runtime.GOOS != "windows" {
		fi, err := os.Stdout.Stat()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		raw = (fi.Mode() & os.ModeCharDevice) == 0
	}
	if len(oneCommand) > 0 && (oneCommand[0] == 'h' || oneCommand[0] == 'H') && strings.Split(strings.ToLower(oneCommand), " ")[0] == "help" {
		showHelp()
		return
	}

	addr := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := client.Dial(addr)
	if err != nil {
		if _, ok := err.(net.Error); ok {
			fmt.Fprintln(os.Stderr, refusedErrorString(addr))
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		return
	}
	defer conn.Close()
	livemode := false
	aof := false
	defer func() {
		if livemode {
			var err error
			if aof {
				_, err = io.Copy(os.Stdout, conn.Reader())
				fmt.Fprintln(os.Stderr, "")
			} else {
				var msg []byte
				for {
					msg, err = conn.ReadMessage()
					if err != nil {
						break
					}
					fmt.Fprintln(os.Stderr, string(msg))
				}
			}
			if err != nil && err != io.EOF {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}
	}()

	line := liner.NewLiner()
	defer line.Close()

	var commands []string
	for name, command := range core.Commands {
		commands = append(commands, name)
		groupsM[command.Group] = append(groupsM[command.Group], name)
	}
	sort.Strings(commands)
	var groups []string
	for group, arr := range groupsM {
		groups = append(groups, "@"+group)
		sort.Strings(arr)
		groupsM[group] = arr
	}
	sort.Strings(groups)

	line.SetMultiLineMode(false)
	line.SetCtrlCAborts(true)
	if !(noprompt && tty) {
		line.SetCompleter(func(line string) (c []string) {
			if strings.HasPrefix(strings.ToLower(line), "help ") {
				var nitems []string
				nline := strings.TrimSpace(line[5:])
				if nline == "" || nline[0] == '@' {
					for _, n := range groups {
						if strings.HasPrefix(strings.ToLower(n), strings.ToLower(nline)) {
							nitems = append(nitems, line[:len(line)-len(nline)]+strings.ToLower(n))
						}
					}
				} else {
					for _, n := range commands {
						if strings.HasPrefix(strings.ToLower(n), strings.ToLower(nline)) {
							nitems = append(nitems, line[:len(line)-len(nline)]+strings.ToUpper(n))
						}
					}
				}
				for _, n := range nitems {
					if strings.HasPrefix(strings.ToLower(n), strings.ToLower(line)) {
						c = append(c, n)
					}
				}
			} else {
				for _, n := range commands {
					if strings.HasPrefix(strings.ToLower(n), strings.ToLower(line)) {
						c = append(c, n)
					}
				}
			}
			return
		})
	}
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	defer func() {
		if f, err := os.Create(historyFile); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}()
	if output == "resp" {
		_, err := conn.Do("output resp")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}
	for {
		var command string
		var err error
		if oneCommand == "" {
			if raw || noprompt {
				command, err = line.Prompt("")
			} else {
				command, err = line.Prompt(addr + "> ")
			}

		} else {
			command = oneCommand
		}
		if err == nil {
			nohist := strings.HasPrefix(command, " ")
			command = strings.TrimSpace(command)
			if command == "" {
				_, err := conn.Do("pInG")
				if err != nil {
					if err != io.EOF {
						fmt.Fprintln(os.Stderr, err.Error())
					} else {
						fmt.Fprintln(os.Stderr, refusedErrorString(addr))
					}
					return
				}
			} else {
				if !nohist {
					line.AppendHistory(command)
				}
				if (command[0] == 'e' || command[0] == 'E') && strings.ToLower(command) == "exit" {
					return
				}
				if (command[0] == 'q' || command[0] == 'Q') && strings.ToLower(command) == "quit" {
					return
				}
				if (command[0] == 'h' || command[0] == 'H') && (strings.ToLower(command) == "help" || strings.HasPrefix(strings.ToLower(command), "help")) {
					err = help(strings.TrimSpace(command[4:]))
					if err != nil {
						return
					}
					continue
				}
				aof = (command[0] == 'a' || command[0] == 'A') && strings.HasPrefix(strings.ToLower(command), "aof ")
				msg, err := conn.Do(command)
				if err != nil {
					if err != io.EOF {
						fmt.Fprintln(os.Stderr, err.Error())
					} else {
						fmt.Fprintln(os.Stderr, refusedErrorString(addr))
					}
					return
				}
				switch strings.ToLower(command) {
				case "output resp":
					if string(msg) == "+OK\r\n" {
						output = "resp"
					}
				case "output json":
					if strings.HasPrefix(string(msg), `{"ok":true`) {
						output = "json"
					}
				}

				mustOutput := true
				if oneCommand == "" && !strings.HasPrefix(string(msg), `{"ok":true`) {
					var cerr connError
					if err := json.Unmarshal(msg, &cerr); err == nil {
						fmt.Fprintln(os.Stderr, "(error) "+cerr.Err)
						mustOutput = false
					}
				} else if string(msg) == client.LiveJSON {
					fmt.Fprintln(os.Stderr, string(msg))
					livemode = true
					break // break out of prompt and just feed data to screen
				}
				if mustOutput {
					if tokml {
						msg = convert2kml(msg)
						fmt.Fprintln(os.Stdout, string(msg))
					} else if output == "resp" {
						if !raw {
							msg = convert2termresp(msg)
						}
						fmt.Fprintln(os.Stdout, string(msg))
					} else {
						if raw {
							fmt.Fprintln(os.Stdout, string(msg))
						} else {
							fmt.Fprintln(os.Stdout, string(msg))
						}
					}
				}
			}
		} else if err == liner.ErrPromptAborted {
			return
		} else {
			fmt.Fprintf(os.Stderr, "Error reading line: %s", err.Error())
		}
		if oneCommand != "" {
			return
		}
	}
}

func convert2termresp(msg []byte) []byte {
	rd := resp.NewReader(bytes.NewBuffer(msg))
	out := ""
	for {
		v, _, err := rd.ReadValue()
		if err != nil {
			break
		}
		out += convert2termrespval(v, 0)
	}
	return []byte(strings.TrimSpace(out))
}

func convert2termrespval(v resp.Value, spaces int) string {
	switch v.Type() {
	default:
		return v.String()
	case resp.BulkString:
		if v.IsNull() {
			return "(nil)"
		}
		return "\"" + v.String() + "\""
	case resp.Integer:
		return "(integer) " + v.String()
	case resp.Error:
		return "(error) " + v.String()
	case resp.Array:
		arr := v.Array()
		if len(arr) == 0 {
			return "(empty list or set)"
		}
		out := ""
		nspaces := spaces + numlen(len(arr))
		for i, v := range arr {
			if i > 0 {
				out += strings.Repeat(" ", spaces)
			}
			iout := strings.TrimSpace(convert2termrespval(v, nspaces+2))
			out += fmt.Sprintf("%d) %s\n", i+1, iout)
		}
		return out
	}
}

func numlen(n int) int {
	l := 1
	if n < 0 {
		l++
		n = n * -1
	}
	for i := 0; i < 1000; i++ {
		if n < 10 {
			break
		}
		l++
		n = n / 10
	}
	return l
}

func convert2kml(msg []byte) []byte {
	k := NewKML()
	var m map[string]interface{}
	if err := json.Unmarshal(msg, &m); err == nil {
		if v, ok := m["points"].([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(map[string]interface{}); ok {
					if v, ok := v["point"].(map[string]interface{}); ok {
						var name string
						var lat, lon float64
						if v, ok := v["id"].(string); ok {
							name = v
						}
						if v, ok := v["lat"].(float64); ok {
							lat = v
						}
						if v, ok := v["lon"].(float64); ok {
							lon = v
						}
						k.AddPoint(name, lat, lon)
					}
				}
			}
		}
		return k.Bytes()
	}
	return []byte(`{"ok":false,"err":"results must contain points"}`)
}

func help(arg string) error {
	var groupsA []string
	for group := range groupsM {
		groupsA = append(groupsA, "@"+group)
	}
	groups := "Groups: " + strings.Join(groupsA, ", ") + "\n"

	if arg == "" {
		fmt.Fprintf(os.Stderr, "tile38-cli %s (git:%s)\n", core.Version, core.GitSHA)
		fmt.Fprintf(os.Stderr, `Type:   "help @<group>" to get a list of commands in <group>`+"\n")
		fmt.Fprintf(os.Stderr, `        "help <command>" for help on <command>`+"\n")
		if !(noprompt && tty) {
			fmt.Fprintf(os.Stderr, `        "help <tab>" to get a list of possible help topics`+"\n")
		}
		fmt.Fprintf(os.Stderr, `        "quit" to exit`+"\n")
		if noprompt && tty {
			fmt.Fprintf(os.Stderr, groups)
		}
		return nil
	}
	showGroups := false
	found := false
	if strings.HasPrefix(arg, "@") {
		for _, command := range groupsM[arg[1:]] {
			fmt.Fprintf(os.Stderr, "%s\n", core.Commands[command].TermOutput("  "))
			found = true
		}
		if !found {
			showGroups = true
		}
	} else {
		if command, ok := core.Commands[strings.ToUpper(arg)]; ok {
			fmt.Fprintf(os.Stderr, "%s\n", command.TermOutput("  "))
			found = true
		}
	}
	if showGroups {
		if noprompt && tty {
			fmt.Fprintf(os.Stderr, groups)
		}
	} else if !found {
		if noprompt && tty {
			help("")
		}
	}
	return nil
}
