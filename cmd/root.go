// cmd/root.go

// Package cmd provides the root command for the gitego CLI application.
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// The version of the application.
var version = "dev" // "dev" is a good default for local builds

var (
	// versionFlag is a flag to print the version and exit.
	versionFlag bool
)

// rootCmd represents the base command when called without any subcommands.
// It's the main entry point for the CLI application.
var rootCmd = &cobra.Command{
	Use:   "gitego",
	Short: "A clever, context-aware identity manager for Git.",
	Long: `gitego is a command-line tool to seamlessly manage your Git "alter egos".

It allows you to define, switch between, and automatically apply different
user profiles (user.name, user.email), SSH keys, and Personal Access Tokens
depending on your current working directory or other contexts.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// If the version flag is passed, print the version and exit.
		if versionFlag {
			fmt.Printf("gitego version %s\n", version)

			return
		}
		// Otherwise, show the help information.
		if err := cmd.Help(); err != nil {
			log.Fatalf("Failed to show help: %v", err)
		}
	},
}

func init() {
	// Add the --version flag to the root command.
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print gitego's version number")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}
