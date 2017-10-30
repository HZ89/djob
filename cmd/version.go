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
	"bytes"
	"fmt"

	"github.com/hashicorp/serf/serf"
	"github.com/mitchellh/cli"
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
