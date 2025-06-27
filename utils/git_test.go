// utils/git_test.go

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// mockExecCommand remains the same.
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestGetEffectiveGitConfig(t *testing.T) {
	// Store the original function and defer its restoration.
	originalExecCommand := execCommand
	// Patch the package-level variable.
	execCommand = mockExecCommand
	defer func() { execCommand = originalExecCommand }()

	val, err := GetEffectiveGitConfig("user.email")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if val != "test@example.com" {
		t.Errorf("expected 'test@example.com', but got '%s'", val)
	}
}

func TestSetGlobalGitConfig(t *testing.T) {
	// Store the original function and defer its restoration.
	originalExecCommand := execCommand
	// Patch the package-level variable.
	execCommand = mockExecCommand
	defer func() { execCommand = originalExecCommand }()

	err := SetGlobalGitConfig("user.name", "Test User")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

// TestHelperProcess remains the same.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}

		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command to mock\n")
		os.Exit(1)
	}

	if args[0] == "git" && args[1] == "config" {
		if len(args) == 3 && args[2] == "user.email" {
			fmt.Fprint(os.Stdout, "test@example.com")
			return
		}

		if len(args) == 5 && args[2] == "--global" && args[3] == "user.name" {
			return
		}
	}

	fmt.Fprintf(os.Stderr, "unhandled mock command: %s\n", strings.Join(args, " "))
	os.Exit(1)
}
