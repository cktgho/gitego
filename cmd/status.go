// cmd/status.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgreenwell/gitego/config"
	"github.com/bgreenwell/gitego/utils" // <-- NEW: Import our utils package
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Displays the current effective Git user and any active gitego rule.",
	Long:  `...`, // (omitting for brevity, no changes here)
	Run: func(cmd *cobra.Command, args []string) {
		// Use our new, shared helper function!
		name, _ := utils.GetEffectiveGitConfig("user.name")
		email, err := utils.GetEffectiveGitConfig("user.email")
		if err != nil {
			fmt.Println("Not inside a Git repository or user not configured.")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Warning: Could not load gitego config: %v\n", err)
		}

		source := "Global Git Config"
		if cfg != nil && len(cfg.AutoRules) > 0 {
			currentDir, _ := os.Getwd()
			currentAbsDir, _ := filepath.Abs(currentDir)
			for _, rule := range cfg.AutoRules {
				rulePath, _ := filepath.Abs(strings.TrimSuffix(rule.Path, "/"))
				if strings.HasPrefix(currentAbsDir, rulePath) {
					source = fmt.Sprintf("gitego auto-rule for profile '%s' (path: %s)", rule.Profile, rule.Path)
					break
				}
			}
		}

		fmt.Println("--- Git Identity Status ---")
		fmt.Printf("  Name:   %s\n", name)
		fmt.Printf("  Email:  %s\n", email)
		fmt.Printf("  Source: %s\n", source)
		fmt.Println("---------------------------")
	},
}

// The local getEffectiveGitConfig function has been REMOVED from this file.

func init() {
	rootCmd.AddCommand(statusCmd)
}
