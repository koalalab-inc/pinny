/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package cmd

import (
	"github.com/koalalab-inc/pinny/cmd/actions"
	"github.com/koalalab-inc/pinny/cmd/docker"

	"github.com/spf13/cobra"
)

var version string

var rootCmd = NewRootCmd()

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "pinny",
		Short:   "\nHash-pining for your OSS dependencies",
		Version: version,
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
