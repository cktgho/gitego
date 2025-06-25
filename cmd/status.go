// cmd/status.go

package cmd

import (
	"fmt"

	"github.com/bgreenwell/gitego/config"
	"github.com/bgreenwell/gitego/utils"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Displays the current effective Git user and any active gitego rule.",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use our shared helper function.
		name, _ := utils.GetEffectiveGitConfig("user.name")
		email, err := utils.GetEffectiveGitConfig("user.email")
		if err != nil {
			fmt.Println("Not inside a Git repository or user not configured.")
			return
		}

		// Load gitego config to check for rules.
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Warning: Could not load gitego config: %v\n", err)
		}

		// Determine the source of the configuration.
		var source string
		if cfg != nil {
			_, ruleSource := cfg.GetActiveProfileForCurrentDir()
			if ruleSource != "No active gitego profile" {
				source = ruleSource
			} else {
				source = "Global Git Config"
			}
		} else {
			source = "Global Git Config"
		}

		fmt.Println("--- Git Identity Status ---")
		fmt.Printf("  Name:   %s\n", name)
		fmt.Printf("  Email:  %s\n", email)
		fmt.Printf("  Source: %s\n", source)
		fmt.Println("---------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
