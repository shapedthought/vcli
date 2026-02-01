package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// severityConfigFile represents the JSON structure for customizing severity classifications.
// Place as severity-config.json in ~/.vcli/ or VCLI_SETTINGS_PATH.
//
// Example:
//
//	{
//	  "job": { "isDisabled": "CRITICAL", "schedule": "WARNING" },
//	  "repository": { "type": "CRITICAL" },
//	  "sobr": { "isEnabled": "CRITICAL" },
//	  "encryption": { "hint": "WARNING" },
//	  "kms": { "type": "CRITICAL" }
//	}
type severityConfigFile struct {
	Job        map[string]string `json:"job,omitempty"`
	Repository map[string]string `json:"repository,omitempty"`
	Sobr       map[string]string `json:"sobr,omitempty"`
	Encryption map[string]string `json:"encryption,omitempty"`
	Kms        map[string]string `json:"kms,omitempty"`
}

var severityOverridesLoaded bool

// loadSeverityOverrides loads optional severity classification overrides
// from severity-config.json. Safe to call multiple times; only loads once.
func loadSeverityOverrides() {
	if severityOverridesLoaded {
		return
	}

	configPath := findSeverityConfig()
	if configPath == "" {
		severityOverridesLoaded = true
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	var config severityConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid severity config at %s: %v\n", configPath, err)
		return
	}

	applySeverityOverrides(config.Job, jobSeverityMap)
	applySeverityOverrides(config.Repository, repoSeverityMap)
	applySeverityOverrides(config.Sobr, sobrSeverityMap)
	applySeverityOverrides(config.Encryption, encryptionSeverityMap)
	applySeverityOverrides(config.Kms, kmsSeverityMap)

	severityOverridesLoaded = true
}

func findSeverityConfig() string {
	// Check VCLI_SETTINGS_PATH first
	settingsPath := os.Getenv("VCLI_SETTINGS_PATH")
	if settingsPath != "" {
		p := filepath.Join(settingsPath, "severity-config.json")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Fall back to ~/.vcli/
	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".vcli", "severity-config.json")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

func applySeverityOverrides(overrides map[string]string, sm SeverityMap) {
	for field, sev := range overrides {
		severity := Severity(strings.ToUpper(sev))
		switch severity {
		case SeverityCritical, SeverityWarning, SeverityInfo:
			sm[field] = severity
		default:
			fmt.Fprintf(os.Stderr, "Warning: Unknown severity %q for field %q (use CRITICAL, WARNING, or INFO)\n", sev, field)
		}
	}
}
