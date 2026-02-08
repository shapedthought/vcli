package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// VCLIConfig represents the owlctl.yaml configuration file
type VCLIConfig struct {
	// APIVersion is the config schema version
	APIVersion string `yaml:"apiVersion,omitempty"`

	// Kind is the config document kind
	Kind string `yaml:"kind,omitempty"`

	// Groups maps group names to their configuration
	Groups map[string]GroupConfig `yaml:"groups,omitempty"`

	// Targets maps target names to their VBR server connection configuration
	Targets map[string]TargetConfig `yaml:"targets,omitempty"`

	// ConfigDir is the directory containing the owlctl.yaml file.
	// Populated during load, not serialized.
	ConfigDir string `yaml:"-"`

	// Legacy fields (still parsed, deprecated)

	// CurrentEnvironment is the active environment (e.g., "production", "dev")
	CurrentEnvironment string `yaml:"currentEnvironment,omitempty"`

	// Environments maps environment names to their overlay configurations
	Environments map[string]EnvironmentConfig `yaml:"environments,omitempty"`

	// DefaultOverlayDir is the directory to search for overlay files
	DefaultOverlayDir string `yaml:"defaultOverlayDir,omitempty"`
}

// TargetConfig defines a named VBR server connection
type TargetConfig struct {
	// URL is the VBR server address (e.g., "https://vbr-prod.example.com")
	URL string `yaml:"url" json:"url"`

	// Description is a human-readable description of the target
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GroupConfig defines a named group of spec files with optional profile and overlay
type GroupConfig struct {
	// Description is a human-readable description of the group
	Description string `yaml:"description,omitempty"`

	// Profile is the path to a Profile YAML file (base defaults)
	Profile string `yaml:"profile,omitempty"`

	// Overlay is the path to an Overlay YAML file (policy patch)
	Overlay string `yaml:"overlay,omitempty"`

	// Specs is the list of spec file paths in this group
	Specs []string `yaml:"specs"`
}

// EnvironmentConfig defines settings for a specific environment
type EnvironmentConfig struct {
	// Overlay is the path to the overlay file for this environment
	Overlay string `yaml:"overlay,omitempty"`

	// Profile is the VBR profile to use for this environment
	Profile string `yaml:"profile,omitempty"`

	// Labels are applied to all resources in this environment
	Labels map[string]string `yaml:"labels,omitempty"`
}

const (
	// DefaultConfigName is the default owlctl configuration file name
	DefaultConfigName = "owlctl.yaml"
	// EnvVarConfigPath allows overriding the config file location
	EnvVarConfigPath = "OWLCTL_CONFIG"
)

// LoadConfig loads the owlctl.yaml configuration file
// Searches in order: OWLCTL_CONFIG env var, current directory, home directory
func LoadConfig() (*VCLIConfig, error) {
	var configPath string

	// Check environment variable first
	if envPath := os.Getenv(EnvVarConfigPath); envPath != "" {
		configPath = envPath
	} else {
		// Check current directory
		if _, err := os.Stat(DefaultConfigName); err == nil {
			configPath = DefaultConfigName
		} else {
			// Check home directory
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			homeConfigPath := filepath.Join(home, ".owlctl", DefaultConfigName)
			if _, err := os.Stat(homeConfigPath); err == nil {
				configPath = homeConfigPath
			}
		}
	}

	// If no config file found, return default config
	if configPath == "" {
		return &VCLIConfig{
			Environments: make(map[string]EnvironmentConfig),
			Groups:       make(map[string]GroupConfig),
			Targets:      make(map[string]TargetConfig),
		}, nil
	}

	return LoadConfigFrom(configPath)
}

// LoadConfigFrom loads owlctl configuration from a specific file
func LoadConfigFrom(path string) (*VCLIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config VCLIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Initialize maps if nil
	if config.Environments == nil {
		config.Environments = make(map[string]EnvironmentConfig)
	}
	if config.Groups == nil {
		config.Groups = make(map[string]GroupConfig)
	}
	if config.Targets == nil {
		config.Targets = make(map[string]TargetConfig)
	}

	// Resolve ConfigDir from the file path
	absPath, err := filepath.Abs(path)
	if err != nil {
		config.ConfigDir = filepath.Dir(path)
	} else {
		config.ConfigDir = filepath.Dir(absPath)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the default location
func SaveConfig(config *VCLIConfig) error {
	configPath := DefaultConfigName
	if envPath := os.Getenv(EnvVarConfigPath); envPath != "" {
		configPath = envPath
	}

	return SaveConfigTo(config, configPath)
}

// SaveConfigTo saves the configuration to a specific file
func SaveConfigTo(config *VCLIConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetGroup returns the group configuration for the given name
func (c *VCLIConfig) GetGroup(name string) (GroupConfig, error) {
	group, exists := c.Groups[name]
	if !exists {
		return GroupConfig{}, fmt.Errorf("group %q not found in configuration", name)
	}
	return group, nil
}

// ListGroups returns sorted group names
func (c *VCLIConfig) ListGroups() []string {
	names := make([]string, 0, len(c.Groups))
	for name := range c.Groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetTarget returns the target configuration for the given name
func (c *VCLIConfig) GetTarget(name string) (TargetConfig, error) {
	target, exists := c.Targets[name]
	if !exists {
		return TargetConfig{}, fmt.Errorf("target %q not found in configuration", name)
	}
	if target.URL == "" {
		return TargetConfig{}, fmt.Errorf("target %q has no URL configured", name)
	}
	return target, nil
}

// ListTargets returns sorted target names
func (c *VCLIConfig) ListTargets() []string {
	names := make([]string, 0, len(c.Targets))
	for name := range c.Targets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ResolvePath resolves a path relative to the owlctl.yaml file's directory.
// If the path is already absolute, it is returned as-is.
func (c *VCLIConfig) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.ConfigDir == "" {
		return path
	}
	return filepath.Join(c.ConfigDir, path)
}

// HasDeprecatedFields returns true if the config uses deprecated fields
func (c *VCLIConfig) HasDeprecatedFields() bool {
	return c.CurrentEnvironment != "" || len(c.Environments) > 0
}

// WarnDeprecatedFields prints a deprecation warning to stderr if deprecated fields are in use
func (c *VCLIConfig) WarnDeprecatedFields() {
	if !c.HasDeprecatedFields() {
		return
	}
	fmt.Fprintln(os.Stderr, "Warning: 'currentEnvironment' and 'environments' in owlctl.yaml are deprecated.")
	fmt.Fprintln(os.Stderr, "         Migrate to 'groups' with profiles and overlays.")
	fmt.Fprintln(os.Stderr, "         See docs/migration-v0.10-to-v0.11.md for migration steps.")
}

// GetEnvironmentOverlay returns the overlay file path for the given environment
// If environment is empty, uses CurrentEnvironment from config
func (c *VCLIConfig) GetEnvironmentOverlay(environment string) (string, error) {
	env := environment
	if env == "" {
		env = c.CurrentEnvironment
	}

	if env == "" {
		return "", fmt.Errorf("no environment specified and no current environment set")
	}

	envConfig, exists := c.Environments[env]
	if !exists {
		return "", fmt.Errorf("environment %s not found in configuration", env)
	}

	if envConfig.Overlay == "" {
		return "", fmt.Errorf("no overlay defined for environment %s", env)
	}

	// If overlay path is relative, make it relative to DefaultOverlayDir
	overlayPath := envConfig.Overlay
	if !filepath.IsAbs(overlayPath) && c.DefaultOverlayDir != "" {
		overlayPath = filepath.Join(c.DefaultOverlayDir, overlayPath)
	}

	return overlayPath, nil
}

// GetCurrentEnvironment returns the current environment configuration
func (c *VCLIConfig) GetCurrentEnvironment() (string, EnvironmentConfig, error) {
	if c.CurrentEnvironment == "" {
		return "", EnvironmentConfig{}, fmt.Errorf("no current environment set")
	}

	envConfig, exists := c.Environments[c.CurrentEnvironment]
	if !exists {
		return "", EnvironmentConfig{}, fmt.Errorf("current environment %s not found in configuration", c.CurrentEnvironment)
	}

	return c.CurrentEnvironment, envConfig, nil
}

// SetEnvironment sets the current environment
func (c *VCLIConfig) SetEnvironment(environment string) error {
	if _, exists := c.Environments[environment]; !exists {
		return fmt.Errorf("environment %s not found in configuration", environment)
	}

	c.CurrentEnvironment = environment
	return nil
}

// AddEnvironment adds or updates an environment configuration
func (c *VCLIConfig) AddEnvironment(name string, config EnvironmentConfig) {
	if c.Environments == nil {
		c.Environments = make(map[string]EnvironmentConfig)
	}
	c.Environments[name] = config
}
