// cmd/rm.go

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
	Short: "Removes a saved user profile and all associated rules.",
	Long: `Removes a profile, its associated credentials, any auto-switch rules
that use it from the gitego config, and cleans up corresponding rules
from your global .gitconfig file.`,
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
			fmt.Printf("Are you sure you want to remove the profile '%s' and all its rules? This cannot be undone. [y/N]: ", profileName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(response)) != "y" {
				fmt.Println("Removal cancelled.")
				return
			}
		}

		// 1. Remove the includeIf directive from the global .gitconfig.
		if err := config.RemoveIncludeIf(profileName); err != nil {
			fmt.Printf("Warning: Failed to remove rule from .gitconfig: %v\n", err)
		}

		// 2. Delete the profile-specific .gitconfig file.
		home, _ := os.UserHomeDir()
		profileGitconfigPath := filepath.Join(home, ".gitego", "profiles", fmt.Sprintf("%s.gitconfig", profileName))
		_ = os.Remove(profileGitconfigPath) // Ignore error if file doesn't exist.

		// 3. Remove any auto-rules from gitego's config that use this profile.
		var keptRules []*config.AutoRule
		for _, rule := range cfg.AutoRules {
			if rule.Profile != profileName {
				keptRules = append(keptRules, rule)
			}
		}
		cfg.AutoRules = keptRules

		// 4. Delete the profile itself.
		delete(cfg.Profiles, profileName)

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		// 5. Remove the PAT from the OS keychain.
		_ = config.DeleteToken(profileName)

		fmt.Printf("✓ Profile '%s' and all associated rules removed successfully.\n", profileName)
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force removal without confirmation")
	rootCmd.AddCommand(rmCmd)
}
