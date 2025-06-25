// cmd/root.go

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	// This is the action that will run if no sub-commands are specified.
	// For now, we will just print the help information.
	Run: func(cmd *cobra.Command, args []string) {
		// If the user just types "gitego", show them the help.
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// os.Exit finishes the program with a given status code.
	// If Execute() returns an error, it will print the error to stderr
	// and the program will exit with a status code of 1.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}

