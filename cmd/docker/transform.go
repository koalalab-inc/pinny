/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package docker

import (
	"fmt"
	"os"

	"github.com/koalalab-inc/pinny/pkg/docker"
	"github.com/spf13/cobra"
)

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Pin all third party Docker images used in your Dockerfile using lock file only",
	Long: `
	Pin all third party Docker images used in your Dockerfile using lock file only

	Similar to pin command, but uses lock file only to pin Docker images.
	
	Throws an error if lock file is not present.
	To generate a lock file, Use: 
	|> pinny docker lock

	Example:
		pinny docker transform
		pinny docker transform [-f Dockerfile]
		pinny docker transform -f DevDockerfile

	See help for pin command for more details.
	> pinny docker pin --help
`,
	Run: func(cmd *cobra.Command, args []string) {
		offline := true
		err := docker.GeneratePinnedDockerfile(dockerfile, offline)
		srcFile := dockerfile
		tmpFile := fmt.Sprintf("%s.pinned.tmp", dockerfile)
		outFile := fmt.Sprintf("%s.pinned", dockerfile)
		if inplace {
			outFile = srcFile
		}
		if err != nil {
			os.Remove(tmpFile)
		} else {
			err = os.Rename(tmpFile, outFile)
		}
		cobra.CheckErr(err)
	},
}

func init() {
	transformCmd.Flags().StringVarP(&dockerfile, "dockerfile", "f", "Dockerfile", "Dockerfile to use")
	transformCmd.Flags().BoolVarP(&inplace, "inplace", "i", false, "Update the Dockerfile in place")

}
