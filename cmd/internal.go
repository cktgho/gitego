// cmd/internal.go
package cmd

import "github.com/spf13/cobra"

// internalCmd is a hidden command that groups all internal subcommands.
var internalCmd = &cobra.Command{
	Use:    "internal",
	Short:  "Internal commands for gitego.",
	Hidden: true, // This hides it from the main help output.
}

func init() {
	rootCmd.AddCommand(internalCmd)
}
