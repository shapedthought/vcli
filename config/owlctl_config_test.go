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
