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
	PAT      string `yaml:"-"` // The "-" tag is critical for security; it prevents this field from being serialized to YAML.
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
	// gitegoConfigPath is the path to gitego's own config file.
	gitegoConfigPath string
	// gitConfigPath is the path to the user's global .gitconfig file.
	gitConfigPath string
	// profilesDir is the path to the directory where gitego stores profile-specific gitconfigs.
	profilesDir string
)

// init runs automatically, setting up our required paths.
func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("could not get user home directory: %v", err))
	}
	gitegoConfigPath = filepath.Join(home, ".gitego", "config.yaml")
	profilesDir = filepath.Join(home, ".gitego", "profiles")
	gitConfigPath = filepath.Join(home, ".gitconfig")
}

// Load reads and decodes the gitego config.yaml file.
func Load() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]*Profile),
	}
	data, err := os.ReadFile(gitegoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	return cfg, nil
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

// EnsureProfileGitconfig creates a dedicated .gitconfig file for a specific profile.
// This file will contain the user.name and user.email for that profile.
// Example file path: ~/.gitego/profiles/work.gitconfig
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

// AddIncludeIf modifies the user's global ~/.gitconfig to include our new profile config.
// It adds an [includeIf "gitdir:path/to/work/"] directive.
func AddIncludeIf(profileName string, dirPath string) error {
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
}
