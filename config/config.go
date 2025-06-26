// config/config.go

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
			return cfg, nil
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	validateConfig(cfg)

	return cfg, nil
}

func validateConfig(cfg *Config) {
	if cfg.ActiveProfile != "" {
		if _, exists := cfg.Profiles[cfg.ActiveProfile]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: Active profile '%s' not found. It may have been deleted.\n", cfg.ActiveProfile)
		}
	}

	for _, rule := range cfg.AutoRules {
		if _, exists := cfg.Profiles[rule.Profile]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: Auto-switch rule for path '%s' points to a non-existent profile '%s'.\n", rule.Path, rule.Profile)
		}
	}
}

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
	profileName = c.ActiveProfile
	source = "Global gitego default"
	if c.ActiveProfile == "" {
		source = "No active gitego profile"
	}

	if len(c.AutoRules) == 0 {
		return profileName, source
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return profileName, source
	}

	evalDir, err := filepath.EvalSymlinks(currentDir)
	if err != nil {
		evalDir = currentDir
	}

	currentAbsDir, err := filepath.Abs(evalDir)
	if err != nil {
		return profileName, source
	}
	currentAbsDir = filepath.ToSlash(currentAbsDir)

	// Ensure both paths have a trailing slash for comparison
	if !strings.HasSuffix(currentAbsDir, "/") {
		currentAbsDir += "/"
	}

	bestMatchPath := ""
	for _, rule := range c.AutoRules {
		ruleEvalPath, err := filepath.EvalSymlinks(rule.Path)
		if err != nil {
			ruleEvalPath = rule.Path
		}
		ruleAbsPath, err := filepath.Abs(ruleEvalPath)
		if err != nil {
			continue
		}
		ruleAbsPath = filepath.ToSlash(ruleAbsPath)

		if !strings.HasSuffix(ruleAbsPath, "/") {
			ruleAbsPath += "/"
		}

		compareDir := currentAbsDir
		compareRulePath := ruleAbsPath
		if runtime.GOOS == "windows" {
			compareDir = strings.ToLower(compareDir)
			compareRulePath = strings.ToLower(compareRulePath)
		}

		if strings.HasPrefix(compareDir, compareRulePath) {
			if len(ruleAbsPath) > len(bestMatchPath) {
				bestMatchPath = ruleAbsPath
				profileName = rule.Profile
				source = fmt.Sprintf("gitego auto-rule for profile '%s'", rule.Profile)
			}
		}
	}

	return profileName, source
}

func EnsureProfileGitconfig(profileName string, profile *Profile) error {
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("could not create profiles directory: %w", err)
	}

	content := fmt.Sprintf("[user]\n    name = %s\n    email = %s\n", profile.Name, profile.Email)

	if profile.SSHKey != "" {
		sshCommand := fmt.Sprintf("ssh -i %s", profile.SSHKey)
		coreBlock := fmt.Sprintf("\n[core]\n    sshCommand = %s\n", sshCommand)
		content += coreBlock
	}

	filePath := filepath.Join(profilesDir, fmt.Sprintf("%s.gitconfig", profileName))
	return os.WriteFile(filePath, []byte(content), 0644)
}

func AddIncludeIf(profileName string, dirPath string) error {
	profileConfigPath := filepath.ToSlash(filepath.Join(profilesDir, fmt.Sprintf("%s.gitconfig", profileName)))
	includeLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, profileConfigPath)

	displayConfigPath := fmt.Sprintf("~/.gitego/profiles/%s.gitconfig", profileName)
	displayLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, displayConfigPath)

	file, err := os.Open(gitConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not open global .gitconfig: %w", err)
	}
	if file != nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lineFromConfig := filepath.ToSlash(scanner.Text())
			if strings.Contains(lineFromConfig, profileConfigPath) {
				fmt.Printf("✓ Auto-switch rule for profile '%s' on path '%s' already exists.\n", profileName, dirPath)
				return nil
			}
		}
	}

	f, err := os.OpenFile(gitConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open .gitconfig for writing: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n# gitego auto-switch rule\n" + includeLine + "\n"); err != nil {
		return fmt.Errorf("could not write to .gitconfig: %w", err)
	}

	fmt.Printf("✓ Added auto-switch rule to ~/.gitconfig:\n%s\n", displayLine)
	return nil
}

// RemoveIncludeIf finds and removes the includeIf directive associated with a profile.
func RemoveIncludeIf(profileName string) error {
	profileConfigFilename := fmt.Sprintf("%s.gitconfig", profileName)
	profileConfigPath := filepath.ToSlash(filepath.Join(profilesDir, profileConfigFilename))

	input, err := os.ReadFile(gitConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(input), "\n")
	var newLines []string
	var removing bool = false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// This is the start of a block to potentially remove
		if strings.HasPrefix(trimmedLine, "[includeIf") {
			// Look ahead to find the associated path line
			for j := i + 1; j < len(lines); j++ {
				nextLineTrimmed := strings.TrimSpace(lines[j])
				if strings.HasPrefix(nextLineTrimmed, "[") {
					// We've hit a new section before finding a path, so this isn't our block
					break
				}
				if strings.HasPrefix(nextLineTrimmed, "path") {
					// We found the path line. Check if it matches our target.
					if strings.Contains(filepath.ToSlash(nextLineTrimmed), profileConfigPath) {
						removing = true
						// Skip the comment line if it's there
						if i > 0 && strings.TrimSpace(lines[i-1]) == "# gitego auto-switch rule" {
							if len(newLines) > 0 {
								newLines = newLines[:len(newLines)-1]
							}
						}
					}
					break // Stop looking ahead
				}
			}
		}

		if !removing {
			newLines = append(newLines, line)
		}

		// A new section header means we are definitely done with the previous block
		if removing && strings.HasPrefix(trimmedLine, "[") && !strings.HasPrefix(trimmedLine, "[includeIf") {
			removing = false
		}
	}

	output := strings.Join(newLines, "\n")
	// Clean up any excess blank lines that might result from the deletion
	for strings.Contains(output, "\n\n\n") {
		output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	}

	output = strings.TrimSpace(output)
	if output != "" {
		output += "\n"
	}

	return os.WriteFile(gitConfigPath, []byte(output), 0644)
}
