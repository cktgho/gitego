// cmd/rm_test.go

package cmd

import (
	"testing"

	"github.com/bgreenwell/gitego/config"
)

func TestRmCommand(t *testing.T) {
	// 1. Setup: Create mock config and state trackers
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work":     {Name: "Work User", Email: "work@example.com"},
			"personal": {Name: "Personal User", Email: "personal@example.com"},
		},
		AutoRules: []*config.AutoRule{
			{Path: "/path/to/work", Profile: "work"},
			{Path: "/path/to/personal", Profile: "personal"},
		},
	}

	var removedIncludeIf, removedProfileCfg, deletedToken string
	var saved bool

	// 2. Create a test runner with mock functions
	runner := &rmRunner{
		load: func() (*config.Config, error) {
			// Return a copy to prevent the test from modifying the original mock
			cfgCopy := *mockCfg
			return &cfgCopy, nil
		},
		save: func(c *config.Config) error {
			saved = true
			mockCfg = c // Update the "persisted" config
			return nil
		},
		removeIncludeIf: func(profileName string) error {
			removedIncludeIf = profileName
			return nil
		},
		removeProfileCfg: func(profileName string) error {
			removedProfileCfg = profileName
			return nil
		},
		deleteToken: func(profileName string) error {
			deletedToken = profileName
			return nil
		},
	}

	// 3. Execute the command to remove the "work" profile
	args := []string{"work"}
	forceFlag = true // Use force to bypass interactive prompt in test
	runner.run(rmCmd, args)
	forceFlag = false // Reset flag

	// 4. Assertions
	if _, exists := mockCfg.Profiles["work"]; exists {
		t.Error("Expected 'work' profile to be deleted from config, but it still exists.")
	}
	if len(mockCfg.Profiles) != 1 {
		t.Errorf("Expected 1 profile to remain, but found %d", len(mockCfg.Profiles))
	}

	if len(mockCfg.AutoRules) != 1 || mockCfg.AutoRules[0].Profile != "personal" {
		t.Error("Expected auto-rule for 'work' profile to be removed.")
	}

	if !saved {
		t.Error("Expected config.Save() to be called, but it wasn't.")
	}

	if removedIncludeIf != "work" {
		t.Error("Expected RemoveIncludeIf to be called for 'work' profile.")
	}

	if removedProfileCfg != "work" {
		t.Error("Expected the profile config file for 'work' to be removed.")
	}

	if deletedToken != "work" {
		t.Error("Expected DeleteToken to be called for 'work' profile.")
	}
}
