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

var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin all third party Docker images used in your Dockerfile",
	Long: `
	Pin all third party Docker images used in your Dockerfile

	This command will look for all the occurences of FROM key in your Dockerfile
	and update them to use the digest of the image instead of the tag.

	You can expet to see output with substitutions like this:
	alpine:3.18 -> alpine@sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978

	Pin command looks for Dockerfile in the current directory by default. You can
	use -f flag to specify a different Dockerfile.

	By default a <Dockerfile>.pinned file is created. You can use -o flag to
	specify a different output file.

	To update the Dockerfile in place, use -i flag:
	|> pinny docker pin -i

	Dockerfile - before
	| FROM golang:alpine AS builder
	| WORKDIR /app
	| COPY . .
	| RUN go build -o myapp
	|
	| FROM alpine:latest
	| WORKDIR /app
	| COPY --from=builder /app/myapp .
	| EXPOSE 8080
	| CMD ["./myapp"]
	
	Dockerfile - after
	| FROM golang@sha256:110b07af87238fbdc5f1df52b00927cf58ce3de358eeeb1854f10a8b5e5e1411 AS builder
	| WORKDIR /app
	| COPY . .
	| RUN go build -o myapp
	|
	| FROM alpine@sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978
	| WORKDIR /app
	| COPY --from=builder /app/myapp .
	| EXPOSE 8080
	| CMD ["./myapp"]

`,
	Run: func(cmd *cobra.Command, args []string) {
		offline := false
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
	pinCmd.Flags().BoolVarP(&inplace, "inplace", "i", false, "Update the Dockerfile in place")
	pinCmd.Flags().StringVarP(&dockerfile, "dockerfile", "f", "Dockerfile", "Dockerfile to use")
}
