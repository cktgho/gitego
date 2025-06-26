// config/config_test.go

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_Success tests the successful loading and parsing of a valid config file.
func TestLoad_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testConfigContent := `
profiles:
  work:
    name: Test User
    email: test@work.com
    username: testuser
    ssh_key: ~/.ssh/id_work
  personal:
    name: Test Personal
    email: test@personal.com
active_profile: personal
auto_rules:
  - path: /tmp/work/
    profile: work
`
	tempConfigFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(tempConfigFile, []byte(testConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = tempConfigFile
	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if len(cfg.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, but got %d", len(cfg.Profiles))
	}
	if profile, ok := cfg.Profiles["work"]; !ok {
		t.Error("Expected 'work' profile to exist, but it doesn't")
	} else if profile.Name != "Test User" {
		t.Errorf("Expected work profile name to be 'Test User', got '%s'", profile.Name)
	}
	if cfg.ActiveProfile != "personal" {
		t.Errorf("Expected active profile to be 'personal', got '%s'", cfg.ActiveProfile)
	}
	if len(cfg.AutoRules) != 1 {
		t.Errorf("Expected 1 auto_rule, but got %d", len(cfg.AutoRules))
	}
}

// TestLoad_NonExistentFile tests the behavior when the config file does not exist.
func TestLoad_NonExistentFile(t *testing.T) {
	// Point to a config file in a non-existent directory.
	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = "/tmp/non/existent/path/config.yaml"
	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error for a non-existent file: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned a nil config for a non-existent file.")
	}

	// Expect a new, empty config struct.
	if len(cfg.Profiles) != 0 || len(cfg.AutoRules) != 0 || cfg.ActiveProfile != "" {
		t.Error("Expected an empty config struct when file does not exist, but got data.")
	}
}

// TestLoad_EmptyFile tests the behavior with an empty config file.
func TestLoad_EmptyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an empty file.
	tempConfigFile := filepath.Join(tempDir, "config.yaml")
	if _, err := os.Create(tempConfigFile); err != nil {
		t.Fatalf("Failed to create empty config file: %v", err)
	}

	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = tempConfigFile
	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error for an empty file: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned a nil config for an empty file.")
	}

	if len(cfg.Profiles) != 0 || len(cfg.AutoRules) != 0 || cfg.ActiveProfile != "" {
		t.Error("Expected an empty config struct for an empty file, but got data.")
	}
}

// TestLoad_MalformedYAML tests that an error is returned for an invalid YAML file.
func TestLoad_MalformedYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitego-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Malformed YAML with an unclosed quote.
	malformedContent := `profiles: { work: { name: "Test User }`

	tempConfigFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(tempConfigFile, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("Failed to write malformed config file: %v", err)
	}

	originalConfigPath := gitegoConfigPath
	gitegoConfigPath = tempConfigFile
	defer func() {
		gitegoConfigPath = originalConfigPath
	}()

	_, err = Load()
	if err == nil {
		t.Error("Expected an error when loading malformed YAML, but got nil.")
	}
}
