package cmd

import (
	"github.com/mitchellh/cli"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"version.uuzu.com/zhuhuipeng/djob/djob"
)

type AgentCmd struct {
	Ui      cli.Ui
	agent   *djob.Agent
	args    []string
	Version string
}

func (c *AgentCmd) Run(args []string) int {
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

	for {
		switch sig := <-signalCh; sig {
		case syscall.SIGHUP:
			c.reload()
		case syscall.SIGTERM:
			return c.stop(false)
		case syscall.SIGQUIT:
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
	return c.agent.Stop(graceful)
}

func (c *AgentCmd) Synopsis() string {
	return "Run djob agent"
}

func (c *AgentCmd) Help() string {
	helpText := `
	Usage: djob agent [options]
	    Run djob agent
	Options:
	    --config=./config     config file path
	    --pid                 pid file path
	    --logfile             log file path
	`
	return strings.TrimSpace(helpText)
}
