// cmd/install_hook_test.go

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestGitRepo creates a temporary directory structure that mimics a Git repo.
func setupTestGitRepo(t *testing.T) (repoRoot string, hooksDir string) {
	t.Helper() // Marks this as a test helper function.

	repoRoot, err := os.MkdirTemp("", "gitego-testhook-")
	if err != nil {
		t.Fatalf("Failed to create temp repo root: %v", err)
	}

	hooksDir = filepath.Join(repoRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create temp hooks dir: %v", err)
	}

	return repoRoot, hooksDir
}

// captureOutput captures stdout and stdin for a given function call.
func captureOutput(t *testing.T, stdinContent string, action func()) string {
	t.Helper()

	originalStdout := os.Stdout
	originalStdin := os.Stdin
	defer func() {
		os.Stdout = originalStdout
		os.Stdin = originalStdin
	}()

	// Mock stdin
	r, w, _ := os.Pipe()
	if stdinContent != "" {
		w.WriteString(stdinContent)
	}
	w.Close()
	os.Stdin = r

	// Capture stdout
	readOut, writeOut, _ := os.Pipe()
	os.Stdout = writeOut

	action()

	writeOut.Close()
	var buf bytes.Buffer
	io.Copy(&buf, readOut)

	return buf.String()
}

func TestInstallHook(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	t.Run("when no hook exists", func(t *testing.T) {
		repoRoot, hooksDir := setupTestGitRepo(t)
		defer os.RemoveAll(repoRoot)

		// Change into the repo directory for the test and ensure we change back.
		os.Chdir(repoRoot)
		defer os.Chdir(originalWd)

		// Execute the command
		output := captureOutput(t, "", func() {
			installHookCmd.Run(installHookCmd, []string{})
		})

		// Assertions
		hookPath := filepath.Join(hooksDir, "pre-commit")
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			t.Fatal("Expected pre-commit hook file to be created, but it was not.")
		}

		content, _ := os.ReadFile(hookPath)
		if !strings.Contains(string(content), "gitego internal check-commit") {
			t.Error("Hook file was created, but does not contain the correct gitego command.")
		}
		if !strings.Contains(output, "hook installed successfully") {
			t.Errorf("Expected success message, but got: %s", output)
		}
	})

	t.Run("when hook exists and user confirms append", func(t *testing.T) {
		repoRoot, hooksDir := setupTestGitRepo(t)
		defer os.RemoveAll(repoRoot)
		os.Chdir(repoRoot)
		defer os.Chdir(originalWd)

		// Create a pre-existing hook file
		hookPath := filepath.Join(hooksDir, "pre-commit")
		initialContent := "#!/bin/sh\necho 'running other checks...'\n"
		os.WriteFile(hookPath, []byte(initialContent), 0755)

		// Execute the command, simulating 'y' for the prompt
		output := captureOutput(t, "y\n", func() {
			installHookCmd.Run(installHookCmd, []string{})
		})

		// Assertions
		finalContent, _ := os.ReadFile(hookPath)
		if !strings.HasPrefix(string(finalContent), initialContent) {
			t.Error("Expected hook to append, but it overwrote the original content.")
		}
		if !strings.Contains(string(finalContent), "gitego internal check-commit") {
			t.Error("Hook file was not appended with the correct gitego command.")
		}
		if !strings.Contains(output, "appended successfully") {
			t.Errorf("Expected append success message, but got: %s", output)
		}
	})

	t.Run("when hook is already installed", func(t *testing.T) {
		repoRoot, hooksDir := setupTestGitRepo(t)
		defer os.RemoveAll(repoRoot)
		os.Chdir(repoRoot)
		defer os.Chdir(originalWd)

		// Create a hook that already contains our command
		hookPath := filepath.Join(hooksDir, "pre-commit")
		os.WriteFile(hookPath, []byte("#!/bin/sh\ngitego internal check-commit\n"), 0755)

		output := captureOutput(t, "", func() {
			installHookCmd.Run(installHookCmd, []string{})
		})

		if !strings.Contains(output, "already installed") {
			t.Errorf("Expected 'already installed' message, but got: %s", output)
		}
	})
}
