/*
Copyright © 2023 Koalalab Inc <dev@koalalab.com>
*/
package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/koalalab-inc/pinny/pkg/actions"
	"github.com/spf13/cobra"
)

var dryRun bool

var workflowDir = ".github/workflows"

var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin all third party Github Actions used in your workflows",
	Long: `
	Pin all third party Github Actions used in your workflows. This will
	update the workflow files in your .github/workflows directory to use
	the digest of the action instead of the ref.

	This command will look for all the occurences of uses key in your workflow
	files and update them to use the digest of the action instead of the ref.

	You can expet to see output with substitutions like this:
	actions/checkout@v3.1.0 -> actions/checkout@93ea575cb5d8a053eaa0ac8fa3b40d7e05a33cc8

	Workflow files are updated in place. 
	e.g.:

	workflow.yaml - before
	| name: Build
	| on:
	|   push:
	|     branches:
	|       - main
	|   pull_request:
	|     branches:
	|       - main
	| 
	| jobs:
	|   build:
	|     runs-on: ubuntu-latest
	|     steps:
	|       - name: Checkout code
	|         uses: actions/checkout@v3.1.0
	| 
	|       - name: Set up Go
	|         uses: actions/setup-go@v2
	|         with:
	|           go-version: 1.17

	workflow.yaml - after
	| name: Build
	| on:
	|   push:
	|     branches:
	|       - main
	|   pull_request:
	|     branches:
	|       - main
	| 
	| jobs:
	|   build:
	|     runs-on: ubuntu-latest
	|     steps:
	|       - name: Checkout code
	|         uses: actions/checkout@93ea575cb5d8a053eaa0ac8fa3b40d7e05a33cc8 # actions/checkout@v3.1.0
	| 
	|       - name: Set up Go
	|         uses: actions/setup-go@bfdd3570ce990073878bf10f6b2d79082de49492 # actions/setup-go@v2 | v2.2.0
	|         with:
	|           go-version: 1.17

`,
	Run: func(cmd *cobra.Command, args []string) {
		err := PinWorkflows(cmd)
		cobra.CheckErr(err)
	},
}

func init() {
	pinCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Print the changes without updating the workflow files")
}

func PinWorkflows(cmd *cobra.Command) error {
	workflows, err := os.ReadDir(workflowDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("no workflows found in %s", workflowDir)
	} else if err != nil {
		return err
	}

	errFlag := false

	for _, workflow := range workflows {
		workflowName := workflow.Name()
		isYAML := strings.HasSuffix(workflowName, ".yml") || strings.HasSuffix(workflowName, ".yaml")
		if !workflow.IsDir() && isYAML {
			err = actions.PinWorkflow(workflow.Name())
			if err != nil {
				errFlag = true
				break
			}
		}
	}
	for _, workflow := range workflows {
		srcFile := fmt.Sprintf("%s/%s", workflowDir, workflow.Name())
		tmpFile := fmt.Sprintf("%s/%s.tmp", workflowDir, workflow.Name())
		if errFlag {
			os.Remove(tmpFile)
		} else {
			if !workflow.IsDir() {
				if dryRun {
					cmd.Printf("Pinned %s\n", srcFile)
					file, err := os.ReadFile(tmpFile)
					if err == nil {
						cmd.Println(string(file) + "\n")
					}
					os.Remove(tmpFile)
				} else {
					os.Rename(tmpFile, srcFile)
				}
			}
		}
	}
	return err
}
