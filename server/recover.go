package server

import (
	"runtime"
)

// RecoverHandler defines a function which calls the provided ServeFunc
// within a protective recover() function.
func RecoverHandler(serveFn ServeFunc) error {
	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			log.Errorf("Error: %s", err)
			log.Debugf("Stack of %d bytes: %s\n", count, string(trace))
			return
		}
	}()

	if err := serveFn(); err != nil {
		log.Error("Error: ", err)
		return err
	}

	return nil
}
