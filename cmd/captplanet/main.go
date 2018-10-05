// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"
	"storj.io/storj/internal/memory"
	"storj.io/storj/pkg/process"
)

var (
	mon = monkit.Package()

	rootCmd = &cobra.Command{
		Use:   "captplanet",
		Short: "Captain Planet! With our powers combined!",
	}

	defaultConfDir = "$HOME/.storj/capt"
)

func init() {
	exe, err := os.Executable()
	if err == nil {
		rootCmd.Use = exe
	}
}

func main() {
	go dumpHandler()
	// Figure out the executable name.
	exe, err := os.Executable()
	if err != nil {
		exe = "captplanet.exe"
	}
	// Make the windows graphical launch text a bit more friendly.
	cobra.MousetrapHelpText = fmt.Sprintf("This is a command line tool.\n\n"+
		"This needs to be run from a Command Prompt.\n"+
		"Try running \"%s help\" for more information", exe)

	// process.Exec will load this for this command.
	runCmd.Flags().String("config",
		filepath.Join(defaultConfDir, "config.yaml"), "path to configuration")
	setupCmd.Flags().String("config",
		filepath.Join(defaultConfDir, "setup.yaml"), "path to configuration")
	process.Exec(rootCmd)
}

// dumpHandler listens for Ctrl+\ on Unix
func dumpHandler() {
	if runtime.GOOS == "windows" {
		// unsupported on Windows
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	for range sigs {
		dumpGoroutines()
	}
}

func dumpGoroutines() {
	buf := make([]byte, memory.MB)
	n := runtime.Stack(buf, true)

	p := time.Now().Format("dump-2006-01-02T15-04-05.999999999.log")
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	fmt.Fprintf(os.Stderr, "Writing stack traces to \"%v\"\n", p)

	err := ioutil.WriteFile(p, buf[:n], 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
