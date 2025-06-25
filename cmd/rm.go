// cmd/rm.go

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bgreenwell/gitego/config" // IMPORTANT: Use your module path
	"github.com/spf13/cobra"
)

var (
	// forceFlag will bypass the confirmation prompt for the rm command.
	forceFlag bool
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm <profile_name>",
	Short: "Removes a saved user profile.",
	Long: `Removes a profile and its associated credentials from the gitego configuration.
This is a destructive operation. By default, you will be prompted for confirmation.
Use the --force flag to bypass this prompt.`,
	Aliases: []string{"remove"},
	// Use Cobra's built-in argument validation to ensure exactly one argument is passed.
	Args: cobra.ExactArgs(1),
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

		// Add confirmation logic.
		if !forceFlag {
			fmt.Printf("Are you sure you want to remove the profile '%s'? This cannot be undone. [y/N]: ", profileName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')

			if strings.TrimSpace(strings.ToLower(response)) != "y" {
				fmt.Println("Removal cancelled.")
				return
			}
		}

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
	// Add the --force flag.
	rmCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force removal without confirmation")
	rootCmd.AddCommand(rmCmd)
}
