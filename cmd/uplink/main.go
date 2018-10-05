// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"storj.io/storj/cmd/uplink/cmd"
	"storj.io/storj/pkg/process"
)

func main() {
	// Figure out the executable name.
	exe, err := os.Executable()
	if err != nil {
		exe = "uplink.exe"
	}
	// Make the windows graphical launch text a bit more friendly.
	cobra.MousetrapHelpText = fmt.Sprintf("This is a command line tool.\n\n"+
		"This needs to be run from a Command Prompt.\n"+
		"Try running \"%s help\" for more information", exe)
	process.Exec(cmd.RootCmd)
}
