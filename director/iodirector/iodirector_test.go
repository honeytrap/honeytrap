package iodirector_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director/iodirector"
	"github.com/honeytrap/honeytrap/utils/tests"
)

// TestIODirector defines a test to validates the behaviour of the iodirector.
func TestIODirector(t *testing.T) {
	director := iodirector.New(&config.Config{
		Director: "io",
		Directors: config.DirectorConfig{
			IOConfig: config.IOConfig{
				ServiceAddr: "google.com:80",
			},
		},
	}, nil)

	container, err := director.NewContainer("127.0.0.1:3000")
	if err != nil {
		tests.Failed("Should have successfully created container: %+q.", err)
	}
	tests.Passed("Should have successfully created container.")

	conn, err := container.Dial()
	if err != nil {
		tests.Failed("Should have successfully created connection by container: %+q.", err)
	}
	tests.Passed("Should have successfully created connection by container.")

	fmt.Fprintf(conn, "GET /\r\n")

	var bu bytes.Buffer

	for {
		memory := make([]byte, 1024)
		n, err := conn.Read(memory)
		if err != nil {
			break
		}

		bu.Write(memory[:n])
	}

	if bu.Len() == 0 {
		tests.Failed("Should have successfully retrieved data from connection.")
	}
	tests.Passed("Should have successfully retrieved data from connection.")
}
