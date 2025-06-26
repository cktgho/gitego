// config/config.go

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile represents a single user profile with a name and email.
type Profile struct {
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	Username string `yaml:"username,omitempty"`
	SSHKey   string `yaml:"ssh_key,omitempty"`
	PAT      string `yaml:"-"`
}

// AutoRule represents a single directory-to-profile mapping.
type AutoRule struct {
	Path    string `yaml:"path"`
	Profile string `yaml:"profile"`
}

// Config represents the entire structure of our config file.
type Config struct {
	Profiles      map[string]*Profile `yaml:"profiles"`
	AutoRules     []*AutoRule         `yaml:"auto_rules,omitempty"`
	ActiveProfile string              `yaml:"active_profile,omitempty"`
}

var (
	gitegoConfigPath string
	gitConfigPath    string
	profilesDir      string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("could not get user home directory: %v", err))
	}
	gitegoConfigPath = filepath.Join(home, ".gitego", "config.yaml")
	profilesDir = filepath.Join(home, ".gitego", "profiles")
	gitConfigPath = filepath.Join(home, ".gitconfig")
}

// Load reads and decodes the gitego config.yaml file and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]*Profile),
	}
	data, err := os.ReadFile(gitegoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// It's not an error if the config file doesn't exist yet.
			return cfg, nil
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	// --- New Validation Logic ---
	validateConfig(cfg)

	return cfg, nil
}

// validateConfig checks for inconsistencies, like rules pointing to non-existent profiles.
func validateConfig(cfg *Config) {
	// Check if the global active profile is valid
	if cfg.ActiveProfile != "" {
		if _, exists := cfg.Profiles[cfg.ActiveProfile]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: Active profile '%s' not found. It may have been deleted.\n", cfg.ActiveProfile)
		}
	}

	// Check all auto-switch rules
	for _, rule := range cfg.AutoRules {
		if _, exists := cfg.Profiles[rule.Profile]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: Auto-switch rule for path '%s' points to a non-existent profile '%s'.\n", rule.Path, rule.Profile)
		}
	}
}

// Save writes the gitego Config struct to the config.yaml file.
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not serialize config to yaml: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(gitegoConfigPath), 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	if err := os.WriteFile(gitegoConfigPath, data, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}
	return nil
}

func (c *Config) GetActiveProfileForCurrentDir() (profileName, source string) {
	// 1. Default to the globally active profile, if any.
	profileName = c.ActiveProfile
	source = "Global gitego default"
	if c.ActiveProfile == "" {
		source = "No active gitego profile"
	}

	// 2. Check for a more specific auto-rule that matches the current directory.
	if len(c.AutoRules) > 0 {
		currentDir, err := os.Getwd()
		if err != nil {
			// If we can't get the CWD, we can't check rules, so we just return the default.
			return
		}
		currentAbsDir, err := filepath.Abs(currentDir)
		if err != nil {
			return
		}
		// Find the most specific (longest) matching path.
		bestMatchPath := ""
		for _, rule := range c.AutoRules {
			rulePath, err := filepath.Abs(strings.TrimSuffix(rule.Path, "/"))
			if err != nil {
				continue // Skip invalid paths in config.
			}
			if strings.HasPrefix(currentAbsDir, rulePath) {
				// If this path is more specific than the last one we found, use it.
				if len(rulePath) > len(bestMatchPath) {
					bestMatchPath = rulePath
					profileName = rule.Profile
					source = fmt.Sprintf("gitego auto-rule for profile '%s'", rule.Profile)
				}
			}
		}
	}
	return
}

func EnsureProfileGitconfig(profileName string, profile *Profile) error {
	// Ensure the directory for these profile-specific configs exists.
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("could not create profiles directory: %w", err)
	}

	// Start with the [user] block.
	content := fmt.Sprintf("[user]\n    name = %s\n    email = %s\n", profile.Name, profile.Email)

	// If an SSH key is specified for the profile, add the [core] block.
	if profile.SSHKey != "" {
		// This tells Git to use a specific SSH command for this profile.
		sshCommand := fmt.Sprintf("ssh -i %s", profile.SSHKey)
		coreBlock := fmt.Sprintf("\n[core]\n    sshCommand = %s\n", sshCommand)
		content += coreBlock
	}

	filePath := filepath.Join(profilesDir, fmt.Sprintf("%s.gitconfig", profileName))

	// Write the content to the file.
	return os.WriteFile(filePath, []byte(content), 0644)
}

/* func AddIncludeIf(profileName string, dirPath string) error {
	profileConfigPath := filepath.Join(profilesDir, fmt.Sprintf("%s.gitconfig", profileName))
	// We use tilde for user-friendly display, but filepath.Join for actual path logic.
	displayPath := filepath.Join("~/.gitego/profiles", fmt.Sprintf("%s.gitconfig", profileName))

	// This is the line we want to add to the global .gitconfig
	includeLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, profileConfigPath)
	displayLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, displayPath)

	// Read the global .gitconfig file to check if the line already exists.
	file, err := os.Open(gitConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not open global .gitconfig: %w", err)
	}
	if file != nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// If our include directive is already present, we don't need to do anything.
			if strings.Contains(scanner.Text(), profileConfigPath) {
				fmt.Printf("✓ Auto-switch rule for profile '%s' already exists.\n", profileName)
				return nil
			}
		}
	}

	// Append the new includeIf directive to the end of the global .gitconfig.
	f, err := os.OpenFile(gitConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open .gitconfig for writing: %w", err)
	}
	defer f.Close()

	// We add newlines before our entry to ensure it's separated from previous content.
	if _, err := f.WriteString("\n# gitego auto-switch rule\n" + includeLine + "\n"); err != nil {
		return fmt.Errorf("could not write to .gitconfig: %w", err)
	}

	fmt.Printf("✓ Added auto-switch rule to ~/.gitconfig:\n%s\n", displayLine)
	return nil
} */

// AddIncludeIf modifies the user's global ~/.gitconfig to include our new profile config.
// It adds an [includeIf "gitdir:path/to/work/"] directive.
func AddIncludeIf(profileName string, dirPath string) error {
	// Ensure both paths use forward slashes for cross-platform compatibility ---
	profileConfigPath := filepath.ToSlash(filepath.Join(profilesDir, fmt.Sprintf("%s.gitconfig", profileName)))

	// The display path for user feedback remains the same.
	displayPath := filepath.Join("~/.gitego/profiles", fmt.Sprintf("%s.gitconfig", profileName))

	// This is the line we want to add to the global .gitconfig
	includeLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, profileConfigPath)
	displayLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, displayPath)

	// Read the global .gitconfig file to check if the line already exists.
	file, err := os.Open(gitConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not open global .gitconfig: %w", err)
	}
	if file != nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// If our include directive is already present, we don't need to do anything.
			if strings.Contains(scanner.Text(), profileConfigPath) {
				fmt.Printf("✓ Auto-switch rule for profile '%s' already exists.\n", profileName)
				return nil
			}
		}
	}

	// Append the new includeIf directive to the end of the global .gitconfig.
	f, err := os.OpenFile(gitConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open .gitconfig for writing: %w", err)
	}
	defer f.Close()

	// We add newlines before our entry to ensure it's separated from previous content.
	if _, err := f.WriteString("\n# gitego auto-switch rule\n" + includeLine + "\n"); err != nil {
		return fmt.Errorf("could not write to .gitconfig: %w", err)
	}

	fmt.Printf("✓ Added auto-switch rule to ~/.gitconfig:\n%s\n", displayLine)
	return nil
}
