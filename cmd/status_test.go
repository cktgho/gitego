// cmd/status_test.go

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bgreenwell/gitego/config"
)

func TestStatusCommand(t *testing.T) {
	// 1. Setup a temporary directory structure for our test
	tempDir, err := os.MkdirTemp("", "gitego-status-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workDir := filepath.Join(tempDir, "work", "project")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work dir: %v", err)
	}

	// Save original working directory to restore later
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// 2. Setup mock config
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {
				Name:  "Work User",
				Email: "work@example.com",
			},
			"global": {
				Name:  "Global User",
				Email: "global@example.com",
			},
		},
		AutoRules: []*config.AutoRule{
			// FIX: Add a trailing slash to the path to ensure correct prefix matching.
			{Path: filepath.Join(tempDir, "work") + string(os.PathSeparator), Profile: "work"},
		},
		ActiveProfile: "global",
	}

	// 3. Create a test runner with mocked dependencies
	runner := &statusRunner{
		load: func() (*config.Config, error) {
			return mockCfg, nil
		},
		getGitConfig: func(key string) (string, error) { return "", nil },
	}

	// --- Scenario 1: Test inside the auto-rule directory ---
	t.Run("inside auto-rule directory", func(t *testing.T) {
		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("Failed to change directory to workDir: %v", err)
		}

		runner.getGitConfig = func(key string) (string, error) {
			if key == "user.name" {
				return "Work User", nil
			}
			return "work@example.com", nil
		}

		var buf bytes.Buffer
		statusCmd.SetOut(&buf)
		runner.run(statusCmd, []string{})
		output := buf.String()

		expectedSource := "gitego auto-rule for profile 'work'"
		if !strings.Contains(output, expectedSource) {
			t.Errorf("Expected output to contain source '%s', but it didn't.\nOutput:\n%s", expectedSource, output)
		}
		if !strings.Contains(output, "Work User") {
			t.Errorf("Expected output to contain name 'Work User', but it didn't.\nOutput:\n%s", output)
		}
	})

	// --- Scenario 2: Test outside any auto-rule directory ---
	t.Run("outside auto-rule directory", func(t *testing.T) {
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change directory to tempDir: %v", err)
		}

		runner.getGitConfig = func(key string) (string, error) {
			if key == "user.name" {
				return "Global User", nil
			}
			return "global@example.com", nil
		}

		var buf bytes.Buffer
		statusCmd.SetOut(&buf)
		runner.run(statusCmd, []string{})
		output := buf.String()

		// The expected source is "Global gitego default" because an active_profile is set.
		expectedSource := "Global gitego default"
		if !strings.Contains(output, expectedSource) {
			t.Errorf("Expected output to contain source '%s', but it didn't.\nOutput:\n%s", expectedSource, output)
		}
		if !strings.Contains(output, "Global User") {
			t.Errorf("Expected output to contain name 'Global User', but it didn't.\nOutput:\n%s", output)
		}
	})
}
