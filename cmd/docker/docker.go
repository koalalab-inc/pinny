/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package docker

import (
	"github.com/spf13/cobra"
)

var dockerHelpTemplate = `
{{.Name}} - {{.Short}}

Usage:
	{{.UseLine}}

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

var dockerfile string
var inplace bool

var DockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "\nHash-pining for your third party Docker images",
}

func init() {
	commands := []*cobra.Command{
		pinCmd,
		transformCmd,
		digestCmd,
		lockCmd,
	}
	for _, cmd := range commands {
		cmd.SetHelpTemplate(dockerHelpTemplate)
		DockerCmd.AddCommand(cmd)
	}
}
