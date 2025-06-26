// cmd/edit_test.go

package cmd

import (
	"testing"

	"github.com/bgreenwell/gitego/config"
)

func TestEditCommand(t *testing.T) {
	// 1. Setup: Create an initial mock config
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {
				Name:     "Original Name",
				Email:    "original@example.com",
				Username: "original_user",
			},
		},
	}
	var patSetForProfile string
	var patValue string

	// 2. Create the test runner with mocked dependencies
	runner := &editor{
		load: func() (*config.Config, error) {
			// Return a copy to ensure the original mock isn't mutated directly
			cfgCopy := *mockCfg
			return &cfgCopy, nil
		},
		save: func(c *config.Config) error {
			mockCfg = c // Update our "persisted" config
			return nil
		},
		setToken: func(profileName, token string) error {
			patSetForProfile = profileName
			patValue = token
			return nil
		},
	}

	// 3. Execute the command's logic
	args := []string{"work"}

	// Simulate the user providing only the --email and --pat flags
	editCmd.Flags().Set("email", "new-email@example.com")
	editCmd.Flags().Set("pat", "new-pat-123")

	runner.run(editCmd, args)

	// Reset flags after run to avoid affecting other tests
	defer editCmd.Flags().Set("email", "")
	defer editCmd.Flags().Set("pat", "")

	// 4. Assertions
	updatedProfile, ok := mockCfg.Profiles["work"]
	if !ok {
		t.Fatal("The 'work' profile was unexpectedly deleted.")
	}

	// This field should be changed
	if updatedProfile.Email != "new-email@example.com" {
		t.Errorf("Expected email to be updated to 'new-email@example.com', got '%s'", updatedProfile.Email)
	}

	// These fields should NOT have changed
	if updatedProfile.Name != "Original Name" {
		t.Errorf("Expected name to remain 'Original Name', but it was changed to '%s'", updatedProfile.Name)
	}
	if updatedProfile.Username != "original_user" {
		t.Errorf("Expected username to remain 'original_user', but it was changed to '%s'", updatedProfile.Username)
	}

	// Check if the PAT was set correctly
	if patSetForProfile != "work" || patValue != "new-pat-123" {
		t.Error("Expected SetToken to be called with the new PAT for the 'work' profile.")
	}
}
