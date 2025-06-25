// cmd/check_commit.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bgreenwell/gitego/config"
	"github.com/bgreenwell/gitego/utils"
	"github.com/spf13/cobra"
)

var checkCommitCmd = &cobra.Command{
	Use:   "check-commit",
	Short: "Internal: checks commit author against expected profile.",
	Run: func(cmd *cobra.Command, args []string) {

		gitEmail, err := utils.GetEffectiveGitConfig("user.email")
		if err != nil {
			// Not in a git repo or no email set, nothing to check.
			os.Exit(0)
		}

		cfg, err := config.Load()
		if err != nil || len(cfg.AutoRules) == 0 {
			// No gitego config or no rules, so nothing to check.
			os.Exit(0)
		}

		// Use the new centralized function to find the expected profile.
		expectedProfileName, _ := cfg.GetActiveProfileForCurrentDir()

		if expectedProfileName == "" || expectedProfileName == cfg.ActiveProfile {
			// If no specific rule applies, or if the rule points to the default, don't warn.
			os.Exit(0)
		}

		expectedProfile, exists := cfg.Profiles[expectedProfileName]
		if !exists {
			os.Exit(0)
		}

		// If the currently configured email matches the expected profile's email, all is well.
		if gitEmail == expectedProfile.Email {
			os.Exit(0)
		}

		// --- If we reach here, there is a mismatch. Warn the user. ---

		fmt.Fprintf(os.Stderr, "\n--- gitego Safety Check ---\n")
		fmt.Fprintf(os.Stderr, "Warning: Your effective Git email for this repo is '%s'.\n", gitEmail)
		fmt.Fprintf(os.Stderr, "However, the profile expected for this directory is '%s' ('%s').\n", expectedProfileName, expectedProfile.Email)
		fmt.Fprintf(os.Stderr, "---------------------------\n")
		fmt.Fprintf(os.Stderr, "Do you want to abort the commit? [Y/n]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')

		if strings.TrimSpace(strings.ToLower(response)) == "n" {
			fmt.Fprintln(os.Stderr, "Commit proceeding with mismatched user.")
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, "Commit aborted by user.")
			os.Exit(1)
		}
	},
}

func init() {
	internalCmd.AddCommand(checkCommitCmd)
}
