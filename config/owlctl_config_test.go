package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "owlctl-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test config
	testConfig := &VCLIConfig{
		CurrentEnvironment: "production",
		DefaultOverlayDir:  "overlays",
		Environments: map[string]EnvironmentConfig{
			"production": {
				Overlay: "prod-overlay.yaml",
				Profile: "vbr-prod",
				Labels:  map[string]string{"env": "production"},
			},
			"development": {
				Overlay: "dev-overlay.yaml",
				Profile: "vbr-dev",
				Labels:  map[string]string{"env": "development"},
			},
		},
	}

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	if err := SaveConfigTo(testConfig, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config back
	loaded, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config
	if loaded.CurrentEnvironment != "production" {
		t.Errorf("Expected CurrentEnvironment=production, got %s", loaded.CurrentEnvironment)
	}

	if len(loaded.Environments) != 2 {
		t.Errorf("Expected 2 environments, got %d", len(loaded.Environments))
	}

	prodEnv, exists := loaded.Environments["production"]
	if !exists {
		t.Fatal("Production environment not found")
	}

	if prodEnv.Overlay != "prod-overlay.yaml" {
		t.Errorf("Expected overlay=prod-overlay.yaml, got %s", prodEnv.Overlay)
	}

	if prodEnv.Profile != "vbr-prod" {
		t.Errorf("Expected profile=vbr-prod, got %s", prodEnv.Profile)
	}
}

func TestGetEnvironmentOverlay(t *testing.T) {
	config := &VCLIConfig{
		CurrentEnvironment: "production",
		DefaultOverlayDir:  "/overlays",
		Environments: map[string]EnvironmentConfig{
			"production": {
				Overlay: "prod.yaml",
			},
			"development": {
				Overlay: "/absolute/dev.yaml",
			},
		},
	}

	// Test with current environment
	overlay, err := config.GetEnvironmentOverlay("")
	if err != nil {
		t.Fatalf("Failed to get overlay: %v", err)
	}
	expected := filepath.Join("/overlays", "prod.yaml")
	if overlay != expected {
		t.Errorf("Expected overlay=%s, got %s", expected, overlay)
	}

	// Test with explicit environment
	overlay, err = config.GetEnvironmentOverlay("development")
	if err != nil {
		t.Fatalf("Failed to get overlay: %v", err)
	}
	// Absolute path should not be joined with DefaultOverlayDir
	if overlay != "/absolute/dev.yaml" {
		t.Errorf("Expected overlay=/absolute/dev.yaml, got %s", overlay)
	}

	// Test with non-existent environment
	_, err = config.GetEnvironmentOverlay("staging")
	if err == nil {
		t.Error("Expected error for non-existent environment")
	}
}

func TestSetEnvironment(t *testing.T) {
	config := &VCLIConfig{
		CurrentEnvironment: "production",
		Environments: map[string]EnvironmentConfig{
			"production": {Overlay: "prod.yaml"},
			"development": {Overlay: "dev.yaml"},
		},
	}

	// Change to development
	if err := config.SetEnvironment("development"); err != nil {
		t.Fatalf("Failed to set environment: %v", err)
	}

	if config.CurrentEnvironment != "development" {
		t.Errorf("Expected CurrentEnvironment=development, got %s", config.CurrentEnvironment)
	}

	// Try to set non-existent environment
	if err := config.SetEnvironment("staging"); err == nil {
		t.Error("Expected error when setting non-existent environment")
	}
}

func TestAddEnvironment(t *testing.T) {
	config := &VCLIConfig{}

	config.AddEnvironment("staging", EnvironmentConfig{
		Overlay: "staging.yaml",
		Profile: "vbr-staging",
		Labels:  map[string]string{"env": "staging"},
	})

	if len(config.Environments) != 1 {
		t.Errorf("Expected 1 environment, got %d", len(config.Environments))
	}

	staging, exists := config.Environments["staging"]
	if !exists {
		t.Fatal("Staging environment not found")
	}

	if staging.Overlay != "staging.yaml" {
		t.Errorf("Expected overlay=staging.yaml, got %s", staging.Overlay)
	}
}

func TestGroupConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-group-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `apiVersion: owlctl.veeam.com/v1
kind: Config
groups:
  sql-tier:
    description: SQL Server backup group
    profile: profiles/gold.yaml
    overlay: overlays/compliance.yaml
    specs:
      - specs/sql-vm-01.yaml
      - specs/sql-vm-02.yaml
  file-tier:
    description: File server backups
    specs:
      - specs/file-server.yaml
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify groups parsed
	if len(cfg.Groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(cfg.Groups))
	}

	// GetGroup
	sqlGroup, err := cfg.GetGroup("sql-tier")
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}
	if sqlGroup.Description != "SQL Server backup group" {
		t.Errorf("Expected description, got %s", sqlGroup.Description)
	}
	if sqlGroup.Profile != "profiles/gold.yaml" {
		t.Errorf("Expected profile path, got %s", sqlGroup.Profile)
	}
	if sqlGroup.Overlay != "overlays/compliance.yaml" {
		t.Errorf("Expected overlay path, got %s", sqlGroup.Overlay)
	}
	if len(sqlGroup.Specs) != 2 {
		t.Errorf("Expected 2 specs, got %d", len(sqlGroup.Specs))
	}

	// GetGroup — not found
	_, err = cfg.GetGroup("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent group")
	}

	// ListGroups — sorted
	names := cfg.ListGroups()
	if len(names) != 2 || names[0] != "file-tier" || names[1] != "sql-tier" {
		t.Errorf("Expected sorted [file-tier, sql-tier], got %v", names)
	}

	// ConfigDir populated
	if cfg.ConfigDir != tmpDir {
		t.Errorf("Expected ConfigDir=%s, got %s", tmpDir, cfg.ConfigDir)
	}
}

func TestGroupConfigBackwardCompat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-compat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Config with only legacy environments (no groups)
	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `currentEnvironment: production
defaultOverlayDir: overlays
environments:
  production:
    overlay: prod.yaml
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Legacy fields still work
	if cfg.CurrentEnvironment != "production" {
		t.Errorf("Expected CurrentEnvironment=production, got %s", cfg.CurrentEnvironment)
	}
	if len(cfg.Environments) != 1 {
		t.Errorf("Expected 1 environment, got %d", len(cfg.Environments))
	}

	// Groups initialized but empty
	if cfg.Groups == nil {
		t.Error("Expected Groups to be initialized")
	}
	if len(cfg.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(cfg.Groups))
	}

	// HasDeprecatedFields
	if !cfg.HasDeprecatedFields() {
		t.Error("Expected HasDeprecatedFields=true for legacy config")
	}
}

func TestResolvePath(t *testing.T) {
	cfg := &VCLIConfig{
		ConfigDir: "/home/user/vbr-config",
	}

	// Relative path resolved against ConfigDir
	resolved := cfg.ResolvePath("specs/sql.yaml")
	expected := filepath.Join("/home/user/vbr-config", "specs/sql.yaml")
	if resolved != expected {
		t.Errorf("Expected %s, got %s", expected, resolved)
	}

	// Absolute path returned as-is
	resolved = cfg.ResolvePath("/absolute/path/spec.yaml")
	if resolved != "/absolute/path/spec.yaml" {
		t.Errorf("Expected absolute path unchanged, got %s", resolved)
	}

	// Empty ConfigDir — return relative path as-is
	cfgEmpty := &VCLIConfig{}
	resolved = cfgEmpty.ResolvePath("specs/sql.yaml")
	if resolved != "specs/sql.yaml" {
		t.Errorf("Expected unchanged relative path, got %s", resolved)
	}
}

func TestHasDeprecatedFields(t *testing.T) {
	// No deprecated fields
	cfg := &VCLIConfig{
		Groups: map[string]GroupConfig{
			"test": {Specs: []string{"spec.yaml"}},
		},
	}
	if cfg.HasDeprecatedFields() {
		t.Error("Expected HasDeprecatedFields=false for groups-only config")
	}

	// Has currentEnvironment
	cfg2 := &VCLIConfig{CurrentEnvironment: "prod"}
	if !cfg2.HasDeprecatedFields() {
		t.Error("Expected HasDeprecatedFields=true")
	}

	// Has environments
	cfg3 := &VCLIConfig{
		Environments: map[string]EnvironmentConfig{
			"prod": {Overlay: "prod.yaml"},
		},
	}
	if !cfg3.HasDeprecatedFields() {
		t.Error("Expected HasDeprecatedFields=true")
	}
}
