package utils

import (
	"runtime"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:utils")

// RecoverHandler defines a function which is called to handle panic recovery.
func RecoverHandler() {
	if err := recover(); err != nil {
		trace := make([]byte, 1024)
		count := runtime.Stack(trace, true)
		log.Errorf("Error: %s", err)
		log.Debugf("Stack of %d bytes: %s\n", count, string(trace))
		return
	}
}
