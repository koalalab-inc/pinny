/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package actions

import (
	"github.com/spf13/cobra"
)

var actionsHelpTemplate = `
{{.Name}} - {{.Short}}

Usage:
	{{.UseLine}}

	If you are being limited by the Github API, you can set the GITHUB_TOKEN
	environment variable to a Github Personal Access Token to increase your
	rate limit. Your PAT should have the repo scope.

	GITHUB_TOKEN=<your personal access token> {{.UseLine}}

	Read more about Github Personal Access Tokens here:
	https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens

Options:
	{{.LocalFlags.FlagUsages | trimRightSpace}}
{{if gt (len .Commands) 0}}
Available Commands:
{{range .Commands}}{{if .IsAvailableCommand}}
	{{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}

Description:
	{{.Long}}
`

var ActionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "\nHash-pining for your third party Github Actions",
}

func init() {
	commands := []*cobra.Command{
		pinCmd,
		digestCmd,
	}
	for _, cmd := range commands {
		cmd.SetHelpTemplate(actionsHelpTemplate)
		ActionsCmd.AddCommand(cmd)
	}
}
