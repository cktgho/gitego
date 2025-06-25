// cmd/rm.go

package cmd

import (
	"fmt"

	"github.com/bgreenwell/gitego/config" // IMPORTANT: Use your module path
	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm <profile_name>",
	Short: "Removes a saved user profile.",
	Long: `Removes a profile from the gitego configuration file.
The profile name is case-sensitive.`,
	Aliases: []string{"remove"},
	// Use Cobra's built-in argument validation to ensure exactly one argument is passed.
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		// Load the existing configuration.
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		// Check if the profile to be removed actually exists.
		if _, exists := cfg.Profiles[profileName]; !exists {
			fmt.Printf("Error: Profile '%s' not found.\n", profileName)
			return
		}

		// Delete the profile from the map; this is a built-in Go function.
		delete(cfg.Profiles, profileName)

		// Save the modified configuration back to the file.
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		// Also delete the associated PAT from the OS keychain.
		// We ignore the error here, as the token may not have existed.
		_ = config.DeleteToken(profileName)

		fmt.Printf("✓ Profile '%s' removed successfully.\n", profileName)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
