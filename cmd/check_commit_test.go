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
func runCheckCommitTest(t *testing.T, cfg *config.Config, gitEmail, userInput string) (exitCode int, stderr string) {
	t.Helper()

	exitCode = -1 // Default to a value that indicates it was not called.

	// Mock the exit function. Instead of calling os.Exit, we panic.
	// This immediately stops the execution flow, just like os.Exit would.
	mockExit := func(code int) {
		exitCode = code
		panic("os.Exit called")
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
		stderr:     &stderrBuf, // Always capture stderr
		exit:       mockExit,
	}

	// We use defer to recover from the panic we triggered in our mock exit function.
	// This allows the test to continue and check the results.
	defer func() {
		if r := recover(); r != nil {
			if r != "os.Exit called" {
				// If it's a different panic, we should re-throw it.
				panic(r)
			}
		}
	}()

	// Execute the command's logic synchronously.
	runner.run(&cobra.Command{}, []string{})

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
