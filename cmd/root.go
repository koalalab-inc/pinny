/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package cmd

import (
	"os"

	"github.com/koalalab-inc/pinny/cmd/actions"
	"github.com/koalalab-inc/pinny/cmd/docker"

	"github.com/spf13/cobra"
)

var currentVersion, currentVersionSet = os.LookupEnv("PINNY_VERSION")

var rootCmd = NewRootCmd()

func NewRootCmd() *cobra.Command {
	if !currentVersionSet {
		currentVersion = "0.0.1"
	}
	return &cobra.Command{
		Use:     "pinny",
		Short:   "\nHash-pining for your OSS dependencies",
		Version: currentVersion,
	}
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	rootCmd.SetVersionTemplate("Pinny v{{.Version}}\n")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.AddCommand(docker.DockerCmd)
	rootCmd.AddCommand(actions.ActionsCmd)
}
