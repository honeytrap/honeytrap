package process

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("Honeytrap")

// Command defines the command to be executed and it's arguments
type Command struct {
	Name string
	Args []string
}

// Run executes the giving command and returns the bytes.Buffer for both
// the Stdout and Stderr.
func (c Command) Run(ctx context.Context, out, err io.Writer) error {
	proc := exec.Command(c.Name, c.Args...)
	proc.Stdout = out
	proc.Stderr = err

	if err := proc.Start(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		if proc.Process != nil {
			proc.Process.Kill()
		}
	}()

	if err := proc.Wait(); err != nil {
		return err
	}

	return nil
}

//============================================================================================

// SyncProcess defines a struct which is used to execute a giving set of
// script values.
type SyncProcess struct {
	Commands []Command `json:"commands"`
}

// SyncExec executes the giving series of commands attached to the
// process.
func (p SyncProcess) SyncExec(ctx context.Context, pipeOut, pipeErr io.Writer) error {
	for _, command := range p.Commands {
		if err := command.Run(ctx, pipeOut, pipeErr); err != nil {
			return err
		}
	}

	return nil
}

//============================================================================================

// AsyncProcess defines a struct which is used to execute a giving set of
// script values.
type AsyncProcess struct {
	Commands []Command `json:"commands"`
}

// AsyncExec executes the giving series of commands attached to the
// process.
func (p AsyncProcess) AsyncExec(ctx context.Context, pipeOut, pipeErr io.Writer) error {
	var waiter sync.WaitGroup

	for _, command := range p.Commands {
		go func(cmd Command) {
			waiter.Add(1)
			defer waiter.Done()

			cmd.Run(ctx, pipeOut, pipeErr)
		}(command)
	}

	waiter.Wait()
	return nil
}

//============================================================================================

// ScriptProcess defines a shell script execution structure which attempts to copy
// given script into a local file path and attempts to execute content.
// Shell states the shell to be used for execution: /bin/sh, /bin/bash
type ScriptProcess struct {
	Source string `json:"source"`
	Shell  string `json:"shell"`
}

// Exec executes a copy of the giving script source in a temporary file which it then executes
// the contents.
func (sp ScriptProcess) Exec(ctx context.Context, pipeOut, pipeErr io.Writer) error {
	tmpFile, err := ioutil.TempFile("/tmp", "proc-shell")
	if err != nil {
		return err
	}

	if _, err := tmpFile.Write([]byte(sp.Source)); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}

	tmpFile.Close()

	defer os.Remove(tmpFile.Name())

	proc := exec.Command(sp.Shell, tmpFile.Name())
	proc.Stdout = pipeOut
	proc.Stderr = pipeErr

	if err := proc.Start(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		if proc.Process != nil {
			proc.Process.Kill()
		}
	}()

	if err := proc.Wait(); err != nil {
		return err
	}

	return nil
}
