package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
		var paths []string

		if recurse {
			for _, path := range args {
				err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
					if err != nil {
						log.Fatal(err)
						return err
					}
					if info.IsDir() {
						paths = append(paths, p)
					}
					return nil
				})
				if err != nil {
					fmt.Println(err)
				}
			}
		} else {
			paths = args
		}

		updates := checker.CheckForUpdates(paths, includePrerelease, sshPrivKeyPath, sshPrivKeyPwd)
		for _, update := range updates {
			fmt.Printf("%#v\n", update)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var includePrerelease bool
var recurse bool
var sshPrivKeyPath string
var sshPrivKeyPwd string

func init() {
	rootCmd.Flags().BoolVarP(&includePrerelease, "include-prerelease", "p", false, "Include prerelease versions")
	rootCmd.Flags().BoolVarP(&recurse, "recurse", "r", false, "Recurse into all sub-directories")
	rootCmd.Flags().StringVarP(&sshPrivKeyPath, "ssh-private-key-path", "s", "", "Specify a private key to use when cloning via SSH")
	rootCmd.Flags().StringVarP(&sshPrivKeyPwd, "ssh-private-key-pwd", "w", "", "Specify a password for the private key if required")
}
