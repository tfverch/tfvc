package cmd

import (
	"os"

	"github.com/ryan-jan/tfvc/internal/checker"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tfvc",
	Short: "A brief description of your application",
	Long: `A longer description
`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			updates := checker.CheckForUpdates(path, includePrerelease, sshPrivKeyPath, sshPrivKeyPwd)
			for _, update := range updates {
				update.DefaultOutput()
			}
			// TODO: write code to update readme with table
			// if updateReadme {
			// 	if err := updates.Format(os.Stdout, "markdown"); err != nil {
			// 		log.Fatal(err)
			// 	}
			// }
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var updateReadme bool
var readmePath string
var includePrerelease bool
var sshPrivKeyPath string
var sshPrivKeyPwd string

func init() {
	// rootCmd.Flags().BoolVarP(&updateReadme, "update-readme", "r", false, "Update the README with a markdown table listing all versions")
	// rootCmd.Flags().StringVarP(&readmePath, "readme-path", "a", "./README.md", "Specify the path to a markdown file to update")
	rootCmd.Flags().BoolVarP(&includePrerelease, "include-prerelease", "p", false, "Include prerelease versions")
	rootCmd.Flags().StringVarP(&sshPrivKeyPath, "ssh-private-key-path", "s", "", "Specify a private key to use when cloning via SSH")
	rootCmd.Flags().StringVarP(&sshPrivKeyPwd, "ssh-private-key-pwd", "w", "", "Specify a password for the private key if required")
}
