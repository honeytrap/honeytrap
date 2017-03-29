package controller

import (
	"strings"
	"time"

	"github.com/tidwall/tile38/controller/server"
)

func (c *Controller) cmdOutput(msg *server.Message) (res string, err error) {
	start := time.Now()
	vs := msg.Values[1:]
	var arg string
	var ok bool
	if len(vs) != 0 {
		if _, arg, ok = tokenval(vs); !ok || arg == "" {
			return "", errInvalidNumberOfArguments
		}
		// Setting the original message output type will be picked up by the
		// server prior to the next command being executed.
		switch strings.ToLower(arg) {
		default:
			return "", errInvalidArgument(arg)
		case "json":
			msg.OutputType = server.JSON
		case "resp":
			msg.OutputType = server.RESP
		}
		return server.OKMessage(msg, start), nil
	}
	// return the output
	switch msg.OutputType {
	default:
		return "", nil
	case server.JSON:
		return `{"ok":true,"output":"json","elapsed":` + time.Now().Sub(start).String() + `}`, nil
	case server.RESP:
		return "$4\r\nresp\r\n", nil
	}
}
