// cmd/add.go

package cmd

import (
	"errors"
	"fmt"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

var (
	// These variables will hold the values from the --name and --email flags.
	addName     string
	addEmail    string
	addUsername string
	addSSHKey   string
	addPAT      string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <profile_name>",
	Short: "Adds a new user profile to the gitego config.",
	Long: `Adds a new user profile, associating a profile name (e.g., "work")
with a specific Git user name and email address.`,
	// We expect exactly one argument: the name for the new profile.
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires exactly one argument: the profile name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		// Load the existing configuration.
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		// Check if a profile with this name already exists.
		if _, exists := cfg.Profiles[profileName]; exists {
			fmt.Printf("Error: Profile '%s' already exists.\n", profileName)
			fmt.Printf("Use 'gitego edit %s' to modify it, or 'gitego rm %s' to remove it.\n", profileName, profileName)
			return
		}

		// Create a new Profile struct from the flags.
		newProfile := &config.Profile{
			Name:     addName,
			Email:    addEmail,
			Username: addUsername,
			SSHKey:   addSSHKey,
		}

		// Add the new profile to our config map.
		cfg.Profiles[profileName] = newProfile

		// Save the updated configuration back to the file.
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		// If a PAT was provided, store it securely in the OS keychain.
		if addPAT != "" {
			if err := config.SetToken(profileName, addPAT); err != nil {
				fmt.Printf("Warning: Profile saved, but failed to store PAT securely: %v\n", err)
				return
			}
		}

		fmt.Printf("✓ Profile '%s' added successfully.\n", profileName)
	},
}

func init() {
	// Add the 'add' command to our root command.
	rootCmd.AddCommand(addCmd)

	// Define the --name and --email flags for the 'add' command andbind them
	// to the variables we declared at the top.
	addCmd.Flags().StringVarP(&addName, "name", "n", "", "The user.name for the profile")
	addCmd.Flags().StringVarP(&addEmail, "email", "e", "", "The user.email for the profile")
	addCmd.Flags().StringVar(&addUsername, "username", "", "Login username for the service (e.g., GitHub username)")
	addCmd.Flags().StringVar(&addSSHKey, "ssh-key", "", "Path to the SSH key for this profile (optional)")
	addCmd.Flags().StringVar(&addPAT, "pat", "", "Personal Access Token for this profile (stored securely)")

	// Mark the flags as required, so Cobra will return an error if they are not provided.
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("email")
}
