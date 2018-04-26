package generic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/honeytrap/honeytrap/scripter"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// setMethods sets the methods required for the generic scripts in the Handle method of the generic service
func (s *genericService) setMethods(connW scripter.ConnectionWrapper) error {
	if setErr := connW.SetStringFunction("getRequest", getRequest(connW)); setErr != nil {
		 return setErr
	}

	if setErr := connW.SetVoidFunction("restWrite", restWrite(connW)); setErr != nil {
		return setErr
	}

	return nil
}

//restWrite returns a function that can write a REST response on a connection
func restWrite(connW scripter.ConnectionWrapper) func() {
	return func() {
		params, _ := connW.GetParameters([]string{"status", "response", "headers"})

		status, _ := strconv.Atoi(params["status"])
		buf := connW.GetScrConn().GetConnectionBuffer()
		br := bufio.NewReader(buf)

		req, err := http.ReadRequest(br)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Errorf("Error while reading buffered request connection, %s", err)
			return
		}

		defer req.Body.Close()

		header := http.Header{}

		header.Set("date", (time.Now()).String())
		header.Set("connection", "Keep-Alive")
		header.Set("content-type", "application/json")

		var headers map[string]string
		json.Unmarshal([]byte(params["data"]), &headers)
		for name, value := range headers {
			header.Set(name, value)
		}

		resp := http.Response{
			StatusCode:    status,
			Status:        http.StatusText(status),
			Proto:         req.Proto,
			ProtoMajor:    req.ProtoMajor,
			ProtoMinor:    req.ProtoMinor,
			Request:       req,
			Header:        header,
			Body:          ioutil.NopCloser(bytes.NewBufferString(params["response"])),
			ContentLength: int64(len(params["response"])),
		}

		if err := resp.Write(connW.GetScrConn().GetConn()); err != nil {
			log.Errorf("Writing of scripter - REST message was not successful, %s", err)
		}
	}
}

//getRequest returns a function that can read in a HTTP request from a connection in JSON
func getRequest(connW scripter.ConnectionWrapper) func() string {
	return func() string {
		params, _ := connW.GetParameters([]string{"withBody"})

		buf := connW.GetScrConn().GetConnectionBuffer()
		buf.Reset()
		tee := io.TeeReader(connW.GetScrConn().GetConn(), buf)

		br := bufio.NewReader(tee)

		req, err := http.ReadRequest(br)
		if err == io.EOF {
			log.Infof("Payload is empty.", err)
			return ""
		} else if err != nil {
			log.Errorf("Failed to parse payload to HTTP Request, Error: %s", err)
			return ""
		}

		m := map[string]interface{}{}
		m["method"] = req.Method
		m["header"] = req.Header
		m["host"] = req.Host
		m["form"] = req.Form
		body := make([]byte, 1024)
		if params["withBody"] == "1" {

			defer req.Body.Close()
			n, _ := req.Body.Read(body)

			body = body[:n]
			var js2 map[string]interface{}
			if json.Unmarshal([]byte(body), &js2) == nil {
				m["body"] = js2
			} else {
				m["body"] = string(body)
			}
		}

		result, err := json.Marshal(m)
		if err != nil {
			log.Errorf("Failed to parse request struct to json, Error: %s", err)
			return "{}"
		}

		return string(result)
	}
}
