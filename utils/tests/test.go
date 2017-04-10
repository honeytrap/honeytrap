package tests

import (
	"fmt"
	"log"
	"os"
	"testing"
)

// succeedMark is the Unicode codepoint for a check mark.
const succeedMark = "\u2713"

var logger = log.New(os.Stdout, "", log.Lshortfile)

// Info logs the info message using the giving message and values.
func Info(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t-\t %s\n", fmt.Sprintf(message, val...)))
	}
}

// Passed logs the failure message using the giving message and values.
func Passed(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t%s\t %s\n", succeedMark, fmt.Sprintf(message, val...)))
	}
}

// failedMark is the Unicode codepoint for an X mark.
const failedMark = "\u2717"

// Failed logs the failure message using the giving message and values.
func Failed(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t%s\t %s\n", failedMark, fmt.Sprintf(message, val...)))
	}

	os.Exit(1)
}

// Errored logs the error message using the giving message and values.
func Errored(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t%s\t %s\n", failedMark, fmt.Sprintf(message, val...)))
	}
}
