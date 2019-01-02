package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	osStdin  = os.Stdin
	osStdout = os.Stdout
	osStderr = os.Stderr
	osExit   = os.Exit
)

type Config struct {
	In       io.Reader
	Out      io.Writer
	Err      io.Writer
	ExitFunc func(int)
}

func (c Config) Input() io.Reader {
	if c.In == nil {
		return osStdin
	}
	return c.In
}

func (c Config) Output() io.Writer {
	if c.Out == nil {
		return osStdout
	}
	return c.Out
}

func (c Config) Error() io.Writer {
	if c.Err == nil {
		return osStderr
	}
	return c.Err
}

func (c Config) Exit(code int) {
	if c.ExitFunc == nil {
		osExit(code)
		// return in case
		// of exit function not returning
		return
	}
	c.ExitFunc(code)
}

type Command struct {
	*cobra.Command
	Config
}
