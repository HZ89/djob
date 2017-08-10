package cmd

import (
	"github.com/mitchellh/cli"
	"version.uuzu.com/zhuhuipeng/djob/djob"
	"os"
	"os/signal"
	"syscall"
)

type AgentCmd struct {
	Ui      cli.Ui
	agent   *djob.Agent
	args    []string
	version string
}

func (c *AgentCmd) Run(args []string, verison string) int {
	copy(c.args, args)
	c.version = verison
	c.agent = djob.New(c.args, c.version)
	go c.agent.Run()
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
