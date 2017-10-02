package main

import (
	"fmt"
	"io"

	"github.com/urfave/cli"
)

const (
	// ExitCodeOK ...
	ExitCodeOK int = 0
	// ExitCodeError ..
	ExitCodeError int = 1
)

// CLI ...
type CLI struct {
	outStream io.Writer
	errStream io.Writer
}

// Run ...
func (c *CLI) Run(args []string) int {
	app := cli.NewApp()
	app.Name = "epgrec-program-finder"
	app.Version = "0.0.1"
	app.Usage = "epgrec program finder"
	app.Action = func(ctx *cli.Context) error {
		fmt.Println("Hello world!")
		return nil
	}

	err := app.Run(args)
	if err != nil {
		fmt.Fprintln(c.errStream, err)
		return ExitCodeError
	}

	return ExitCodeOK
}
