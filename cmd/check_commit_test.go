// cmd/check_commit_test.go

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

// runCheckCommitTest is a helper to execute the check-commit command with mocks.
// It now uses a channel for robust synchronization.
func runCheckCommitTest(t *testing.T, cfg *config.Config, gitEmail, userInput string) (exitCode int, stderr string) {
	t.Helper()

	exitCode = -1 // Default to a value that indicates it was not called.
	exitSignal := make(chan int, 1)

	// Mock the exit function to send the exit code to our channel.
	mockExit := func(code int) {
		exitSignal <- code
	}

	var stderrBuf bytes.Buffer

	runner := &checkCommitRunner{
		getGitConfig: func(key string) (string, error) {
			if key == "user.email" {
				return gitEmail, nil
			}
			return "", nil
		},
		loadConfig: func() (*config.Config, error) { return cfg, nil },
		stdin:      strings.NewReader(userInput),
		stderr:     &stderrBuf, // Capture stderr.
		exit:       mockExit,
	}

	// Execute the command's logic.
	runner.run(&cobra.Command{}, []string{})

	// Block until the mock exit function has been called.
	// This makes the test synchronous and reliable.
	exitCode = <-exitSignal

	return exitCode, stderrBuf.String()
}

func TestCheckCommitCommand(t *testing.T) {
	// Setup a base mock config.
	mockCfg := &config.Config{
		Profiles: map[string]*config.Profile{
			"work": {Email: "work@example.com"},
		},
		AutoRules: []*config.AutoRule{
			// The path needs to match the current dir for the test to activate the rule.
			{Path: ".", Profile: "work"},
		},
	}

	t.Run("when emails match", func(t *testing.T) {
		exitCode, _ := runCheckCommitTest(t, mockCfg, "work@example.com", "")
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for matching emails, but got %d", exitCode)
		}
	})

	t.Run("when emails mismatch and user aborts", func(t *testing.T) {
		// User types "y" or just presses Enter.
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "other@email.com", "\n")

		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when user aborts, but got %d", exitCode)
		}
		if !strings.Contains(stderr, "Commit aborted by user") {
			t.Errorf("Expected 'aborted' message in stderr, but it was missing. Got:\n%s", stderr)
		}
	})

	t.Run("when emails mismatch and user proceeds", func(t *testing.T) {
		// User types "n".
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "other@email.com", "n\n")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 when user proceeds, but got %d", exitCode)
		}
		if !strings.Contains(stderr, "Commit proceeding with mismatched user") {
			t.Errorf("Expected 'proceeding' message in stderr, but it was missing. Got:\n%s", stderr)
		}
	})
}
