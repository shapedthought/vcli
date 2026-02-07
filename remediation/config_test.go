package remediation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFieldPolicy_KnownImmutable(t *testing.T) {
	cfg := &Config{
		Repository: make(map[string]FieldPolicy),
	}

	// Test known immutable field
	policy, reason := cfg.GetFieldPolicy("VBRRepository", "type")
	if policy != PolicySkip {
		t.Errorf("expected skip for known immutable field, got %s", policy)
	}
	if reason == "" {
		t.Error("expected reason for known immutable field")
	}
}

func TestGetFieldPolicy_Remediable(t *testing.T) {
	cfg := &Config{
		Repository: make(map[string]FieldPolicy),
	}

	// Test field that is not known immutable
	policy, reason := cfg.GetFieldPolicy("VBRRepository", "description")
	if policy != PolicyRemediable {
		t.Errorf("expected remediable for normal field, got %s", policy)
	}
	if reason != "" {
		t.Errorf("expected no reason for remediable field, got %s", reason)
	}
}

func TestGetFieldPolicy_UserConfigOverride(t *testing.T) {
	cfg := &Config{
		Repository: map[string]FieldPolicy{
			"type": PolicyRemediable, // Override the built-in hint
		},
	}

	// User config should override known immutable
	policy, _ := cfg.GetFieldPolicy("VBRRepository", "type")
	if policy != PolicyRemediable {
		t.Errorf("expected user config to override built-in hint, got %s", policy)
	}
}

func TestGetFieldPolicy_UserConfigSkip(t *testing.T) {
	cfg := &Config{
		Job: map[string]FieldPolicy{
			"storage.backupRepositoryId": PolicySkip,
		},
	}

	// User config should skip field
	policy, reason := cfg.GetFieldPolicy("VBRJob", "storage.backupRepositoryId")
	if policy != PolicySkip {
		t.Errorf("expected skip from user config, got %s", policy)
	}
	if reason != "Skipped by remediation-config.yaml" {
		t.Errorf("expected config reason, got %s", reason)
	}
}

func TestFilterChanges(t *testing.T) {
	cfg := &Config{
		Repository: map[string]FieldPolicy{
			"path": PolicySkip,
		},
	}

	changes := []FieldChange{
		{Path: "description", OldValue: "old", NewValue: "new"},
		{Path: "path", OldValue: "/old/path", NewValue: "/new/path"},
		{Path: "type", OldValue: "ReFS", NewValue: "NTFS"}, // Known immutable
	}

	toApply, toSkip := cfg.FilterChanges("VBRRepository", changes)

	if len(toApply) != 1 {
		t.Errorf("expected 1 field to apply, got %d", len(toApply))
	}
	if len(toSkip) != 2 {
		t.Errorf("expected 2 fields to skip, got %d", len(toSkip))
	}

	// Verify the right field is applied
	if toApply[0].Path != "description" {
		t.Errorf("expected description to be applied, got %s", toApply[0].Path)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Ensure no config file exists
	origPath := os.Getenv("OWLCTL_SETTINGS_PATH")
	os.Setenv("OWLCTL_SETTINGS_PATH", "/nonexistent/path/that/does/not/exist")
	defer os.Setenv("OWLCTL_SETTINGS_PATH", origPath)

	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("expected no error for missing config, got %v", err)
	}
	if cfg == nil {
		t.Error("expected empty config, got nil")
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temp directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "remediation-config.yaml")

	configContent := `
repository:
  path: skip
  description: remediable
job:
  storage.backupRepositoryId: skip
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set env to use temp dir
	origPath := os.Getenv("OWLCTL_SETTINGS_PATH")
	os.Setenv("OWLCTL_SETTINGS_PATH", tmpDir)
	defer os.Setenv("OWLCTL_SETTINGS_PATH", origPath)

	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify loaded policies
	if cfg.Repository["path"] != PolicySkip {
		t.Errorf("expected skip for path, got %s", cfg.Repository["path"])
	}
	if cfg.Repository["description"] != PolicyRemediable {
		t.Errorf("expected remediable for description, got %s", cfg.Repository["description"])
	}
	if cfg.Job["storage.backupRepositoryId"] != PolicySkip {
		t.Errorf("expected skip for backupRepositoryId, got %s", cfg.Job["storage.backupRepositoryId"])
	}
}
