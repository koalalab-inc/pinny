/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package docker

import (
	"github.com/koalalab-inc/pinny/pkg/docker"
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Generate a lock file for the docker images referenced in Dockerfile",
	Long: `
	Generate a lock file for the docker images referenced in Dockerfile.
	A pinny-lock.json file will be generated in the current directory.

	Example:
		pinny docker lock
		pinny docker lock [-f Dockerfile]
		pinny docker lock -f DevDockerfile

	A sample pinny-lock.json file looks like this:
	Dockerfile:
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
	|
	Generated pinny-lock.json:
	| {
	|	"docker://docker.io/library/alpine:latest": "sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978",
	|	"docker://docker.io/library/golang:alpine": "sha256:110b07af87238fbdc5f1df52b00927cf58ce3de358eeeb1854f10a8b5e5e1411",
	|	"generated_at": "Tue, 28 Nov 2023 13:25:30 IST",
	|	"generated_by": "Pinny"
	| }
	
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := docker.GeneratePinnyLockFile(dockerfile)
		cobra.CheckErr(err)
	},
}

func init() {
	lockCmd.Flags().StringVarP(&dockerfile, "dockerfile", "f", "Dockerfile", "Dockerfile to use")
}
