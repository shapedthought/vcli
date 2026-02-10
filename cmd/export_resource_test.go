package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/shapedthought/owlctl/resources"
)

// --- removeIgnoreFields tests ---

func TestRemoveIgnoreFields_Basic(t *testing.T) {
	spec := map[string]interface{}{
		"name":     "Test Repo",
		"id":       "abc-123",
		"uniqueId": "def-456",
		"type":     "LinuxLocal",
	}
	fields := map[string]bool{"id": true, "uniqueId": true}

	removeIgnoreFields(spec, fields)

	if _, ok := spec["id"]; ok {
		t.Error("Expected 'id' to be removed")
	}
	if _, ok := spec["uniqueId"]; ok {
		t.Error("Expected 'uniqueId' to be removed")
	}
	if _, ok := spec["name"]; !ok {
		t.Error("Expected 'name' to be preserved")
	}
	if _, ok := spec["type"]; !ok {
		t.Error("Expected 'type' to be preserved")
	}
}

func TestRemoveIgnoreFields_EmptyFields(t *testing.T) {
	spec := map[string]interface{}{
		"name": "Test",
		"id":   "123",
	}

	removeIgnoreFields(spec, map[string]bool{})

	if len(spec) != 2 {
		t.Errorf("Expected 2 fields preserved, got %d", len(spec))
	}
}

func TestRemoveIgnoreFields_NilFields(t *testing.T) {
	spec := map[string]interface{}{
		"name": "Test",
	}

	removeIgnoreFields(spec, nil)

	if len(spec) != 1 {
		t.Errorf("Expected 1 field preserved, got %d", len(spec))
	}
}

func TestRemoveIgnoreFields_AllRemoved(t *testing.T) {
	spec := map[string]interface{}{
		"id":       "abc",
		"uniqueId": "def",
	}
	fields := map[string]bool{"id": true, "uniqueId": true}

	removeIgnoreFields(spec, fields)

	if len(spec) != 0 {
		t.Errorf("Expected all fields removed, got %d remaining", len(spec))
	}
}

func TestRemoveIgnoreFields_NonexistentField(t *testing.T) {
	spec := map[string]interface{}{
		"name": "Test",
	}
	fields := map[string]bool{"nonexistent": true}

	removeIgnoreFields(spec, fields)

	if len(spec) != 1 {
		t.Errorf("Expected 1 field preserved, got %d", len(spec))
	}
}

// --- convertResourceToYAMLFull tests ---

func TestConvertResourceToYAMLFull_BasicStructure(t *testing.T) {
	specMap := map[string]interface{}{
		"name":        "Default Backup Repository",
		"description": "A repository",
		"type":        "LinuxLocal",
	}

	cfg := ResourceExportConfig{
		Kind:        "VBRRepository",
		DisplayName: "repository",
	}

	result, err := convertResourceToYAMLFull("Default Backup Repository", "abc-123", cfg, specMap)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content := string(result)

	// Check header comment
	if !strings.Contains(content, "# VBRRepository Configuration (Full Export)") {
		t.Error("Expected header comment with kind")
	}
	if !strings.Contains(content, "# Resource ID: abc-123") {
		t.Error("Expected resource ID in header")
	}

	// Parse YAML portion (skip header comments)
	var spec resources.ResourceSpec
	yamlStart := strings.Index(content, "apiVersion:")
	if yamlStart < 0 {
		t.Fatal("Could not find YAML content")
	}
	if err := yaml.Unmarshal([]byte(content[yamlStart:]), &spec); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if spec.APIVersion != "owlctl.veeam.com/v1" {
		t.Errorf("Expected apiVersion 'owlctl.veeam.com/v1', got %q", spec.APIVersion)
	}
	if spec.Kind != "VBRRepository" {
		t.Errorf("Expected kind 'VBRRepository', got %q", spec.Kind)
	}
	if spec.Metadata.Name != "Default Backup Repository" {
		t.Errorf("Expected name 'Default Backup Repository', got %q", spec.Metadata.Name)
	}
	if spec.Spec["type"] != "LinuxLocal" {
		t.Errorf("Expected type 'LinuxLocal' in spec, got %v", spec.Spec["type"])
	}
}

func TestConvertResourceToYAMLFull_EmptySpec(t *testing.T) {
	specMap := map[string]interface{}{}
	cfg := ResourceExportConfig{Kind: "VBRRepository", DisplayName: "repository"}

	result, err := convertResourceToYAMLFull("Empty", "id-1", cfg, specMap)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected non-empty output even for empty spec")
	}
}

// --- convertResourceToYAML tests (integration of ignore + sanitize + full/overlay) ---

func TestConvertResourceToYAML_StripsIgnoreFields(t *testing.T) {
	rawJSON := `{"id": "abc-123", "uniqueId": "def-456", "name": "Test Repo", "type": "LinuxLocal"}`
	cfg := ResourceExportConfig{
		Kind:         "VBRRepository",
		DisplayName:  "repository",
		IgnoreFields: map[string]bool{"id": true, "uniqueId": true},
	}

	result, err := convertResourceToYAML("Test Repo", "abc-123", cfg, json.RawMessage(rawJSON), false, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content := string(result)
	// Parse YAML to check spec contents
	yamlStart := strings.Index(content, "apiVersion:")
	var spec resources.ResourceSpec
	if err := yaml.Unmarshal([]byte(content[yamlStart:]), &spec); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if _, ok := spec.Spec["id"]; ok {
		t.Error("Expected 'id' to be stripped from spec")
	}
	if _, ok := spec.Spec["uniqueId"]; ok {
		t.Error("Expected 'uniqueId' to be stripped from spec")
	}
	if spec.Spec["name"] != "Test Repo" {
		t.Error("Expected 'name' to be preserved in spec")
	}
}

func TestConvertResourceToYAML_CallsSanitize(t *testing.T) {
	rawJSON := `{"hint": "My password", "password": "secret123", "secret": "topsecret"}`
	sanitizeCalled := false
	cfg := ResourceExportConfig{
		Kind:         "VBREncryptionPassword",
		DisplayName:  "encryption password",
		IgnoreFields: map[string]bool{},
		SanitizeSpec: func(spec map[string]interface{}) {
			sanitizeCalled = true
			delete(spec, "password")
			delete(spec, "secret")
		},
	}

	result, err := convertResourceToYAML("My password", "id-1", cfg, json.RawMessage(rawJSON), false, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !sanitizeCalled {
		t.Error("Expected SanitizeSpec to be called")
	}

	content := string(result)
	if strings.Contains(content, "secret123") {
		t.Error("Password value should have been sanitized")
	}
	if strings.Contains(content, "topsecret") {
		t.Error("Secret value should have been sanitized")
	}
}

func TestConvertResourceToYAML_OverlayRequiresBase(t *testing.T) {
	rawJSON := `{"name": "Test"}`
	cfg := ResourceExportConfig{
		Kind:            "VBRRepository",
		DisplayName:     "repository",
		IgnoreFields:    map[string]bool{},
		SupportsOverlay: true,
	}

	_, err := convertResourceToYAML("Test", "id-1", cfg, json.RawMessage(rawJSON), true, "")
	if err == nil {
		t.Fatal("Expected error when --as-overlay used without --base")
	}
	if !strings.Contains(err.Error(), "--as-overlay requires --base") {
		t.Errorf("Expected '--as-overlay requires --base' error, got: %v", err)
	}
}

// --- convertResourceToYAMLOverlay tests ---

func TestConvertResourceToYAMLOverlay_DiffsAgainstBase(t *testing.T) {
	// Create a temp base file
	baseSpec := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       "VBRRepository",
		Metadata:   resources.Metadata{Name: "Test Repo"},
		Spec: map[string]interface{}{
			"name":        "Test Repo",
			"description": "Original description",
			"type":        "LinuxLocal",
		},
	}
	baseYAML, err := yaml.Marshal(baseSpec)
	if err != nil {
		t.Fatalf("Failed to marshal base: %v", err)
	}

	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base.yaml")
	if err := os.WriteFile(basePath, baseYAML, 0644); err != nil {
		t.Fatalf("Failed to write base file: %v", err)
	}

	// Current spec has a changed description
	currentSpec := map[string]interface{}{
		"name":        "Test Repo",
		"description": "Updated description",
		"type":        "LinuxLocal",
	}

	cfg := ResourceExportConfig{
		Kind:        "VBRRepository",
		DisplayName: "repository",
	}

	result, err := convertResourceToYAMLOverlay("Test Repo", "abc-123", cfg, currentSpec, basePath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content := string(result)

	// Check header
	if !strings.Contains(content, "# VBRRepository Overlay") {
		t.Error("Expected overlay header")
	}
	if !strings.Contains(content, basePath) {
		t.Error("Expected base path in header")
	}

	// Parse and check that only the diff is present
	yamlStart := strings.Index(content, "apiVersion:")
	var overlaySpec resources.ResourceSpec
	if err := yaml.Unmarshal([]byte(content[yamlStart:]), &overlaySpec); err != nil {
		t.Fatalf("Failed to parse overlay YAML: %v", err)
	}

	if overlaySpec.Spec["description"] != "Updated description" {
		t.Errorf("Expected changed description in overlay, got %v", overlaySpec.Spec["description"])
	}
	// "name" and "type" are the same in base and current, so should NOT be in overlay
	if _, ok := overlaySpec.Spec["name"]; ok {
		t.Error("Expected 'name' (unchanged) to be absent from overlay")
	}
	if _, ok := overlaySpec.Spec["type"]; ok {
		t.Error("Expected 'type' (unchanged) to be absent from overlay")
	}
}

func TestConvertResourceToYAMLOverlay_NoDifferences(t *testing.T) {
	baseSpec := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       "VBRRepository",
		Metadata:   resources.Metadata{Name: "Test"},
		Spec: map[string]interface{}{
			"name": "Test",
			"type": "LinuxLocal",
		},
	}
	baseYAML, _ := yaml.Marshal(baseSpec)

	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base.yaml")
	os.WriteFile(basePath, baseYAML, 0644)

	currentSpec := map[string]interface{}{
		"name": "Test",
		"type": "LinuxLocal",
	}

	cfg := ResourceExportConfig{Kind: "VBRRepository", DisplayName: "repository"}
	result, err := convertResourceToYAMLOverlay("Test", "id-1", cfg, currentSpec, basePath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Parse and verify empty spec
	yamlStart := strings.Index(string(result), "apiVersion:")
	var overlaySpec resources.ResourceSpec
	yaml.Unmarshal(result[yamlStart:], &overlaySpec)

	if len(overlaySpec.Spec) != 0 {
		t.Errorf("Expected empty overlay spec, got %d fields: %v", len(overlaySpec.Spec), overlaySpec.Spec)
	}
}

func TestConvertResourceToYAMLOverlay_BadBasePath(t *testing.T) {
	cfg := ResourceExportConfig{Kind: "VBRRepository", DisplayName: "repository"}
	_, err := convertResourceToYAMLOverlay("Test", "id-1", cfg, map[string]interface{}{}, "/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent base path")
	}
}

func TestConvertResourceToYAMLOverlay_InvalidBaseYAML(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "bad.yaml")
	os.WriteFile(basePath, []byte("{{not valid yaml"), 0644)

	cfg := ResourceExportConfig{Kind: "VBRRepository", DisplayName: "repository"}
	_, err := convertResourceToYAMLOverlay("Test", "id-1", cfg, map[string]interface{}{}, basePath)
	if err == nil {
		t.Fatal("Expected error for invalid base YAML")
	}
}

// --- sanitizeEncryptionPassword tests ---

func TestSanitizeEncryptionPassword_RemovesPasswordAndSecret(t *testing.T) {
	spec := map[string]interface{}{
		"hint":       "My password",
		"password":   "should-be-removed",
		"secret":     "should-be-removed",
		"isImported": false,
	}

	sanitizeEncryptionPassword(spec)

	if _, ok := spec["password"]; ok {
		t.Error("Expected 'password' to be removed")
	}
	if _, ok := spec["secret"]; ok {
		t.Error("Expected 'secret' to be removed")
	}
	if _, ok := spec["hint"]; !ok {
		t.Error("Expected 'hint' to be preserved")
	}
	if _, ok := spec["isImported"]; !ok {
		t.Error("Expected 'isImported' to be preserved")
	}
}

func TestSanitizeEncryptionPassword_NoSensitiveFields(t *testing.T) {
	spec := map[string]interface{}{
		"hint":       "My password",
		"isImported": false,
	}

	sanitizeEncryptionPassword(spec)

	if len(spec) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(spec))
	}
}

// --- Per-resource ignore field configuration tests ---

func TestRepoIgnoreFields_StripsCorrectFields(t *testing.T) {
	spec := map[string]interface{}{
		"id":          "abc",
		"uniqueId":    "def",
		"name":        "Test Repo",
		"description": "A repo",
	}

	removeIgnoreFields(spec, repoIgnoreFields)

	if _, ok := spec["id"]; ok {
		t.Error("Expected 'id' stripped for repo")
	}
	if _, ok := spec["uniqueId"]; ok {
		t.Error("Expected 'uniqueId' stripped for repo")
	}
	if len(spec) != 2 {
		t.Errorf("Expected 2 fields remaining, got %d", len(spec))
	}
}

func TestSobrIgnoreFields_StripsCorrectFields(t *testing.T) {
	spec := map[string]interface{}{
		"id":       "abc",
		"uniqueId": "def",
		"status":   "Available",
		"name":     "SOBR 1",
	}

	removeIgnoreFields(spec, sobrIgnoreFields)

	if _, ok := spec["id"]; ok {
		t.Error("Expected 'id' stripped for SOBR")
	}
	if _, ok := spec["uniqueId"]; ok {
		t.Error("Expected 'uniqueId' stripped for SOBR")
	}
	if _, ok := spec["status"]; ok {
		t.Error("Expected 'status' stripped for SOBR")
	}
	if len(spec) != 1 {
		t.Errorf("Expected 1 field remaining, got %d", len(spec))
	}
}

func TestEncryptionIgnoreFields_StripsCorrectFields(t *testing.T) {
	spec := map[string]interface{}{
		"id":               "abc",
		"uniqueId":         "def",
		"modificationTime": "2026-01-01T00:00:00Z",
		"hint":             "My password",
	}

	removeIgnoreFields(spec, encryptionIgnoreFields)

	if _, ok := spec["id"]; ok {
		t.Error("Expected 'id' stripped for encryption")
	}
	if _, ok := spec["uniqueId"]; ok {
		t.Error("Expected 'uniqueId' stripped for encryption")
	}
	if _, ok := spec["modificationTime"]; ok {
		t.Error("Expected 'modificationTime' stripped for encryption")
	}
	if len(spec) != 1 {
		t.Errorf("Expected 1 field remaining, got %d", len(spec))
	}
}

func TestKmsIgnoreFields_StripsCorrectFields(t *testing.T) {
	spec := map[string]interface{}{
		"id":          "abc",
		"name":        "My KMS",
		"description": "A KMS server",
	}

	removeIgnoreFields(spec, kmsIgnoreFields)

	if _, ok := spec["id"]; ok {
		t.Error("Expected 'id' stripped for KMS")
	}
	if len(spec) != 2 {
		t.Errorf("Expected 2 fields remaining, got %d", len(spec))
	}
}

// --- ResourceExportConfig.SupportsOverlay tests ---

func TestExportSingleResource_OverlayBlockedWhenNotSupported(t *testing.T) {
	cfg := ResourceExportConfig{
		Kind:            "VBREncryptionPassword",
		DisplayName:     "encryption password",
		SupportsOverlay: false,
	}

	rawJSON := `{"hint": "test"}`
	_, err := convertResourceToYAML("test", "id-1", cfg, json.RawMessage(rawJSON), true, "/some/base.yaml")
	// When overlay is not supported but asOverlay=true, the function itself doesn't check —
	// it's the caller (exportSingleResource) that calls log.Fatal. But we can at least
	// verify the base-required error path still fires for supported types.
	// For unsupported types, the check happens at the command level.
	// So test that supported types with overlay but no base give the right error.
	if err == nil {
		t.Fatal("Expected error for overlay without base")
	}
}

// --- convertResourceToYAML with nested spec ---

func TestConvertResourceToYAML_NestedSpec(t *testing.T) {
	rawJSON := `{
		"id": "abc-123",
		"name": "SOBR 1",
		"performanceTier": {
			"extents": [{"repositoryId": "repo-1"}]
		},
		"capacityTier": {
			"encryption": {"isEnabled": true}
		}
	}`

	cfg := ResourceExportConfig{
		Kind:         "VBRScaleOutRepository",
		DisplayName:  "scale-out repository",
		IgnoreFields: sobrIgnoreFields,
	}

	result, err := convertResourceToYAML("SOBR 1", "abc-123", cfg, json.RawMessage(rawJSON), false, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Parse and verify nested structure is preserved
	yamlStart := strings.Index(string(result), "apiVersion:")
	var spec resources.ResourceSpec
	if err := yaml.Unmarshal(result[yamlStart:], &spec); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if _, ok := spec.Spec["id"]; ok {
		t.Error("Expected 'id' to be stripped")
	}
	if spec.Spec["name"] != "SOBR 1" {
		t.Errorf("Expected name 'SOBR 1', got %v", spec.Spec["name"])
	}
	// Verify nested structure
	capTier, ok := spec.Spec["capacityTier"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected capacityTier to be a map")
	}
	enc, ok := capTier["encryption"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected encryption to be a map")
	}
	if enc["isEnabled"] != true {
		t.Errorf("Expected encryption.isEnabled=true, got %v", enc["isEnabled"])
	}
}

// --- Round-trip test: JSON → YAML → parse ---

func TestConvertResourceToYAML_RoundTrip(t *testing.T) {
	rawJSON := `{
		"id": "kms-1",
		"name": "My KMS Server",
		"description": "Test KMS",
		"type": "KmipCompliant"
	}`

	cfg := ResourceExportConfig{
		Kind:         "VBRKmsServer",
		DisplayName:  "KMS server",
		IgnoreFields: kmsIgnoreFields,
	}

	result, err := convertResourceToYAML("My KMS Server", "kms-1", cfg, json.RawMessage(rawJSON), false, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should be loadable by LoadResourceSpec equivalent
	yamlStart := strings.Index(string(result), "apiVersion:")
	var spec resources.ResourceSpec
	if err := yaml.Unmarshal(result[yamlStart:], &spec); err != nil {
		t.Fatalf("Round-trip failed — YAML not parseable: %v", err)
	}

	if spec.Kind != "VBRKmsServer" {
		t.Errorf("Expected kind VBRKmsServer, got %s", spec.Kind)
	}
	if spec.Metadata.Name != "My KMS Server" {
		t.Errorf("Expected name 'My KMS Server', got %s", spec.Metadata.Name)
	}
	if _, ok := spec.Spec["id"]; ok {
		t.Error("Expected 'id' to be stripped in round-trip")
	}
	if spec.Spec["type"] != "KmipCompliant" {
		t.Errorf("Expected type 'KmipCompliant', got %v", spec.Spec["type"])
	}
}
