// cmd/auto.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgreenwell/gitego/config" // IMPORTANT: Use your module path
	"github.com/spf13/cobra"
)

var autoCmd = &cobra.Command{
	Use:   "auto <path> <profile_name>",
	Short: "Automatically switch profiles based on directory.",
	Long: `Configures your global .gitconfig to automatically use a specific
profile whenever you are working inside the given directory path.

This works by creating a profile-specific gitconfig file and then using
Git's powerful 'includeIf' directive to apply it conditionally.

Example:
  gitego auto ~/work work`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		profileName := args[1]

		// --- 1. Load gitego's configuration first to check its state ---
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}
		profile, exists := cfg.Profiles[profileName]
		if !exists {
			fmt.Printf("Error: Profile '%s' not found in gitego.\n", profileName)
			return
		}

		// --- 2. Clean and prepare the directory path ---
		if strings.HasPrefix(path, "~/") {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[2:])
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("Error resolving path '%s': %v\n", path, err)
			return
		}
		cleanPath := filepath.ToSlash(absPath)
		if !strings.HasSuffix(cleanPath, "/") {
			cleanPath += "/"
		}

		// --- 3. Check if the rule already exists in gitego's config ---
		for _, rule := range cfg.AutoRules {
			if rule.Path == cleanPath && rule.Profile == profileName {
				fmt.Printf("✓ Auto-switch rule for profile '%s' on path '%s' already exists.\n", profileName, path)
				return // The rule is already fully configured, nothing to do.
			}
		}

		// --- 4. If we get here, the rule is new. Perform all actions. ---
		fmt.Printf("Setting up new auto-switch rule for profile '%s'...\n", profileName)

		// Action A: Create the dedicated gitconfig for this profile
		if err := config.EnsureProfileGitconfig(profileName, profile); err != nil {
			fmt.Printf("Error creating profile gitconfig: %v\n", err)
			return
		}

		// Action B: Add the includeIf directive to the main .gitconfig
		if err := config.AddIncludeIf(profileName, cleanPath); err != nil {
			fmt.Printf("Error updating global .gitconfig: %v\n", err)
			return
		}

		// Action C: Save the new rule to gitego's config file
		newRule := &config.AutoRule{
			Path:    cleanPath,
			Profile: profileName,
		}
		cfg.AutoRules = append(cfg.AutoRules, newRule)
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: Git config updated, but failed to save rule to gitego config: %v\n", err)
			return
		}

		fmt.Println("✓ Rule setup complete.")
	},
}

func init() {
	rootCmd.AddCommand(autoCmd)
}
