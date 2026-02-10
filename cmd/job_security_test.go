package cmd

import (
	"testing"
)

// --- enhanceJobDriftSeverity tests ---

func TestEnhanceJobDriftSeverity_IsDisabled_True(t *testing.T) {
	drifts := []Drift{
		{Path: "isDisabled", Action: "modified", State: false, VBR: true, Severity: SeverityCritical},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityCritical {
		t.Errorf("Job disabled (true) should be CRITICAL, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_IsDisabled_False(t *testing.T) {
	drifts := []Drift{
		{Path: "isDisabled", Action: "modified", State: true, VBR: false, Severity: SeverityCritical},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityWarning {
		t.Errorf("Job re-enabled (false) should be WARNING, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_EncryptionDisabled(t *testing.T) {
	drifts := []Drift{
		{Path: "storage.advancedSettings.storageData.encryption.isEnabled", Action: "modified", State: true, VBR: false},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityCritical {
		t.Errorf("Encryption disabled should be CRITICAL, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_EncryptionEnabled(t *testing.T) {
	drifts := []Drift{
		{Path: "storage.advancedSettings.storageData.encryption.isEnabled", Action: "modified", State: false, VBR: true},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityInfo {
		t.Errorf("Encryption enabled should be INFO, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_RetentionReduced(t *testing.T) {
	drifts := []Drift{
		{Path: "storage.retentionPolicy.quantity", Action: "modified", State: float64(30), VBR: float64(14)},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityCritical {
		t.Errorf("Retention reduced should be CRITICAL, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_RetentionIncreased(t *testing.T) {
	drifts := []Drift{
		{Path: "storage.retentionPolicy.quantity", Action: "modified", State: float64(14), VBR: float64(30)},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityWarning {
		t.Errorf("Retention increased should be WARNING, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_AppAwareDisabled(t *testing.T) {
	drifts := []Drift{
		{Path: "guestProcessing.appAwareProcessing.isEnabled", Action: "modified", State: true, VBR: false},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityWarning {
		t.Errorf("App-aware disabled should be WARNING, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_AppAwareEnabled(t *testing.T) {
	drifts := []Drift{
		{Path: "guestProcessing.appAwareProcessing.isEnabled", Action: "modified", State: false, VBR: true},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityInfo {
		t.Errorf("App-aware enabled should be INFO, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_ScheduleDisabled(t *testing.T) {
	drifts := []Drift{
		{Path: "schedule.daily.isEnabled", Action: "modified", State: true, VBR: false},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityWarning {
		t.Errorf("Schedule disabled should be WARNING, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_ScheduleEnabled(t *testing.T) {
	drifts := []Drift{
		{Path: "schedule.daily.isEnabled", Action: "modified", State: false, VBR: true},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityInfo {
		t.Errorf("Schedule enabled should be INFO, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_RunAutomaticallyDisabled(t *testing.T) {
	drifts := []Drift{
		{Path: "schedule.runAutomatically", Action: "modified", State: true, VBR: false},
	}

	result := enhanceJobDriftSeverity(drifts)
	if result[0].Severity != SeverityWarning {
		t.Errorf("RunAutomatically disabled should be WARNING, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_SkipsNonModified(t *testing.T) {
	drifts := []Drift{
		{Path: "isDisabled", Action: "added", State: nil, VBR: true, Severity: SeverityInfo},
	}

	result := enhanceJobDriftSeverity(drifts)
	// Should not be enhanced (action is "added", not "modified")
	if result[0].Severity != SeverityInfo {
		t.Errorf("Non-modified drift should not be enhanced, got %s", result[0].Severity)
	}
}

func TestEnhanceJobDriftSeverity_UnknownPath(t *testing.T) {
	drifts := []Drift{
		{Path: "description", Action: "modified", State: "old", VBR: "new", Severity: SeverityInfo},
	}

	result := enhanceJobDriftSeverity(drifts)
	// Unknown path should not be enhanced
	if result[0].Severity != SeverityInfo {
		t.Errorf("Unknown path should keep original severity, got %s", result[0].Severity)
	}
}

// --- toBool tests ---

func TestToBool(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string 1", "1", true},
		{"int", 42, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toBool(tt.input); got != tt.want {
				t.Errorf("toBool(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- toFloat64 tests ---

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{"float64", float64(3.14), 3.14},
		{"int", int(42), 42},
		{"string", "3.14", 3.14},
		{"invalid string", "abc", 0},
		{"nil", nil, 0},
		{"bool", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toFloat64(tt.input); got != tt.want {
				t.Errorf("toFloat64(%v) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

// --- extractNestedString tests ---

func TestExtractNestedString(t *testing.T) {
	m := map[string]interface{}{
		"storage": map[string]interface{}{
			"backupRepositoryId": "repo-123",
		},
	}

	got := extractNestedString(m, "storage", "backupRepositoryId")
	if got != "repo-123" {
		t.Errorf("Expected 'repo-123', got %q", got)
	}
}

func TestExtractNestedString_Missing(t *testing.T) {
	m := map[string]interface{}{
		"storage": map[string]interface{}{},
	}

	got := extractNestedString(m, "storage", "backupRepositoryId")
	if got != "" {
		t.Errorf("Expected empty string for missing key, got %q", got)
	}
}

func TestExtractNestedString_NotAMap(t *testing.T) {
	m := map[string]interface{}{
		"storage": "not a map",
	}

	got := extractNestedString(m, "storage", "backupRepositoryId")
	if got != "" {
		t.Errorf("Expected empty string when intermediate is not a map, got %q", got)
	}
}

// --- isHardenedRepo tests ---

func TestIsHardenedRepo_True(t *testing.T) {
	spec := map[string]interface{}{"type": "LinuxHardened"}
	if !isHardenedRepo(spec) {
		t.Error("LinuxHardened should be detected as hardened")
	}
}

func TestIsHardenedRepo_False(t *testing.T) {
	spec := map[string]interface{}{"type": "WinLocal"}
	if isHardenedRepo(spec) {
		t.Error("WinLocal should not be detected as hardened")
	}
}

func TestIsHardenedRepo_NoType(t *testing.T) {
	spec := map[string]interface{}{}
	if isHardenedRepo(spec) {
		t.Error("Missing type should not be detected as hardened")
	}
}

// --- toString tests ---

func TestToString(t *testing.T) {
	if got := toString(nil); got != "" {
		t.Errorf("toString(nil) = %q, want empty", got)
	}
	if got := toString("hello"); got != "hello" {
		t.Errorf("toString(\"hello\") = %q, want \"hello\"", got)
	}
	if got := toString(42); got != "42" {
		t.Errorf("toString(42) = %q, want \"42\"", got)
	}
}
