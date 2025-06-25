// utils/git.go

package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetEffectiveGitConfig runs 'git config <key>' without the --global flag.
// This correctly resolves the config value from local > global > system.
func GetEffectiveGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", key)
	output, err := cmd.Output()
	if err != nil {
		// It's not necessarily an error if the config isn't set,
		// so we return the error but allow the caller to decide how to handle it.
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// SetGlobalGitConfig runs 'git config --global <key> <value>'.
// It sets a configuration value in the user's global .gitconfig file.
func SetGlobalGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	// We use CombinedOutput to capture any error messages from Git.
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
