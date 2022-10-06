package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tfverch/tfvc/internal/checker"
)

var rootCmd = &cobra.Command{
	Use:   "tfvc",
	Short: "tfvc is a tool for checking terraform provider and module versions are up to date",
	Long: `A longer description
`,
	Run: func(cmd *cobra.Command, args []string) {
		exitStatus := 0
		for _, path := range args {
			updates, err := checker.Main(path, includePrerelease, sshPrivKeyPath, sshPrivKeyPwd)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				os.Exit(1)
			}
			sort.Slice(updates, func(i, j int) bool {
				return updates[i].StatusInt < updates[j].StatusInt
			})
			for _, update := range updates {
				if includePassed || update.Status != "PASSED" {
					update.DefaultOutput()
				}
			}
			// Need to write code to update readme with table
			// if updateReadme {
			// 	if err := updates.Format(os.Stdout, "markdown"); err != nil {
			// 		log.Fatal(err)
			// 	}
			// }
			if len(updates) > 0 {
				highestStatus := updates[len(updates)-1].StatusInt
				if highestStatus > exitStatus {
					exitStatus = highestStatus
				}
			}
		}
		os.Exit(exitStatus)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// var updateReadme bool //nolint:all.
// var readmePath string //nolint:all.
var includePassed bool
var includePrerelease bool
var sshPrivKeyPath string
var sshPrivKeyPwd string

func init() {
	// rootCmd.Flags().BoolVarP(&updateReadme, "update-readme", "r", false, "Update the README with a markdown table listing all versions") //nolint:all.
	// rootCmd.Flags().StringVarP(&readmePath, "readme-path", "p", "./README.md", "Specify the path to a markdown file to update") //nolint:all.
	rootCmd.Flags().BoolVarP(&includePassed, "include-passed", "a", false, "Include passed checks in console output")
	rootCmd.Flags().BoolVarP(&includePrerelease, "include-prerelease", "e", false, "Include prerelease versions")
	rootCmd.Flags().StringVarP(&sshPrivKeyPath, "ssh-private-key-path", "s", "", "Specify a private key to use when cloning via SSH")
	rootCmd.Flags().StringVarP(&sshPrivKeyPwd, "ssh-private-key-pwd", "w", "", "Specify a password for the private key if required")
}
