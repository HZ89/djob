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

package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/HZ89/djob/djob"
	"github.com/mitchellh/cli"
)

type AgentCmd struct {
	Ui      cli.Ui
	agent   *djob.Agent
	args    []string
	Version string
}

func (c *AgentCmd) Run(args []string) int {
	c.args = make([]string, len(args))
	copy(c.args, args)
	c.agent = djob.New(c.args, c.Version)
	err := c.agent.Run()
	if err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error("Cmd: Start agent failed")
		return 1
	}
	return c.handleSignals()
}

func (c *AgentCmd) handleSignals() int {
	signalCh := make(chan os.Signal, 8)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	if err := c.agent.WritePid(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	for {
		switch sig := <-signalCh; sig {
		case syscall.SIGHUP:
			c.Ui.Info("Got OS SIGHUP reload")
			c.reload()
		case syscall.SIGTERM:
			c.Ui.Info("Got OS SIGTERM stop")
			return c.stop(false)
		case syscall.SIGQUIT:
			c.Ui.Info("Got OS SIGQUIT stop")
			return c.stop(true)
		case os.Interrupt:
			c.Ui.Warn("Got OS Interrupt signal quit immediately")
			return 1
		default:
			c.Ui.Warn("Got a unknown system call signal")
		}
	}
}

func (c *AgentCmd) reload() {
	c.Ui.Info("Reloading...")
	c.agent.Reload(c.args)
	return
}

func (c *AgentCmd) stop(graceful bool) int {
	i := c.agent.Stop(graceful)
	if err := c.agent.RemovePid(); err != nil {
		c.Ui.Warn(err.Error())
	}
	return i
}

func (c *AgentCmd) Synopsis() string {
	return "Run djob agent"
}

func (c *AgentCmd) Help() string {
	helpText := `
	Usage: djob agent [options]
	    Run djob agent
	Options:
	    --config              config file path
	    --pid                 pid file path
	    --logfile             log file path
	`
	return strings.TrimSpace(helpText)
}
