package remediation

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// FieldPolicy determines how a field should be handled during apply
type FieldPolicy string

const (
	// PolicyRemediable means the field should be sent to VBR (default)
	PolicyRemediable FieldPolicy = "remediable"

	// PolicySkip means the field should not be sent to VBR
	PolicySkip FieldPolicy = "skip"
)

// Config represents the remediation-config.yaml file
type Config struct {
	Repository map[string]FieldPolicy `yaml:"repository"`
	Sobr       map[string]FieldPolicy `yaml:"sobr"`
	Job        map[string]FieldPolicy `yaml:"job"`
	Kms        map[string]FieldPolicy `yaml:"kms"`
}

// KnownImmutableField describes a field that is known to be rejected by VBR
type KnownImmutableField struct {
	Reason string // Human-readable explanation
}

// KnownImmutableFields maps resource types to their known immutable fields
var KnownImmutableFields = map[string]map[string]KnownImmutableField{
	"VBRRepository": {
		"type": {
			Reason: "Repository type cannot be changed after creation. Recreate the repository in VBR console.",
		},
		"path": {
			Reason: "Changing path requires storage migration. Perform in VBR console.",
		},
		"host.id": {
			Reason: "Host cannot be changed. Recreate the repository on the new host.",
		},
	},
	"VBRScaleOutRepository": {
		"performanceTier.extents": {
			Reason: "Extents cannot be modified via API. Add or remove extents in VBR console.",
		},
	},
	"VBRJob": {
		// Jobs are generally more flexible; no fields are currently treated as known-immutable here.
	},
	"VBRKmsServer": {
		"type": {
			Reason: "KMS server type cannot be changed after creation.",
		},
	},
}

// LoadConfig loads remediation config from the standard locations.
// Returns an empty config (all fields remediable) if no config file exists.
// Returns an error if a config file exists but fails to parse.
// Locations checked (in order):
//  1. $OWLCTL_SETTINGS_PATH/remediation-config.yaml
//  2. ~/.owlctl/remediation-config.yaml
func LoadConfig() (*Config, error) {
	// Check OWLCTL_SETTINGS_PATH first
	if settingsPath := os.Getenv("OWLCTL_SETTINGS_PATH"); settingsPath != "" {
		configPath := filepath.Join(settingsPath, "remediation-config.yaml")
		cfg, err := loadConfigFile(configPath)
		if err == nil {
			return cfg, nil
		}
		// If file exists but failed to parse, return error
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load %s: %w", configPath, err)
		}
	}

	// Check ~/.owlctl/
	if usr, err := user.Current(); err == nil {
		configPath := filepath.Join(usr.HomeDir, ".owlctl", "remediation-config.yaml")
		cfg, err := loadConfigFile(configPath)
		if err == nil {
			return cfg, nil
		}
		// If file exists but failed to parse, return error
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load %s: %w", configPath, err)
		}
	}

	// No config file found - return empty config (all fields remediable)
	return &Config{
		Repository: make(map[string]FieldPolicy),
		Sobr:       make(map[string]FieldPolicy),
		Job:        make(map[string]FieldPolicy),
		Kms:        make(map[string]FieldPolicy),
	}, nil
}

// loadConfigFile loads a remediation config from a specific file path
func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize nil maps
	if cfg.Repository == nil {
		cfg.Repository = make(map[string]FieldPolicy)
	}
	if cfg.Sobr == nil {
		cfg.Sobr = make(map[string]FieldPolicy)
	}
	if cfg.Job == nil {
		cfg.Job = make(map[string]FieldPolicy)
	}
	if cfg.Kms == nil {
		cfg.Kms = make(map[string]FieldPolicy)
	}

	return &cfg, nil
}

// GetFieldPolicy returns the policy for a field.
// Priority order:
//  1. User config (remediation-config.yaml)
//  2. Known immutable hints
//  3. Default (remediable)
func (c *Config) GetFieldPolicy(resourceKind, fieldPath string) (FieldPolicy, string) {
	// Get the appropriate resource config
	var resourceConfig map[string]FieldPolicy
	switch resourceKind {
	case "VBRRepository":
		resourceConfig = c.Repository
	case "VBRScaleOutRepository":
		resourceConfig = c.Sobr
	case "VBRJob":
		resourceConfig = c.Job
	case "VBRKmsServer":
		resourceConfig = c.Kms
	}

	// Check user config first (overrides everything)
	if resourceConfig != nil {
		if policy, ok := resourceConfig[fieldPath]; ok {
			if policy == PolicySkip {
				return PolicySkip, "Skipped by remediation-config.yaml"
			}
			return PolicyRemediable, ""
		}
	}

	// Check known immutable hints
	if knownFields, ok := KnownImmutableFields[resourceKind]; ok {
		if field, ok := knownFields[fieldPath]; ok {
			return PolicySkip, field.Reason
		}
	}

	// Default: remediable
	return PolicyRemediable, ""
}

// FilterChanges filters field changes based on policy.
// Returns two slices: changes to apply and changes to skip (with reasons).
type SkippedChange struct {
	Path   string
	Reason string
}

func (c *Config) FilterChanges(resourceKind string, changes []FieldChange) ([]FieldChange, []SkippedChange) {
	var toApply []FieldChange
	var toSkip []SkippedChange

	for _, change := range changes {
		policy, reason := c.GetFieldPolicy(resourceKind, change.Path)
		if policy == PolicySkip {
			toSkip = append(toSkip, SkippedChange{
				Path:   change.Path,
				Reason: reason,
			})
		} else {
			toApply = append(toApply, change)
		}
	}

	return toApply, toSkip
}

// FieldChange is duplicated here to avoid import cycle.
// The cmd package defines the canonical version.
type FieldChange struct {
	Path     string
	OldValue interface{}
	NewValue interface{}
}
