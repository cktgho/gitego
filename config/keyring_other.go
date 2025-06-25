//go:build !darwin

// config/keyring.go
package config

import "github.com/zalando/go-keyring"

// Use a constant for the service name to avoid typos.
// This is the name under which gitego stores its own library of PATs.
const gitegoKeyringService = "gitego"

// SetToken securely stores a PAT for a given profile name.
func SetToken(profileName, token string) error {
	return keyring.Set(gitegoKeyringService, profileName, token)
}

// GetToken securely retrieves a PAT for a given profile name.
func GetToken(profileName string) (string, error) {
	return keyring.Get(gitegoKeyringService, profileName)
}

// DeleteToken securely removes a PAT for a given profile name.
func DeleteToken(profileName string) error {
	return keyring.Delete(gitegoKeyringService, profileName)
}

// SetGitCredential sets the PAT in the location that Git's helper will read from.
// This is the "active slot" for github.com.
func SetGitCredential(username string, token string) error {
	// The username parameter is required to match the function signature,
	// but this library uses the full URL as the user key for Git credentials.
	// The line below tells the Go compiler that we are intentionally not using it.
	_ = username

	// The service "git" and user "https://github.com" are the standard keys.
	return keyring.Set("git", "https://github.com", token)
}
