package main

import (
	"fmt"
	"github.com/mitchellh/cli"
	"os"
	"version.uuzu.com/zhuhuipeng/djob/cmd"
)

const VERSION string = "0.1.0"

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}

	c := cli.NewCLI("djob", VERSION)
	c.Args = args
	c.HelpFunc = cli.BasicHelpFunc("djob")

	ui := &cli.BasicUi{Writer: os.Stdout}

	c.Commands = map[string]cli.CommandFactory{
		"version": func() (cli.Command, error) {
			return &cmd.VersionCmd{
				Version: VERSION,
				Ui:      ui,
			}, nil
		},
		"keygen": func() (cli.Command, error) {
			return &cmd.KeygenCmd{
				Ui: ui,
			}, nil
		},
		"agent": func() (cli.Command, error) {
			return &cmd.AgentCmd{
				Ui:      ui,
				Version: VERSION,
			}, nil
		},
	}

	exitStatus, err := c.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(exitStatus)
}
