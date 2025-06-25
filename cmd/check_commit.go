// cmd/check_commit.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
			os.Exit(0)
		}

		cfg, err := config.Load()
		if err != nil || len(cfg.AutoRules) == 0 {
			os.Exit(0)
		}

		currentDir, _ := os.Getwd()
		var expectedProfileName string
		for _, rule := range cfg.AutoRules {
			rulePath, _ := filepath.Abs(strings.TrimSuffix(rule.Path, "/"))
			currentAbsDir, _ := filepath.Abs(currentDir)
			if strings.HasPrefix(currentAbsDir, rulePath) {
				expectedProfileName = rule.Profile
				break
			}
		}

		if expectedProfileName == "" {
			os.Exit(0)
		}

		expectedProfile, exists := cfg.Profiles[expectedProfileName]
		if !exists {
			os.Exit(0)
		}

		if gitEmail == expectedProfile.Email {
			os.Exit(0)
		}

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

// The local getEffectiveGitConfig function has been REMOVED from this file.

func init() {
	internalCmd.AddCommand(checkCommitCmd)
}
