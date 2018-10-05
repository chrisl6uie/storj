// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"storj.io/storj/pkg/cfgstruct"
	"storj.io/storj/pkg/kademlia"
	psserver "storj.io/storj/pkg/piecestore/rpc/server"
	"storj.io/storj/pkg/process"
	"storj.io/storj/pkg/provider"
)

var (
	rootCmd = &cobra.Command{
		Use:   "storagenode",
		Short: "StorageNode",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the storagenode",
		RunE:  cmdRun,
	}
	setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Create config files",
		RunE:  cmdSetup,
	}

	runCfg struct {
		Identity provider.IdentityConfig
		Kademlia kademlia.Config
		Storage  psserver.Config
	}
	setupCfg struct {
		BasePath string `default:"$CONFDIR" help:"base path for setup"`
		CA       provider.CASetupConfig
		Identity provider.IdentitySetupConfig
	}

	defaultConfDir = "$HOME/.storj/storagenode"
)

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
	cfgstruct.Bind(runCmd.Flags(), &runCfg, cfgstruct.ConfDir(defaultConfDir))
	cfgstruct.Bind(setupCmd.Flags(), &setupCfg, cfgstruct.ConfDir(defaultConfDir))

	exe, err := os.Executable()
	if err == nil {
		rootCmd.Use = exe
	}
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	return runCfg.Identity.Run(process.Ctx(cmd), runCfg.Kademlia, runCfg.Storage)
}

func cmdSetup(cmd *cobra.Command, args []string) (err error) {
	setupCfg.BasePath, err = filepath.Abs(setupCfg.BasePath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(setupCfg.BasePath, 0700)
	if err != nil {
		return err
	}

	setupCfg.CA.CertPath = filepath.Join(setupCfg.BasePath, "ca.cert")
	setupCfg.CA.KeyPath = filepath.Join(setupCfg.BasePath, "ca.key")
	setupCfg.Identity.CertPath = filepath.Join(setupCfg.BasePath, "identity.cert")
	setupCfg.Identity.KeyPath = filepath.Join(setupCfg.BasePath, "identity.key")

	err = provider.SetupIdentity(process.Ctx(cmd), setupCfg.CA, setupCfg.Identity)
	if err != nil {
		return err
	}

	overrides := map[string]interface{}{
		"identity.cert-path": setupCfg.Identity.CertPath,
		"identity.key-path":  setupCfg.Identity.KeyPath,
		"storage.path":       filepath.Join(setupCfg.BasePath, "storage"),
	}

	return process.SaveConfig(runCmd.Flags(),
		filepath.Join(setupCfg.BasePath, "config.yaml"), overrides)
}

func main() {
	// Figure out the executable name.
	exe, err := os.Executable()
	if err != nil {
		exe = "storagenode.exe"
	}
	// Make the windows graphical launch text a bit more friendly.
	cobra.MousetrapHelpText = fmt.Sprintf("This is a command line tool.\n\n"+
		"This needs to be run from a Command Prompt.\n"+
		"Try running \"%s help\" for more information", exe)
	runCmd.Flags().String("config",
		filepath.Join(defaultConfDir, "config.yaml"), "path to configuration")
	process.Exec(rootCmd)
}
