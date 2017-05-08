package process_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/honeytrap/honeytrap/process"
	"github.com/honeytrap/tests"
)

// TestCommandProcess validates the behaviours of the process.Command structure.
func TestCommandSyncProcess(t *testing.T) {
	src := process.SyncProcess{
		Commands: []process.Command{
			process.Command{
				Name: "echo",
				Args: []string{"New Login"},
			},
		},
	}

	ctx := context.Background()

	var errBu, outBu bytes.Buffer
	err := src.SyncExec(ctx, &outBu, &errBu)
	if err != nil {
		tests.Failed("Should have successfully executed shell script: %+q.", err)
	}
	tests.Passed("Should have successfully executed shell script.")

	if outBu.String() != "New Login\n" {
		tests.Failed("Should have successfully matched output data with expected value: %+q.", outBu.String())
	}
	tests.Passed("Should have successfully matched output data with expected value.")
}

// TestScriptProcess to validate the behaviour of executing a shell script
// as a file within the executing system.
func TestScriptProcess(t *testing.T) {
	src := process.ScriptProcess{
		Shell:  "/bin/bash",
		Source: `echo "New Login"`,
	}

	ctx := context.Background()

	var errBu, outBu bytes.Buffer
	err := src.Exec(ctx, &outBu, &errBu)
	if err != nil {
		tests.Failed("Should have successfully executed shell script: %+q.", err)
	}
	tests.Passed("Should have successfully executed shell script.")

	if outBu.String() != "New Login\n" {
		tests.Failed("Should have successfully matched output data with expected value: %+q.", outBu.String())
	}
	tests.Passed("Should have successfully matched output data with expected value.")
}

// TestScriptProcessWithCancel to validate the behaviour of executing a shell script
// with a canceling call from the provided context.
func TestScriptProcessWithCancel(t *testing.T) {
	src := process.ScriptProcess{
		Shell: "/bin/bash",
		Source: `echo "New Login"
date
sleep 10
date
echo "Lets run"`,
	}

	var errBu, outBu bytes.Buffer

	ctx, canceller := context.WithTimeout(context.Background(), 5*time.Millisecond)

	defer canceller()

	err := src.Exec(ctx, &outBu, &errBu)
	if err == nil {
		tests.Failed("Should have successfully being killed by kill signal.")
	}
	tests.Passed("Should have successfully being killed by kill signal.")
}
