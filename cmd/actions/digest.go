/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package actions

import (
	"github.com/koalalab-inc/pinny/pkg/actions"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:   "digest <action>",
	Short: "Get the digest of a Github Action",
	Long: `
	Get the digest of a Github Action

	<action> [Required Argument]
	| 
	| The Github Action to get the digest for. This can be in the form of
	| <owner>/<repo>/<path>@<ref> or <owner>/<repo>@<ref> for  github repo actions
	| 
	| e.g.:
	| actions/checkout@v3 or koalalab-inc/pinny.
	|
	|> pinny actions digest actions/checkout@v3
	|> 93ea575cb5d8a053eaa0ac8fa3b40d7e05a33cc8
	| 
	| To get the digest for a docker action, use the command:
	| 
	| pinny docker digest <docker-image>
	| 
	| Run pinny docker digest --help for more information
	
`,
	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		actionString := args[0]
		digest, err := actions.GetDigest(actionString)
		cobra.CheckErr(err)
		cmd.OutOrStdout().Write([]byte(*digest))
	},
}

func init() {}
