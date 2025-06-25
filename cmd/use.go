// cmd/use.go
package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <profile_name>",
	Short: "Sets a profile as the active default for gitego.",
	Long: `Sets a profile as the active default. This profile will be used
for any repository that does not have a specific auto-switch rule.
This command updates your global .gitconfig, sets the active profile for the
credential helper, and preemptively updates the macOS Keychain.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		profile, exists := cfg.Profiles[profileName]
		if !exists {
			fmt.Printf("Error: Profile '%s' not found.\n", profileName)
			return
		}

		// Action 1: Set the global git config for user name and email.
		if err := setGitConfig("user.name", profile.Name); err != nil {
			fmt.Printf("Error setting git user.name: %v\n", err)
			return
		}
		if err := setGitConfig("user.email", profile.Email); err != nil {
			fmt.Printf("Error setting git user.email: %v\n", err)
			return
		}

		// Action 2: Set this profile as the active one in gitego's config.
		cfg.ActiveProfile = profileName
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving active profile setting: %v\n", err)
			return
		}

		// Action 3 (The Fix): If on macOS, also preemptively set the credential
		// in the keychain to prevent the osxkeychain helper from prompting.
		fmt.Println("[DEBUG] Checking if running on macOS...")
		if runtime.GOOS == "darwin" {
			fmt.Println("[DEBUG] macOS detected. Attempting to pre-set keychain credential.")

			token, err := config.GetToken(profileName)
			// Only proceed if the profile has a token.
			if err != nil {
				fmt.Printf("[DEBUG] No token found in gitego's vault for profile '%s'. Skipping keychain update.\n", profileName)
				// We still print the final success message because the main goal was achieved.
				fmt.Printf("✓ Set active profile to '%s'.\n", profileName)
				return
			}

			if profile.Username == "" {
				fmt.Println("[DEBUG] Profile username is empty. This is required to set the keychain entry correctly. Skipping update.")
				// We still print the final success message.
				fmt.Printf("✓ Set active profile to '%s'.\n", profileName)
				return
			}

			fmt.Printf("[DEBUG] Found token for profile '%s'. Username: '%s'. Attempting to write to keychain...\n", profileName, profile.Username)
			if err := config.SetGitCredential(profile.Username, token); err != nil {
				fmt.Printf("Warning: Failed to pre-set keychain credential: %v\n", err)
			} else {
				fmt.Println("[DEBUG] Successfully called SetGitCredential.")
			}
		}

		fmt.Printf("✓ Set active profile to '%s'.\n", profileName)
	},
}

func setGitConfig(key string, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %w\n%s", err, string(output))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(useCmd)
}
