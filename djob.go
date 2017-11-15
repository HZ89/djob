/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"os"

	"github.com/HZ89/djob/cmd"
	"github.com/mitchellh/cli"
)

const VERSION = "0.1.0"

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
