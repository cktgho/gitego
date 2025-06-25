// utils/git.go

package utils

import (
	"os/exec"
	"strings"
)

// GetEffectiveGitConfig runs 'git config <key>' without the --global flag.
// It is EXPORTED (capitalized) so it can be used by other packages (like cmd).
// This correctly resolves the config value from local > global > system.
func GetEffectiveGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
