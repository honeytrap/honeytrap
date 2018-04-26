package scripter

import (
	"encoding/json"
	"fmt"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/utils/files"
	"time"
)

// getRemoteAddr returns a function that returns the remote address of a connection
func getRemoteAddr(c ScrConn) func() string {
	return func() string { return c.GetConn().RemoteAddr().String() }
}

// getLocalAddr returns a function that returns the local address of a connection
func getLocalAddr(c ScrConn) func() string {
	return func() string { return c.GetConn().LocalAddr().String() }
}

// getDatetime returns a function that returns the datetime in unix format
func getDatetime() func() string {
	return func() string {
		t := time.Now()
		return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
	}
}

// getFileDownload returns a function that downloads a file from URL
func getFileDownload(c ScrConn, service string) func() string {
	return func() string {
		params, _ := c.GetParameters([]string{"url", "path"}, service)

		if err := files.Download(params["url"], params["path"]); err != nil {
			log.Errorf("error downloading file: %s", err)
			return "no"
		}
		return "yes"
	}
}

// channelSend returns a function that sends a event over the channel
func channelSend(s Scripter, c ScrConn, service string) func() {
	return func() {
		params, _ := c.GetParameters([]string{"data"}, service)
		var data map[string]interface{}

		json.Unmarshal([]byte(params["data"]), &data)

		message := event.New()
		for key, value := range data {
			event.Custom(key, value)(message)
		}
		event.Custom("destination-ip", c.GetConn().LocalAddr().String())(message)
		event.Custom("source-ip", c.GetConn().RemoteAddr().String())(message)

		s.GetChannel().Send(message)
	}
}

// doLog returns a function that can log a certain message on different log types
func doLog(c ScrConn, service string) func() {
	return func() {
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
	}
}

// getFolder returns a function that returns the script folder path
func getFolder(s Scripter) func() string {
	return func() string {
		return s.GetScriptFolder()
	}
}

// SetBasicMethods sets methods that can be called by each script, returning basic functionality for the scripts
// initiated in the scripter
func SetBasicMethods(s Scripter, c ScrConn, service string) {
	c.SetStringFunction("getRemoteAddr", getRemoteAddr(c), service)
	c.SetStringFunction("getLocalAddr", getLocalAddr(c), service)

	c.SetStringFunction("getDatetime", getDatetime(), service)

	c.SetStringFunction("getFileDownload", getFileDownload(c, service), service)
	c.SetStringFunction("getFolder", getFolder(s), service)

	c.SetVoidFunction("channelSend", channelSend(s, c, service), service)

	c.SetVoidFunction("doLog", doLog(c, service), service)
}
