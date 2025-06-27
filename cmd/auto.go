// cmd/auto.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgreenwell/gitego/config"
	"github.com/spf13/cobra"
)

const (
	// exactArgs is the number of arguments for the auto command.
	exactArgs = 2
)

// autoRunner holds the dependencies for the auto command for mocking.
type autoRunner struct {
	load                   func() (*config.Config, error)
	save                   func(*config.Config) error
	ensureProfileGitconfig func(string, *config.Profile) error
	addIncludeIf           func(string, string) error
}

// run is the core logic for the auto command.
func (ar *autoRunner) run(cmd *cobra.Command, args []string) {
	path := args[0]
	profileName := args[1]

	cfg, err := ar.load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)

		return
	}
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		fmt.Printf("Error: Profile '%s' not found in gitego.\n", profileName)

		return
	}

	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Error resolving path '%s': %v\n", path, err)

		return
	}
	cleanPath := filepath.ToSlash(absPath)
	if !strings.HasSuffix(cleanPath, "/") {
		cleanPath += "/"
	}

	for _, rule := range cfg.AutoRules {
		if rule.Path == cleanPath && rule.Profile == profileName {
			fmt.Printf("✓ Auto-switch rule for profile '%s' on path '%s' already exists.\n", profileName, path)

			return
		}
	}

	fmt.Printf("Setting up new auto-switch rule for profile '%s'...\n", profileName)

	if err := ar.ensureProfileGitconfig(profileName, profile); err != nil {
		fmt.Printf("Error creating profile gitconfig: %v\n", err)

		return
	}

	if err := ar.addIncludeIf(profileName, cleanPath); err != nil {
		fmt.Printf("Error updating global .gitconfig: %v\n", err)

		return
	}

	newRule := &config.AutoRule{
		Path:    cleanPath,
		Profile: profileName,
	}
	cfg.AutoRules = append(cfg.AutoRules, newRule)
	if err := ar.save(cfg); err != nil {
		fmt.Printf("Warning: Git config updated, but failed to save rule to gitego config: %v\n", err)

		return
	}

	fmt.Println("✓ Rule setup complete.")
}

var autoCmd = &cobra.Command{
	Use:   "auto <path> <profile_name>",
	Short: "Automatically switch profiles based on directory.",
	Long: `Configures your global .gitconfig to automatically use a specific
profile whenever you are working inside the given directory path.`,
	Args: cobra.ExactArgs(exactArgs),
	Run: func(cmd *cobra.Command, args []string) {
		runner := &autoRunner{
			load:                   config.Load,
			save:                   func(c *config.Config) error { return c.Save() },
			ensureProfileGitconfig: config.EnsureProfileGitconfig,
			addIncludeIf:           config.AddIncludeIf,
		}
		runner.run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(autoCmd)
}
