// config/keyring_darwin.go

// This file will ONLY be compiled on macOS.

package config

import (
	"fmt"
	"os/exec"

	keyring_other "github.com/zalando/go-keyring"
)

const gitegoKeyringService = "gitego"

// SetToken, GetToken, and DeleteToken manage gitego's internal vault.
func SetToken(profileName, token string) error {
	return keyring_other.Set(gitegoKeyringService, profileName, token)
}
func GetToken(profileName string) (string, error) {
	return keyring_other.Get(gitegoKeyringService, profileName)
}
func DeleteToken(profileName string) error {
	return keyring_other.Delete(gitegoKeyringService, profileName)
}

// SetGitCredential directly overwrites the keychain entry that osxkeychain reads.
func SetGitCredential(username, token string) error {
	_ = exec.Command("security", "delete-internet-password", "-a", username, "-s", "github.com").Run()

	cmd := exec.Command(
		"security",
		"add-internet-password",
		"-a", username,
		"-s", "github.com",
		"-r", "htps",
		"-p", "443",
		"-w", token,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run 'security' command: %w\nOutput: %s", err, string(output))
	}

	return nil
}
