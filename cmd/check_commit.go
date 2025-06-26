// cmd/check_commit.go
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bgreenwell/gitego/config"
	"github.com/bgreenwell/gitego/utils"
	"github.com/spf13/cobra"
)

// checkCommitRunner holds dependencies for mocking.
type checkCommitRunner struct {
	getGitConfig func(string) (string, error)
	loadConfig   func() (*config.Config, error)
	stdin        io.Reader
	stderr       io.Writer
	exit         func(int)
}

// run is the core logic for the check-commit command.
func (r *checkCommitRunner) run(cmd *cobra.Command, args []string) {
	gitEmail, err := r.getGitConfig("user.email")
	if err != nil {
		r.exit(0)
		return
	}

	cfg, err := r.loadConfig()
	if err != nil || len(cfg.AutoRules) == 0 {
		r.exit(0)
		return
	}

	expectedProfileName, _ := cfg.GetActiveProfileForCurrentDir()

	// No specific rule applies, so no check is needed.
	if expectedProfileName == "" || expectedProfileName == cfg.ActiveProfile {
		r.exit(0)
		return
	}

	expectedProfile, exists := cfg.Profiles[expectedProfileName]
	if !exists {
		r.exit(0) // Rule points to a non-existent profile, let validation handle warnings.
		return
	}

	// If emails match, everything is correct.
	if gitEmail == expectedProfile.Email {
		r.exit(0)
		return
	}

	// --- Mismatch found, prompt the user ---
	fmt.Fprintf(r.stderr, "\n--- gitego Safety Check ---\n")
	fmt.Fprintf(r.stderr, "Warning: Your effective Git email for this repo is '%s'.\n", gitEmail)
	fmt.Fprintf(r.stderr, "However, the profile expected for this directory is '%s' ('%s').\n", expectedProfileName, expectedProfile.Email)
	fmt.Fprintf(r.stderr, "---------------------------\n")
	fmt.Fprintf(r.stderr, "Do you want to abort the commit? [Y/n]: ")

	reader := bufio.NewReader(r.stdin)
	response, _ := reader.ReadString('\n')

	if strings.TrimSpace(strings.ToLower(response)) == "n" {
		fmt.Fprintln(r.stderr, "Commit proceeding with mismatched user.")
		r.exit(0)
	} else {
		fmt.Fprintln(r.stderr, "Commit aborted by user.")
		r.exit(1)
	}
}

var checkCommitCmd = &cobra.Command{
	Use:    "check-commit",
	Short:  "Internal: checks commit author against expected profile.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		runner := &checkCommitRunner{
			getGitConfig: utils.GetEffectiveGitConfig,
			loadConfig:   config.Load,
			stdin:        os.Stdin,
			stderr:       os.Stderr,
			exit:         os.Exit,
		}
		runner.run(cmd, args)
	},
}

func init() {
	internalCmd.AddCommand(checkCommitCmd)
}
