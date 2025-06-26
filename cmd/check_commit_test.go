// cmd/check_commit_test.go

package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

// runCheckCommitTest is a helper to execute the check-commit command with mocks.
func runCheckCommitTest(t *testing.T, cfg *config.Config, gitEmail, userInput string) (exitCode int, stderr string) {
	t.Helper()

	exitCode = -1 // Default to a value that indicates it was not called.

	// Mock the exit function to capture the code instead of terminating the test.
	mockExit := func(code int) {
		exitCode = code
	}

	runner := &checkCommitRunner{
		getGitConfig: func(key string) (string, error) {
			if key == "user.email" {
				return gitEmail, nil
			}
			return "", nil
		},
		loadConfig: func() (*config.Config, error) { return cfg, nil },
		stdin:      strings.NewReader(userInput),
		stderr:     io.Discard, // Initially discard stderr for non-prompting tests
		exit:       mockExit,
	}

	// Capture stderr if we expect a prompt
	var stderrBuf bytes.Buffer
	if userInput != "" {
		runner.stderr = &stderrBuf
	}

	// We need a dummy cobra command to pass to the run function
	dummyCmd := &cobra.Command{}

	// Execute the command logic
	// We use a goroutine because the command calls exit(), which stops the goroutine,
	// but allows the test function to continue and check the captured exit code.
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// This can happen if the exit function panics, which is fine for tests.
			}
			done <- true
		}()
		runner.run(dummyCmd, []string{})
	}()
	<-done

	return exitCode, stderrBuf.String()
}

func TestCheckCommitCommand(t *testing.T) {
	// Setup a base mock config
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
		// User types "y" or just presses Enter
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "other@email.com", "\n")

		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when user aborts, but got %d", exitCode)
		}
		if !strings.Contains(stderr, "Commit aborted by user") {
			t.Errorf("Expected 'aborted' message in stderr, but it was missing. Got:\n%s", stderr)
		}
	})

	t.Run("when emails mismatch and user proceeds", func(t *testing.T) {
		// User types "n"
		exitCode, stderr := runCheckCommitTest(t, mockCfg, "other@email.com", "n\n")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 when user proceeds, but got %d", exitCode)
		}
		if !strings.Contains(stderr, "Commit proceeding with mismatched user") {
			t.Errorf("Expected 'proceeding' message in stderr, but it was missing. Got:\n%s", stderr)
		}
	})
}
