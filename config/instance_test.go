package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstanceConfigParsing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-instance-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `apiVersion: owlctl.veeam.com/v1
kind: Config
instances:
  vbr-prod:
    product: vbr
    url: https://vbr-prod.example.com
    port: 9419
    insecure: true
    credentialRef: PROD
    description: Production VBR server
  vbr-dr:
    product: vbr
    url: https://vbr-dr.example.com
    description: DR site VBR
  azure-prod:
    product: azure
    url: https://azure-prod.example.com
    insecure: false
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify instances parsed
	if len(cfg.Instances) != 3 {
		t.Fatalf("Expected 3 instances, got %d", len(cfg.Instances))
	}

	// GetInstance — full fields
	prod, err := cfg.GetInstance("vbr-prod")
	if err != nil {
		t.Fatalf("GetInstance(vbr-prod) failed: %v", err)
	}
	if prod.Product != "vbr" {
		t.Errorf("Expected product=vbr, got %s", prod.Product)
	}
	if prod.URL != "https://vbr-prod.example.com" {
		t.Errorf("Expected URL, got %s", prod.URL)
	}
	if prod.Port != 9419 {
		t.Errorf("Expected port=9419, got %d", prod.Port)
	}
	if prod.Insecure == nil || !*prod.Insecure {
		t.Error("Expected insecure=true")
	}
	if prod.CredentialRef != "PROD" {
		t.Errorf("Expected credentialRef=PROD, got %s", prod.CredentialRef)
	}
	if prod.Description != "Production VBR server" {
		t.Errorf("Expected description, got %s", prod.Description)
	}

	// GetInstance — no credentialRef or port (defaults)
	dr, err := cfg.GetInstance("vbr-dr")
	if err != nil {
		t.Fatalf("GetInstance(vbr-dr) failed: %v", err)
	}
	if dr.Port != 0 {
		t.Errorf("Expected port=0, got %d", dr.Port)
	}
	if dr.CredentialRef != "" {
		t.Errorf("Expected empty credentialRef, got %s", dr.CredentialRef)
	}
	if dr.Insecure != nil {
		t.Error("Expected insecure=nil (unset)")
	}

	// GetInstance — insecure explicitly false
	azure, err := cfg.GetInstance("azure-prod")
	if err != nil {
		t.Fatalf("GetInstance(azure-prod) failed: %v", err)
	}
	if azure.Insecure == nil || *azure.Insecure {
		t.Error("Expected insecure=false (explicitly set)")
	}

	// GetInstance — not found
	_, err = cfg.GetInstance("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent instance")
	}

	// ListInstances — sorted
	names := cfg.ListInstances()
	if len(names) != 3 || names[0] != "azure-prod" || names[1] != "vbr-dr" || names[2] != "vbr-prod" {
		t.Errorf("Expected sorted [azure-prod, vbr-dr, vbr-prod], got %v", names)
	}
}

func TestInstanceConfigValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-instance-validate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `instances:
  no-url:
    product: vbr
  no-product:
    url: https://example.com
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Missing URL
	_, err = cfg.GetInstance("no-url")
	if err == nil {
		t.Error("Expected error for instance with no URL")
	}

	// Missing product
	_, err = cfg.GetInstance("no-product")
	if err == nil {
		t.Error("Expected error for instance with no product")
	}
}

func TestInstanceConfigEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-instance-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `apiVersion: owlctl.veeam.com/v1
kind: Config
groups:
  test:
    specs:
      - spec.yaml
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Instances map should be initialized but empty
	if cfg.Instances == nil {
		t.Error("Expected Instances to be initialized")
	}
	if len(cfg.Instances) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(cfg.Instances))
	}

	names := cfg.ListInstances()
	if len(names) != 0 {
		t.Errorf("Expected 0 instance names, got %d", len(names))
	}
}

func TestResolveInstance(t *testing.T) {
	cfg := &VCLIConfig{
		Instances: map[string]InstanceConfig{
			"vbr-prod": {
				Product:       "vbr",
				URL:           "https://vbr-prod.example.com",
				Port:          9419,
				CredentialRef: "PROD",
			},
			"vbr-dr": {
				Product: "vbr",
				URL:     "https://vbr-dr.example.com",
			},
		},
	}

	// Set env vars for PROD credential ref
	os.Setenv("OWLCTL_PROD_USERNAME", "prod-admin")
	os.Setenv("OWLCTL_PROD_PASSWORD", "prod-secret")
	defer os.Unsetenv("OWLCTL_PROD_USERNAME")
	defer os.Unsetenv("OWLCTL_PROD_PASSWORD")

	// Set default env vars for DR (no credentialRef)
	os.Setenv("OWLCTL_USERNAME", "default-admin")
	os.Setenv("OWLCTL_PASSWORD", "default-secret")
	defer os.Unsetenv("OWLCTL_USERNAME")
	defer os.Unsetenv("OWLCTL_PASSWORD")

	// Resolve with credentialRef
	resolved, err := ResolveInstance(cfg, "vbr-prod")
	if err != nil {
		t.Fatalf("ResolveInstance(vbr-prod) failed: %v", err)
	}
	if resolved.Name != "vbr-prod" {
		t.Errorf("Expected name=vbr-prod, got %s", resolved.Name)
	}
	if resolved.Product != "vbr" {
		t.Errorf("Expected product=vbr, got %s", resolved.Product)
	}
	if resolved.URL != "https://vbr-prod.example.com" {
		t.Errorf("Expected URL, got %s", resolved.URL)
	}
	if resolved.Username != "prod-admin" {
		t.Errorf("Expected username=prod-admin, got %s", resolved.Username)
	}
	if resolved.Password != "prod-secret" {
		t.Errorf("Expected password=prod-secret, got %s", resolved.Password)
	}
	if resolved.Port != 9419 {
		t.Errorf("Expected port=9419, got %d", resolved.Port)
	}
	if resolved.KeychainKey != "instance:vbr-prod" {
		t.Errorf("Expected keychainKey=instance:vbr-prod, got %s", resolved.KeychainKey)
	}

	// Resolve without credentialRef (uses default env vars)
	resolved, err = ResolveInstance(cfg, "vbr-dr")
	if err != nil {
		t.Fatalf("ResolveInstance(vbr-dr) failed: %v", err)
	}
	if resolved.Username != "default-admin" {
		t.Errorf("Expected username=default-admin, got %s", resolved.Username)
	}
	if resolved.Password != "default-secret" {
		t.Errorf("Expected password=default-secret, got %s", resolved.Password)
	}
	if resolved.KeychainKey != "instance:vbr-dr" {
		t.Errorf("Expected keychainKey=instance:vbr-dr, got %s", resolved.KeychainKey)
	}

	// Resolve nonexistent
	_, err = ResolveInstance(cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent instance")
	}

	// Resolve with credentialRef but missing env vars
	cfgMissing := &VCLIConfig{
		Instances: map[string]InstanceConfig{
			"missing-creds": {
				Product:       "vbr",
				URL:           "https://example.com",
				CredentialRef: "MISSING",
			},
		},
	}
	_, err = ResolveInstance(cfgMissing, "missing-creds")
	if err == nil {
		t.Error("Expected error when credentialRef env vars are not set")
	}
}

func TestGroupConfigWithInstance(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-group-instance-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `apiVersion: owlctl.veeam.com/v1
kind: Config
instances:
  vbr-prod:
    product: vbr
    url: https://vbr-prod.example.com
groups:
  prod-jobs:
    description: Production backup jobs
    instance: vbr-prod
    specs:
      - specs/job1.yaml
      - specs/job2.yaml
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	group, err := cfg.GetGroup("prod-jobs")
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if group.Instance != "vbr-prod" {
		t.Errorf("Expected instance=vbr-prod, got %s", group.Instance)
	}
}

func TestGroupSpecsDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-specsdir-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create specsDir with YAML files
	specsDir := filepath.Join(tmpDir, "specs", "jobs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("Failed to create specs dir: %v", err)
	}
	for _, name := range []string{"job1.yaml", "job2.yaml", "job3.yaml"} {
		if err := os.WriteFile(filepath.Join(specsDir, name), []byte("kind: VBRJob"), 0644); err != nil {
			t.Fatalf("Failed to write spec file: %v", err)
		}
	}
	// Also create a non-YAML file that should be ignored
	if err := os.WriteFile(filepath.Join(specsDir, "notes.txt"), []byte("not a spec"), 0644); err != nil {
		t.Fatalf("Failed to write non-spec file: %v", err)
	}

	configPath := filepath.Join(tmpDir, "owlctl.yaml")
	configYAML := `apiVersion: owlctl.veeam.com/v1
kind: Config
groups:
  auto-specs:
    specsDir: specs/jobs
  mixed-specs:
    specs:
      - explicit/spec.yaml
    specsDir: specs/jobs
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := LoadConfigFrom(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// specsDir only
	autoGroup, err := cfg.GetGroup("auto-specs")
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}
	specs, err := cfg.ResolveGroupSpecs(autoGroup)
	if err != nil {
		t.Fatalf("ResolveGroupSpecs failed: %v", err)
	}
	if len(specs) != 3 {
		t.Errorf("Expected 3 specs from specsDir, got %d: %v", len(specs), specs)
	}

	// Mixed: explicit specs + specsDir
	mixedGroup, err := cfg.GetGroup("mixed-specs")
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}
	specs, err = cfg.ResolveGroupSpecs(mixedGroup)
	if err != nil {
		t.Fatalf("ResolveGroupSpecs failed: %v", err)
	}
	if len(specs) != 4 { // 1 explicit + 3 from specsDir
		t.Errorf("Expected 4 specs (1 explicit + 3 from specsDir), got %d: %v", len(specs), specs)
	}
	// First should be the explicit spec
	if specs[0] != "explicit/spec.yaml" {
		t.Errorf("Expected first spec to be explicit, got %s", specs[0])
	}
}
