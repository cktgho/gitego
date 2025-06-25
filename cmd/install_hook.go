// cmd/install_hook.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// The content of the shell script we'll write to .git/hooks/pre-commit
const hookScriptContent = `#!/bin/sh
# gitego pre-commit hook

# This command will check your commit author against the expected profile
# for this directory. If there's a mismatch, it will prompt you.
gitego internal check-commit
`

var installHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Installs the pre-commit hook to safeguard against misattributed commits.",
	Long: `Installs a pre-commit hook in the current Git repository.

This hook will automatically run before every commit to verify that your
commit author details match the gitego profile expected for this directory.
This provides a powerful safety net against accidental misattributed commits.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find the .git directory by starting from the current directory and walking up.
		gitRoot, err := findGitRoot(".")
		if err != nil {
			fmt.Println("Error: Not a git repository (or any of the parent directories).")
			return
		}

		hooksDir := filepath.Join(gitRoot, ".git", "hooks")
		hookPath := filepath.Join(hooksDir, "pre-commit")

		// Check if a pre-commit hook already exists.
		if _, err := os.Stat(hookPath); err == nil {
			fmt.Println("Warning: A pre-commit hook already exists.")
			fmt.Println("Please manually add the following line to your existing hook script:")
			fmt.Println("\ngitego internal check-commit\n")
			return
		}

		// Write the script content to the pre-commit file.
		err = os.WriteFile(hookPath, []byte(hookScriptContent), 0755) // 0755 makes it executable
		if err != nil {
			fmt.Printf("Error installing hook: %v\n", err)
			return
		}

		fmt.Printf("✓ gitego pre-commit hook installed successfully in %s\n", hookPath)
	},
}

// findGitRoot searches for the root of the git repository.
func findGitRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("git root not found")
}

func init() {
	rootCmd.AddCommand(installHookCmd)
}
