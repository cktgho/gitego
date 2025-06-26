// cmd/rm.go

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

var (
	forceFlag bool
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm <profile_name>",
	Short: "Removes a saved user profile.",
	Long: `Removes a profile, its associated credentials, and any auto-switch rules
that use it from the gitego configuration. This is a destructive operation.`,
	Aliases: []string{"remove"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		if _, exists := cfg.Profiles[profileName]; !exists {
			fmt.Printf("Error: Profile '%s' not found.\n", profileName)
			return
		}

		if !forceFlag {
			fmt.Printf("Are you sure you want to remove the profile '%s'? This cannot be undone. [y/N]: ", profileName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(response)) != "y" {
				fmt.Println("Removal cancelled.")
				return
			}
		}

		// Create a new slice to hold the rules we want to keep.
		var keptRules []*config.AutoRule
		for _, rule := range cfg.AutoRules {
			// If the rule's profile does NOT match the one being removed, keep it.
			if rule.Profile != profileName {
				keptRules = append(keptRules, rule)
			}
		}
		// Replace the old slice with the new, filtered one.
		cfg.AutoRules = keptRules

		// Delete the profile from the map.
		delete(cfg.Profiles, profileName)

		// Save the modified configuration back to the file.
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		// Also delete the associated PAT from the OS keychain.
		_ = config.DeleteToken(profileName)

		fmt.Printf("✓ Profile '%s' removed successfully.\n", profileName)
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force removal without confirmation")
	rootCmd.AddCommand(rmCmd)
}
