// cmd/integration_test.go

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	// This will be the path to the compiled test binary.
	gitegoBinary string
)

// TestMain is a special function that runs once for the entire package.
// We use it to compile our gitego binary for the integration tests.
func TestMain(m *testing.M) {
	// Build the gitego binary into a temporary location.
	tempDir, err := os.MkdirTemp("", "gitego-integration-")
	if err != nil {
		panic("Failed to create temp dir for binary: " + err.Error())
	}
	defer os.RemoveAll(tempDir)

	gitegoBinary = filepath.Join(tempDir, "gitego.exe") // .exe is safe for non-Windows OSes
	buildCmd := exec.Command("go", "build", "-o", gitegoBinary, "..")
	if err := buildCmd.Run(); err != nil {
		panic("Failed to build gitego binary: " + err.Error())
	}

	// Run the actual tests.
	exitCode := m.Run()
	os.Exit(exitCode)
}

// setupTestEnvironment creates a clean temporary home directory for a test run.
func setupTestEnvironment(t *testing.T) (homeDir string) {
	t.Helper()
	homeDir, err := os.MkdirTemp("", "gitego-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir) // for Windows
	return homeDir
}

func TestIntegration_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	homeDir := setupTestEnvironment(t)
	defer os.RemoveAll(homeDir)

	gitConfigPath := filepath.Join(homeDir, ".gitconfig")
	gitegoConfigPath := filepath.Join(homeDir, ".gitego", "config.yaml")

	// --- Step 1: Add a profile ---
	addCmd := exec.Command(gitegoBinary, "add", "work", "--name", "Test User", "--email", "test@work.com")
	addCmd.Env = os.Environ() // Ensure it uses the temp HOME
	if err := addCmd.Run(); err != nil {
		t.Fatalf("gitego add command failed: %v", err)
	}

	// --- Step 2: Use the profile ---
	useCmd := exec.Command(gitegoBinary, "use", "work")
	useCmd.Env = os.Environ()
	if err := useCmd.Run(); err != nil {
		t.Fatalf("gitego use command failed: %v", err)
	}

	// Assert that .gitconfig was updated
	gitConfigBytes, _ := os.ReadFile(gitConfigPath)
	gitConfigContent := string(gitConfigBytes)
	if !strings.Contains(gitConfigContent, "name = Test User") || !strings.Contains(gitConfigContent,
		"email = test@work.com") {
		t.Fatal(".gitconfig was not updated correctly by 'use' command.")
	}

	// --- Step 3: Set an auto rule ---
	autoDir := filepath.Join(homeDir, "work-projects")
	os.Mkdir(autoDir, 0755)

	autoCmd := exec.Command(gitegoBinary, "auto", autoDir, "work")
	autoCmd.Env = os.Environ()
	if err := autoCmd.Run(); err != nil {
		t.Fatalf("gitego auto command failed: %v", err)
	}

	// Assert that .gitconfig has the includeIf rule
	gitConfigBytes, _ = os.ReadFile(gitConfigPath)
	gitConfigContent = string(gitConfigBytes)
	if !strings.Contains(gitConfigContent, "[includeIf") {
		t.Fatal(".gitconfig was not updated with an includeIf rule by 'auto' command.")
	}

	// --- Step 4: Remove the profile and assert cleanup ---
	rmCmd := exec.Command(gitegoBinary, "rm", "work", "--force")
	rmCmd.Env = os.Environ()
	if err := rmCmd.Run(); err != nil {
		t.Fatalf("gitego rm command failed: %v", err)
	}

	// Assert that gitego config is now empty
	gitegoConfigBytes, _ := os.ReadFile(gitegoConfigPath)
	gitegoConfigContent := string(gitegoConfigBytes)
	if strings.Contains(gitegoConfigContent, "work:") {
		t.Fatal("Profile 'work' was not removed from gitego's config.yaml.")
	}

	// Assert that the includeIf rule was removed from .gitconfig
	gitConfigBytes, _ = os.ReadFile(gitConfigPath)
	gitConfigContent = string(gitConfigBytes)
	if strings.Contains(gitConfigContent, "[includeIf") {
		t.Fatal("The includeIf rule was not cleaned up from .gitconfig by 'rm' command.")
	}
}
