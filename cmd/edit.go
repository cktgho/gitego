// cmd/edit.go

package cmd

import (
	"fmt"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

var (
	// These variables will hold the values from the flags for the 'edit' command.
	editName     string
	editEmail    string
	editUsername string
	editSSHKey   string
	editPAT      string
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit <profile_name>",
	Short: "Edits an existing user profile.",
	Long: `Edits an existing user profile. You can update the user name, email,
username, SSH key, or Personal Access Token (PAT).
Only the flags you provide will be updated.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		profile, exists := cfg.Profiles[profileName]
		if !exists {
			fmt.Printf("Error: Profile '%s' not found.\n", profileName)
			return
		}

		// Update fields only if the corresponding flag was set by the user.
		if cmd.Flags().Changed("name") {
			profile.Name = editName
		}
		if cmd.Flags().Changed("email") {
			profile.Email = editEmail
		}
		if cmd.Flags().Changed("username") {
			profile.Username = editUsername
		}
		if cmd.Flags().Changed("ssh-key") {
			profile.SSHKey = editSSHKey
		}

		// Save the updated configuration.
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		// If a new PAT was provided, update it in the secure keychain.
		if cmd.Flags().Changed("pat") {
			if err := config.SetToken(profileName, editPAT); err != nil {
				fmt.Printf("Warning: Profile updated, but failed to store new PAT securely: %v\n", err)
				return
			}
		}

		fmt.Printf("✓ Profile '%s' updated successfully.\n", profileName)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Define the flags for the 'edit' command. They are not marked as required,
	// as the user might only want to update one or two fields.
	editCmd.Flags().StringVarP(&editName, "name", "n", "", "The new user.name for the profile")
	editCmd.Flags().StringVarP(&editEmail, "email", "e", "", "The new user.email for the profile")
	editCmd.Flags().StringVar(&editUsername, "username", "", "The new login username for the service")
	editCmd.Flags().StringVar(&editSSHKey, "ssh-key", "", "The new path to the SSH key for this profile")
	editCmd.Flags().StringVar(&editPAT, "pat", "", "The new Personal Access Token for this profile")
}
