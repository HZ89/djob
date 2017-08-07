package cmd

import (
	"github.com/mitchellh/cli"
	"bytes"
	"fmt"
	"github.com/hashicorp/serf/serf"
)


type VersionCmd struct {
	Version string
	Ui      cli.Ui
}

func (c *VersionCmd) Help() string {
	return ""
}

func (c *VersionCmd) Run(_ []string) int {
	var version bytes.Buffer
	fmt.Fprintf(&version, "Djob version %s", c.Version)

	c.Ui.Output(version.String())
	c.Ui.Output(fmt.Sprintf("Serf Agent Protocol: %d (Understands back to :%d)", serf.ProtocolVersionMax, serf.ProtocolVersionMin))
	return 0
}

func (c *VersionCmd) Synopsis() string {
	return "Prints the Djob version"
}
