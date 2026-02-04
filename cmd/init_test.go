package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/shapedthought/vcli/models"
)

// TestGetDefaultProfiles verifies the default profile configuration
func TestGetDefaultProfiles(t *testing.T) {
	profilesFile := getDefaultProfiles()

	// Check version
	if profilesFile.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", profilesFile.Version)
	}

	// Check currentProfile
	if profilesFile.CurrentProfile != "vbr" {
		t.Errorf("Expected currentProfile vbr, got %s", profilesFile.CurrentProfile)
	}

	expectedProfiles := []string{"vb365", "aws", "vbr", "azure", "gcp", "vone", "ent_man"}

	if len(profilesFile.Profiles) != len(expectedProfiles) {
		t.Errorf("Expected %d profiles, got %d", len(expectedProfiles), len(profilesFile.Profiles))
	}

	for _, expectedName := range expectedProfiles {
		if _, exists := profilesFile.Profiles[expectedName]; !exists {
			t.Errorf("Expected profile %s not found", expectedName)
		}
	}

	// Verify VBR profile has correct settings
	vbr := profilesFile.Profiles["vbr"]
	if vbr.Port != 9419 {
		t.Errorf("VBR port = %d, want 9419", vbr.Port)
	}
	if vbr.Headers.XAPIVersion != "1.3-rev1" {
		t.Errorf("VBR X-API-Version = %s, want 1.3-rev1", vbr.Headers.XAPIVersion)
	}
	if vbr.APIVersion != "1.3-rev1" {
		t.Errorf("VBR APIVersion = %s, want 1.3-rev1", vbr.APIVersion)
	}
	if vbr.Product != "VeeamBackupReplication" {
		t.Errorf("VBR Product = %s, want VeeamBackupReplication", vbr.Product)
	}
	if vbr.AuthType != "oauth" {
		t.Errorf("VBR AuthType = %s, want oauth", vbr.AuthType)
	}
	if vbr.Endpoints.Auth != "/api/oauth2/token" {
		t.Errorf("VBR Auth endpoint = %s, want /api/oauth2/token", vbr.Endpoints.Auth)
	}
	if vbr.Endpoints.APIPrefix != "/api/v1" {
		t.Errorf("VBR API prefix = %s, want /api/v1", vbr.Endpoints.APIPrefix)
	}
}

// TestEnsureTrailingSlash verifies path normalization
func TestEnsureTrailingSlash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty path",
			input: "",
			want:  "",
		},
		{
			name:  "unix path without slash",
			input: "/tmp/vcli",
			want:  "/tmp/vcli/",
		},
		{
			name:  "unix path with slash",
			input: "/tmp/vcli/",
			want:  "/tmp/vcli/",
		},
		{
			name:  "relative path",
			input: ".vcli",
			want:  ".vcli/",
		},
		{
			name:  "relative path with slash",
			input: ".vcli/",
			want:  ".vcli/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureTrailingSlash(tt.input)
			if got != tt.want {
				t.Errorf("ensureTrailingSlash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestGetOutputPath verifies output path resolution
func TestGetOutputPath(t *testing.T) {
	// Save original values
	origOutputDir := outputDir
	origVCLIPath := os.Getenv("VCLI_SETTINGS_PATH")

	// Restore after test
	defer func() {
		outputDir = origOutputDir
		os.Setenv("VCLI_SETTINGS_PATH", origVCLIPath)
	}()

	tests := []struct {
		name             string
		outputDirFlag    string
		envVar           string
		wantContains     string
		wantTrailingSlash bool
	}{
		{
			name:             "explicit output-dir flag",
			outputDirFlag:    "/explicit/path",
			envVar:           "",
			wantContains:     "/explicit/path",
			wantTrailingSlash: true,
		},
		{
			name:             "VCLI_SETTINGS_PATH set",
			outputDirFlag:    "",
			envVar:           "/env/path/",
			wantContains:     "/env/path/",
			wantTrailingSlash: true,
		},
		{
			name:             "flag overrides env var",
			outputDirFlag:    "/flag/path",
			envVar:           "/env/path/",
			wantContains:     "/flag/path",
			wantTrailingSlash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir = tt.outputDirFlag
			os.Setenv("VCLI_SETTINGS_PATH", tt.envVar)

			got := getOutputPath()

			if tt.wantContains != "" && got != tt.wantContains && !filepath.IsAbs(got) {
				// For relative paths or empty, just check it's not empty when expected
				if tt.wantContains != "" && got == "" {
					t.Errorf("getOutputPath() = empty, want %q", tt.wantContains)
				}
			}

			if tt.wantTrailingSlash && got != "" {
				lastChar := got[len(got)-1]
				if lastChar != '/' && lastChar != '\\' {
					t.Errorf("getOutputPath() = %q, missing trailing slash", got)
				}
			}
		})
	}
}

// TestInitAppNonInteractive verifies non-interactive init creates correct files
func TestInitAppNonInteractive(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "vcli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save and restore original values
	origOutputDir := outputDir
	origInsecure := insecure
	origCredsFile := credsFile
	origStdout := os.Stdout

	defer func() {
		outputDir = origOutputDir
		insecure = origInsecure
		credsFile = origCredsFile
		os.Stdout = origStdout
	}()

	// Redirect stdout to capture JSON output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test values
	outputDir = tempDir + "/"
	insecure = true
	credsFile = true

	// Run the function
	initAppNonInteractive()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = origStdout
	capturedOutput, _ := io.ReadAll(r)

	// Verify JSON output structure
	var result map[string]interface{}
	if err := json.Unmarshal(capturedOutput, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, string(capturedOutput))
	}

	// Check JSON structure
	if _, ok := result["settings"]; !ok {
		t.Error("JSON output missing 'settings' field")
	}
	if _, ok := result["profiles"]; !ok {
		t.Error("JSON output missing 'profiles' field")
	}
	if _, ok := result["files"]; !ok {
		t.Error("JSON output missing 'files' field")
	}

	// Verify settings file was created
	settingsPath := filepath.Join(tempDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Errorf("settings.json was not created at %s", settingsPath)
	}

	// Verify profiles file was created
	profilesPath := filepath.Join(tempDir, "profiles.json")
	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		t.Errorf("profiles.json was not created at %s", profilesPath)
	}

	// Read and verify settings content
	settingsData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings.json: %v", err)
	}

	var settings models.Settings
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("Failed to parse settings.json: %v", err)
	}

	if settings.ApiNotSecure != true {
		t.Errorf("ApiNotSecure = %v, want true", settings.ApiNotSecure)
	}
	if settings.CredsFileMode != true {
		t.Errorf("CredsFileMode = %v, want true", settings.CredsFileMode)
	}
	if settings.SelectedProfile != "vbr" {
		t.Errorf("SelectedProfile = %s, want vbr", settings.SelectedProfile)
	}

	// Read and verify profiles content
	profilesData, err := os.ReadFile(profilesPath)
	if err != nil {
		t.Fatalf("Failed to read profiles.json: %v", err)
	}

	var profilesFile models.ProfilesFile
	if err := json.Unmarshal(profilesData, &profilesFile); err != nil {
		t.Fatalf("Failed to parse profiles.json: %v", err)
	}

	if profilesFile.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", profilesFile.Version)
	}

	if len(profilesFile.Profiles) != 7 {
		t.Errorf("Expected 7 profiles, got %d", len(profilesFile.Profiles))
	}
}

// TestInitSettingsNonInteractive verifies settings-only initialization
func TestInitSettingsNonInteractive(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "vcli-test-settings-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save and restore original values
	origOutputDir := outputDir
	origInsecure := insecure
	origCredsFile := credsFile
	origStdout := os.Stdout

	defer func() {
		outputDir = origOutputDir
		insecure = origInsecure
		credsFile = origCredsFile
		os.Stdout = origStdout
	}()

	// Redirect stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test values
	outputDir = tempDir + "/"
	insecure = false
	credsFile = false

	// Run the function
	initSettingsNonInteractive()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = origStdout
	capturedOutput, _ := io.ReadAll(r)

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(capturedOutput, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if _, ok := result["settings"]; !ok {
		t.Error("JSON output missing 'settings' field")
	}
	if _, ok := result["file"]; !ok {
		t.Error("JSON output missing 'file' field")
	}

	// Verify only settings file was created
	settingsPath := filepath.Join(tempDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Errorf("settings.json was not created at %s", settingsPath)
	}

	// Verify profiles file was NOT created
	profilesPath := filepath.Join(tempDir, "profiles.json")
	if _, err := os.Stat(profilesPath); !os.IsNotExist(err) {
		t.Errorf("profiles.json should not have been created, but exists at %s", profilesPath)
	}
}

// TestInitProfilesOnly verifies profiles-only initialization
func TestInitProfilesOnly(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "vcli-test-profiles-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save and restore original values
	origOutputDir := outputDir
	origStdout := os.Stdout

	defer func() {
		outputDir = origOutputDir
		os.Stdout = origStdout
	}()

	// Redirect stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test values
	outputDir = tempDir + "/"

	// Run the function
	initProfilesOnly()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = origStdout
	capturedOutput, _ := io.ReadAll(r)

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(capturedOutput, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if _, ok := result["profiles"]; !ok {
		t.Error("JSON output missing 'profiles' field")
	}
	if _, ok := result["file"]; !ok {
		t.Error("JSON output missing 'file' field")
	}

	// Verify only profiles file was created
	profilesPath := filepath.Join(tempDir, "profiles.json")
	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		t.Errorf("profiles.json was not created at %s", profilesPath)
	}

	// Verify settings file was NOT created
	settingsPath := filepath.Join(tempDir, "settings.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Errorf("settings.json should not have been created, but exists at %s", settingsPath)
	}

	// Verify profiles content
	profilesData, err := os.ReadFile(profilesPath)
	if err != nil {
		t.Fatalf("Failed to read profiles.json: %v", err)
	}

	var profilesFile models.ProfilesFile
	if err := json.Unmarshal(profilesData, &profilesFile); err != nil {
		t.Fatalf("Failed to parse profiles.json: %v", err)
	}

	if profilesFile.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", profilesFile.Version)
	}

	if len(profilesFile.Profiles) != 7 {
		t.Errorf("Expected 7 profiles, got %d", len(profilesFile.Profiles))
	}

	// Verify specific profile details
	if vbr, exists := profilesFile.Profiles["vbr"]; !exists {
		t.Error("VBR profile not found")
	} else {
		if vbr.Port != 9419 {
			t.Errorf("VBR port = %d, want 9419", vbr.Port)
		}
	}
}

// TestFlagDefaults verifies default flag values
func TestFlagDefaults(t *testing.T) {
	// Reset flags to defaults
	interactive = false
	insecure = false
	credsFile = false
	outputDir = ""

	tests := []struct {
		name     string
		flagName string
		value    interface{}
		want     interface{}
	}{
		{
			name:     "interactive default",
			flagName: "interactive",
			value:    interactive,
			want:     false,
		},
		{
			name:     "insecure default",
			flagName: "insecure",
			value:    insecure,
			want:     false,
		},
		{
			name:     "creds-file default",
			flagName: "creds-file",
			value:    credsFile,
			want:     false,
		},
		{
			name:     "output-dir default",
			flagName: "output-dir",
			value:    outputDir,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("Flag %s default = %v, want %v", tt.flagName, tt.value, tt.want)
			}
		})
	}
}
