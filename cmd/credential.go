// cmd/credential.go
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

var credentialCmd = &cobra.Command{
	Use:    "credential",
	Short:  "Internal: A Git credential helper.",
	Hidden: true, // Hide this from the standard help command.
	Run: func(cmd *cobra.Command, args []string) {
		// This command is called by Git. We need to determine the correct
		// profile and return its credentials.

		// A credential helper must read the input Git sends it on stdin.
		// We don't need to use the input for our logic, but we must consume it
		// to correctly fulfill the credential helper protocol.
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// We can just ignore the lines for now.
		}

		cfg, err := config.Load()
		if err != nil {
			// If we can't load config, we can't do anything. Exit silently.
			return
		}

		var activeProfileName string

		// 1. Check for an auto-rule that matches the current directory.
		currentDir, _ := os.Getwd()
		currentAbsDir, _ := filepath.Abs(currentDir)
		for _, rule := range cfg.AutoRules {
			rulePath, _ := filepath.Abs(strings.TrimSuffix(rule.Path, "/"))
			if strings.HasPrefix(currentAbsDir, rulePath) {
				activeProfileName = rule.Profile
				break
			}
		}

		// 2. If no auto-rule matched, fall back to the manually set active profile.
		if activeProfileName == "" {
			activeProfileName = cfg.ActiveProfile
		}

		// 3. If no profile is active in any way, exit.
		if activeProfileName == "" {
			return
		}

		// 4. We have a profile! Let's get its details.
		profile, exists := cfg.Profiles[activeProfileName]
		if !exists {
			return // Active profile doesn't actually exist.
		}

		// 5. Retrieve the PAT from the secure OS keychain.
		token, err := config.GetToken(activeProfileName)
		if err != nil || token == "" {
			return // No PAT stored for this profile.
		}

		// 6. Print the credentials to stdout in the format Git expects.
		// Git will read this output to get the username and password.
		fmt.Printf("username=%s\n", profile.Username)
		fmt.Printf("password=%s\n", token)
	},
}

func init() {
	rootCmd.AddCommand(credentialCmd)
}
