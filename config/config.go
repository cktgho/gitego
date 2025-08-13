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

const (
	// dirPermissions are the default permissions for directories created by gitego.
	dirPermissions = 0755
	// filePermissions are the default permissions for files created by gitego.
	filePermissions = 0644
)

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
			fmt.Fprintf(os.Stderr,
				"Warning: Auto-switch rule for path '%s' points to a non-existent profile '%s'.\n",
				rule.Path, rule.Profile)
		}
	}
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not serialize config to yaml: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(gitegoConfigPath), dirPermissions); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	if err := os.WriteFile(gitegoConfigPath, data, filePermissions); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

func (c *Config) GetActiveProfileForCurrentDir() (profileName, source string) {
	profileName = c.ActiveProfile
	source = getDefaultSource(c.ActiveProfile)

	if len(c.AutoRules) == 0 {
		return profileName, source
	}

	currentAbsDir, err := getCurrentAbsDir()
	if err != nil {
		return profileName, source
	}

	bestMatch := c.findBestMatchingRule(currentAbsDir)
	if bestMatch != nil {
		profileName = bestMatch.Profile
		source = fmt.Sprintf("gitego auto-rule for profile '%s'", bestMatch.Profile)
	}

	return profileName, source
}

func getDefaultSource(activeProfile string) string {
	if activeProfile == "" {
		return "No active gitego profile"
	}

	return "Global gitego default"
}

func getCurrentAbsDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	evalDir, err := filepath.EvalSymlinks(currentDir)
	if err != nil {
		evalDir = currentDir
	}

	currentAbsDir, err := filepath.Abs(evalDir)
	if err != nil {
		return "", err
	}

	currentAbsDir = filepath.ToSlash(currentAbsDir)
	if !strings.HasSuffix(currentAbsDir, "/") {
		currentAbsDir += "/"
	}

	return currentAbsDir, nil
}

func (c *Config) findBestMatchingRule(currentAbsDir string) *AutoRule {
	var bestMatch *AutoRule

	bestMatchPath := ""

	for _, rule := range c.AutoRules {
		ruleAbsPath, err := cleanPath(rule.Path)
		if err != nil {
			continue
		}

		if c.isPathMatch(currentAbsDir, ruleAbsPath) && len(ruleAbsPath) > len(bestMatchPath) {
			bestMatchPath = ruleAbsPath
			bestMatch = rule
		}
	}

	return bestMatch
}

func (c *Config) isPathMatch(currentAbsDir, ruleAbsPath string) bool {
	compareDir := currentAbsDir
	compareRulePath := ruleAbsPath

	if runtime.GOOS == "windows" {
		compareDir = strings.ToLower(compareDir)
		compareRulePath = strings.ToLower(compareRulePath)
	}

	return strings.HasPrefix(compareDir, compareRulePath)
}

func cleanPath(path string) (string, error) {
	ruleEvalPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		ruleEvalPath = path
	}

	ruleAbsPath, err := filepath.Abs(ruleEvalPath)
	if err != nil {
		return "", err
	}

	ruleAbsPath = filepath.ToSlash(ruleAbsPath)

	if !strings.HasSuffix(ruleAbsPath, "/") {
		ruleAbsPath += "/"
	}

	return ruleAbsPath, nil
}

func EnsureProfileGitconfig(profileName string, profile *Profile) error {
	if err := os.MkdirAll(profilesDir, dirPermissions); err != nil {
		return fmt.Errorf("could not create profiles directory: %w", err)
	}

	content := fmt.Sprintf("[user]\n    name = %s\n    email = %s\n", profile.Name, profile.Email)

	if profile.SSHKey != "" {
		sshCommand := fmt.Sprintf("ssh -i %s", profile.SSHKey)
		coreBlock := fmt.Sprintf("\n[core]\n    sshCommand = %s\n", sshCommand)
		content += coreBlock
	}

	filePath := filepath.Join(profilesDir, fmt.Sprintf("%s.gitconfig", profileName))

	return os.WriteFile(filePath, []byte(content), filePermissions)
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
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Printf("Warning: Failed to close gitconfig file: %v\n", err)
			}
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lineFromConfig := filepath.ToSlash(scanner.Text())
			if strings.Contains(lineFromConfig, profileConfigPath) {
				fmt.Printf("✓ Auto-switch rule for profile '%s' on path '%s' already exists.\n", profileName, dirPath)

				return nil
			}
		}
	}

	f, err := os.OpenFile(gitConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return fmt.Errorf("could not open .gitconfig for writing: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: Failed to close gitconfig file: %v\n", err)
		}
	}()

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
	newLines := removeGitegoRules(lines, profileConfigPath)
	output := formatOutput(newLines)

	return os.WriteFile(gitConfigPath, []byte(output), filePermissions)
}

func removeGitegoRules(lines []string, profileConfigPath string) []string {
	var newLines []string

	var removing bool

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "[includeIf") && isGitegoRule(lines, i, profileConfigPath) {
			removing = true
			newLines = removeCommentIfPresent(newLines, lines, i)

			continue
		}

		if !removing {
			newLines = append(newLines, line)
		}

		if removing && isNewSection(trimmedLine) {
			removing = false
		}
	}

	return newLines
}

func removeCommentIfPresent(newLines []string, lines []string, i int) []string {
	if i > 0 && strings.TrimSpace(lines[i-1]) == "# gitego auto-switch rule" && len(newLines) > 0 {
		return newLines[:len(newLines)-1]
	}

	return newLines
}

func isNewSection(trimmedLine string) bool {
	return strings.HasPrefix(trimmedLine, "[") && !strings.HasPrefix(trimmedLine, "[includeIf")
}

func formatOutput(lines []string) string {
	output := strings.Join(lines, "\n")
	output = strings.TrimSpace(output)

	if output != "" {
		output += "\n"
	}

	return output
}

func isGitegoRule(lines []string, index int, profileConfigPath string) bool {
	for j := index + 1; j < len(lines); j++ {
		nextLineTrimmed := strings.TrimSpace(lines[j])
		if strings.HasPrefix(nextLineTrimmed, "[") {
			return false
		}

		if strings.HasPrefix(nextLineTrimmed, "path") {
			return strings.Contains(filepath.ToSlash(nextLineTrimmed), profileConfigPath)
		}
	}

	return false
}
