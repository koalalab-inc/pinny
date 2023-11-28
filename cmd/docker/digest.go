/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package docker

import (
	"github.com/koalalab-inc/pinny/pkg/docker"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:   "digest <docker-image>",
	Short: "Get the digest of a Docker image",
	Args:  cobra.ExactArgs(1),
	Long: `
	Get the digest of a Docker image
	
	<docker-image> [Required Argument]
	| 
	| The Docker image to get the digest for. This can be in the form of
	| docker://<image-ref>:tag or docker://<image-ref>
	| latest tag is used if tag is not specified
	| 
	| e.g.:
	| alpine:3.18 or gcr.io/distroless/java11-debian11:nonroot
	| 
	|> pinny docker digest alpine:3.18
	|> sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978
	
`,

	Run: func(cmd *cobra.Command, args []string) {
		imageString := args[0]
		digest, err := docker.GetDigest(imageString)
		cobra.CheckErr(err)
		cmd.OutOrStdout().Write([]byte(*digest))
	},
}

func init() {}
