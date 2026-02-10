package cmd

import (
	"testing"
)

// --- detectDrift tests ---

func TestDetectDrift_NoDrift(t *testing.T) {
	state := map[string]interface{}{
		"name":        "Job1",
		"description": "Test job",
		"isDisabled":  false,
	}
	vbr := map[string]interface{}{
		"name":        "Job1",
		"description": "Test job",
		"isDisabled":  false,
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts, got %d: %+v", len(drifts), drifts)
	}
}

func TestDetectDrift_ModifiedField(t *testing.T) {
	state := map[string]interface{}{
		"description": "Old description",
	}
	vbr := map[string]interface{}{
		"description": "New description",
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 1 {
		t.Fatalf("Expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Path != "description" {
		t.Errorf("Expected path 'description', got %q", drifts[0].Path)
	}
	if drifts[0].Action != "modified" {
		t.Errorf("Expected action 'modified', got %q", drifts[0].Action)
	}
	if drifts[0].State != "Old description" {
		t.Errorf("Expected state 'Old description', got %v", drifts[0].State)
	}
	if drifts[0].VBR != "New description" {
		t.Errorf("Expected VBR 'New description', got %v", drifts[0].VBR)
	}
}

func TestDetectDrift_RemovedField(t *testing.T) {
	state := map[string]interface{}{
		"name":        "Job1",
		"description": "A description",
	}
	vbr := map[string]interface{}{
		"name": "Job1",
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 1 {
		t.Fatalf("Expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Action != "removed" {
		t.Errorf("Expected action 'removed', got %q", drifts[0].Action)
	}
	if drifts[0].Path != "description" {
		t.Errorf("Expected path 'description', got %q", drifts[0].Path)
	}
}

func TestDetectDrift_AddedField(t *testing.T) {
	state := map[string]interface{}{
		"name": "Job1",
	}
	vbr := map[string]interface{}{
		"name":        "Job1",
		"description": "Added in VBR",
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 1 {
		t.Fatalf("Expected 1 drift, got %d", len(drifts))
	}
	if drifts[0].Action != "added" {
		t.Errorf("Expected action 'added', got %q", drifts[0].Action)
	}
	if drifts[0].Path != "description" {
		t.Errorf("Expected path 'description', got %q", drifts[0].Path)
	}
}

func TestDetectDrift_NestedMapDrift(t *testing.T) {
	state := map[string]interface{}{
		"storage": map[string]interface{}{
			"retentionPolicy": map[string]interface{}{
				"quantity": float64(30),
				"type":     "Days",
			},
		},
	}
	vbr := map[string]interface{}{
		"storage": map[string]interface{}{
			"retentionPolicy": map[string]interface{}{
				"quantity": float64(14),
				"type":     "Days",
			},
		},
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 1 {
		t.Fatalf("Expected 1 drift, got %d: %+v", len(drifts), drifts)
	}
	if drifts[0].Path != "storage.retentionPolicy.quantity" {
		t.Errorf("Expected path 'storage.retentionPolicy.quantity', got %q", drifts[0].Path)
	}
}

func TestDetectDrift_IgnoredFields(t *testing.T) {
	state := map[string]interface{}{
		"name":        "Job1",
		"id":          "old-id",
		"description": "Same",
	}
	vbr := map[string]interface{}{
		"name":        "Job1",
		"id":          "new-id",
		"description": "Same",
	}

	ignore := map[string]bool{"id": true}
	drifts := detectDrift(state, vbr, ignore)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts (id should be ignored), got %d: %+v", len(drifts), drifts)
	}
}

func TestDetectDrift_IgnoredNestedFields(t *testing.T) {
	state := map[string]interface{}{
		"virtualMachines": map[string]interface{}{
			"includes": []interface{}{
				map[string]interface{}{
					"name":     "VM1",
					"metadata": "old-meta",
				},
			},
		},
	}
	vbr := map[string]interface{}{
		"virtualMachines": map[string]interface{}{
			"includes": []interface{}{
				map[string]interface{}{
					"name":     "VM1",
					"metadata": "new-meta",
				},
			},
		},
	}

	ignore := map[string]bool{"metadata": true}
	drifts := detectDrift(state, vbr, ignore)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts (metadata in arrays should be ignored), got %d: %+v", len(drifts), drifts)
	}
}

func TestDetectDrift_EmptyMaps(t *testing.T) {
	drifts := detectDrift(map[string]interface{}{}, map[string]interface{}{}, nil)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts for empty maps, got %d", len(drifts))
	}
}

func TestDetectDrift_NumericTypeMismatch(t *testing.T) {
	// JSON produces float64, YAML produces int — should be treated as equal
	state := map[string]interface{}{
		"quantity": 30, // int from YAML
	}
	vbr := map[string]interface{}{
		"quantity": float64(30), // float64 from JSON
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts (int 30 == float64 30), got %d: %+v", len(drifts), drifts)
	}
}

func TestDetectDrift_NilVsEmptySlice(t *testing.T) {
	state := map[string]interface{}{
		"proxyIds": []interface{}{},
	}
	vbr := map[string]interface{}{
		"proxyIds": nil,
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts (empty slice == nil), got %d: %+v", len(drifts), drifts)
	}
}

func TestDetectDrift_NilVsEmptyMap(t *testing.T) {
	state := map[string]interface{}{
		"settings": map[string]interface{}{},
	}
	vbr := map[string]interface{}{
		"settings": nil,
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 0 {
		t.Errorf("Expected 0 drifts (empty map == nil), got %d: %+v", len(drifts), drifts)
	}
}

func TestDetectDrift_ArrayReplacement(t *testing.T) {
	state := map[string]interface{}{
		"days": []interface{}{"Monday", "Tuesday"},
	}
	vbr := map[string]interface{}{
		"days": []interface{}{"Monday", "Wednesday"},
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 1 {
		t.Fatalf("Expected 1 drift for array change, got %d", len(drifts))
	}
	if drifts[0].Action != "modified" {
		t.Errorf("Expected action 'modified', got %q", drifts[0].Action)
	}
}

func TestDetectDrift_TypeChange(t *testing.T) {
	state := map[string]interface{}{
		"value": "a string",
	}
	vbr := map[string]interface{}{
		"value": float64(42),
	}

	drifts := detectDrift(state, vbr, nil)
	if len(drifts) != 1 {
		t.Fatalf("Expected 1 drift for type change, got %d", len(drifts))
	}
	if drifts[0].Action != "modified" {
		t.Errorf("Expected action 'modified', got %q", drifts[0].Action)
	}
}

// --- SeverityMap tests ---

func TestSeverityMap_GetSeverity_FullPath(t *testing.T) {
	sm := SeverityMap{
		"storage.retentionPolicy.quantity": SeverityCritical,
	}
	got := sm.GetSeverity("storage.retentionPolicy.quantity")
	if got != SeverityCritical {
		t.Errorf("Expected CRITICAL, got %s", got)
	}
}

func TestSeverityMap_GetSeverity_LastSegment(t *testing.T) {
	sm := SeverityMap{
		"isDisabled": SeverityCritical,
	}
	got := sm.GetSeverity("some.nested.isDisabled")
	if got != SeverityCritical {
		t.Errorf("Expected CRITICAL via last segment, got %s", got)
	}
}

func TestSeverityMap_GetSeverity_DefaultInfo(t *testing.T) {
	sm := SeverityMap{}
	got := sm.GetSeverity("unknownField")
	if got != SeverityInfo {
		t.Errorf("Expected INFO default, got %s", got)
	}
}

func TestSeverityMap_GetSeverity_FullPathPrecedence(t *testing.T) {
	sm := SeverityMap{
		"storage.retentionPolicy.type": SeverityWarning,
		"type":                         SeverityCritical,
	}
	// Full path should take precedence over last segment
	got := sm.GetSeverity("storage.retentionPolicy.type")
	if got != SeverityWarning {
		t.Errorf("Expected WARNING (full path match), got %s", got)
	}
}

// --- classifyDrifts tests ---

func TestClassifyDrifts(t *testing.T) {
	drifts := []Drift{
		{Path: "isDisabled", Action: "modified"},
		{Path: "description", Action: "modified"},
		{Path: "schedule.daily.isEnabled", Action: "modified"},
	}

	sm := SeverityMap{
		"isDisabled": SeverityCritical,
		"schedule":   SeverityWarning,
	}

	classified := classifyDrifts(drifts, sm)

	if classified[0].Severity != SeverityCritical {
		t.Errorf("isDisabled: expected CRITICAL, got %s", classified[0].Severity)
	}
	if classified[1].Severity != SeverityInfo {
		t.Errorf("description: expected INFO (default), got %s", classified[1].Severity)
	}
	// "schedule.daily.isEnabled" — full path not in map, last segment "isEnabled" not in map → defaults to INFO
	if classified[2].Severity != SeverityInfo {
		t.Errorf("schedule.daily.isEnabled: expected INFO (no match in severity map), got %s", classified[2].Severity)
	}
}

func TestClassifyDrifts_Empty(t *testing.T) {
	result := classifyDrifts(nil, SeverityMap{})
	if result != nil {
		t.Errorf("Expected nil for nil input, got %+v", result)
	}
}

// --- filterDriftsBySeverity tests ---

func TestFilterDriftsBySeverity_Info(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityWarning},
		{Severity: SeverityCritical},
	}

	filtered := filterDriftsBySeverity(drifts, SeverityInfo)
	if len(filtered) != 3 {
		t.Errorf("INFO filter should return all, got %d", len(filtered))
	}
}

func TestFilterDriftsBySeverity_Warning(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityWarning},
		{Severity: SeverityCritical},
	}

	filtered := filterDriftsBySeverity(drifts, SeverityWarning)
	if len(filtered) != 2 {
		t.Errorf("WARNING filter should return 2, got %d", len(filtered))
	}
	for _, d := range filtered {
		if d.Severity == SeverityInfo {
			t.Error("INFO drift should have been filtered out")
		}
	}
}

func TestFilterDriftsBySeverity_Critical(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityWarning},
		{Severity: SeverityCritical},
	}

	filtered := filterDriftsBySeverity(drifts, SeverityCritical)
	if len(filtered) != 1 {
		t.Errorf("CRITICAL filter should return 1, got %d", len(filtered))
	}
	if filtered[0].Severity != SeverityCritical {
		t.Errorf("Expected CRITICAL, got %s", filtered[0].Severity)
	}
}

func TestFilterDriftsBySeverity_Empty(t *testing.T) {
	filtered := filterDriftsBySeverity(nil, SeverityWarning)
	if len(filtered) != 0 {
		t.Errorf("Expected 0 for nil input, got %d", len(filtered))
	}
}

// --- getMaxSeverity tests ---

func TestGetMaxSeverity_Critical(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityCritical},
		{Severity: SeverityWarning},
	}

	if got := getMaxSeverity(drifts); got != SeverityCritical {
		t.Errorf("Expected CRITICAL, got %s", got)
	}
}

func TestGetMaxSeverity_Warning(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityWarning},
	}

	if got := getMaxSeverity(drifts); got != SeverityWarning {
		t.Errorf("Expected WARNING, got %s", got)
	}
}

func TestGetMaxSeverity_InfoOnly(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityInfo},
	}

	if got := getMaxSeverity(drifts); got != SeverityInfo {
		t.Errorf("Expected INFO, got %s", got)
	}
}

func TestGetMaxSeverity_Empty(t *testing.T) {
	if got := getMaxSeverity(nil); got != SeverityInfo {
		t.Errorf("Expected INFO for empty input, got %s", got)
	}
}

// --- exitCodeForDrifts tests ---

func TestExitCodeForDrifts_NoDrifts(t *testing.T) {
	if got := exitCodeForDrifts(nil); got != ExitSuccess {
		t.Errorf("Expected %d (ExitSuccess), got %d", ExitSuccess, got)
	}
}

func TestExitCodeForDrifts_InfoDrift(t *testing.T) {
	drifts := []Drift{{Severity: SeverityInfo}}
	if got := exitCodeForDrifts(drifts); got != ExitDriftWarning {
		t.Errorf("Expected %d (ExitDriftWarning), got %d", ExitDriftWarning, got)
	}
}

func TestExitCodeForDrifts_WarningDrift(t *testing.T) {
	drifts := []Drift{{Severity: SeverityWarning}}
	if got := exitCodeForDrifts(drifts); got != ExitDriftWarning {
		t.Errorf("Expected %d (ExitDriftWarning), got %d", ExitDriftWarning, got)
	}
}

func TestExitCodeForDrifts_CriticalDrift(t *testing.T) {
	drifts := []Drift{{Severity: SeverityCritical}}
	if got := exitCodeForDrifts(drifts); got != ExitDriftCritical {
		t.Errorf("Expected %d (ExitDriftCritical), got %d", ExitDriftCritical, got)
	}
}

func TestExitCodeForDrifts_MixedSeverities(t *testing.T) {
	drifts := []Drift{
		{Severity: SeverityInfo},
		{Severity: SeverityWarning},
		{Severity: SeverityCritical},
	}
	if got := exitCodeForDrifts(drifts); got != ExitDriftCritical {
		t.Errorf("Expected %d (ExitDriftCritical) for mixed, got %d", ExitDriftCritical, got)
	}
}

// --- severityRank tests ---

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		severity Severity
		want     int
	}{
		{SeverityCritical, 3},
		{SeverityWarning, 2},
		{SeverityInfo, 1},
		{Severity("UNKNOWN"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if got := severityRank(tt.severity); got != tt.want {
				t.Errorf("severityRank(%s) = %d, want %d", tt.severity, got, tt.want)
			}
		})
	}
}

// --- valuesEqual tests ---

func TestValuesEqual_NilNil(t *testing.T) {
	if !valuesEqual(nil, nil, nil) {
		t.Error("nil == nil should be true")
	}
}

func TestValuesEqual_NumericTypes(t *testing.T) {
	if !valuesEqual(int(30), float64(30), nil) {
		t.Error("int(30) == float64(30) should be true")
	}
	if !valuesEqual(int64(42), float64(42), nil) {
		t.Error("int64(42) == float64(42) should be true")
	}
	if valuesEqual(int(30), float64(31), nil) {
		t.Error("int(30) != float64(31) should be false")
	}
}

func TestValuesEqual_EmptySliceVsNil(t *testing.T) {
	if !valuesEqual([]interface{}{}, nil, nil) {
		t.Error("empty slice == nil should be true")
	}
	if !valuesEqual(nil, []interface{}{}, nil) {
		t.Error("nil == empty slice should be true")
	}
}

func TestValuesEqual_EmptyMapVsNil(t *testing.T) {
	if !valuesEqual(map[string]interface{}{}, nil, nil) {
		t.Error("empty map == nil should be true")
	}
	if !valuesEqual(nil, map[string]interface{}{}, nil) {
		t.Error("nil == empty map should be true")
	}
}

func TestValuesEqual_IgnoredFieldsInMaps(t *testing.T) {
	a := map[string]interface{}{"name": "VM1", "id": "aaa"}
	b := map[string]interface{}{"name": "VM1", "id": "bbb"}
	ignore := map[string]bool{"id": true}

	if !valuesEqual(a, b, ignore) {
		t.Error("Maps should be equal after ignoring 'id'")
	}
}

func TestValuesEqual_IgnoredFieldsInArrays(t *testing.T) {
	a := []interface{}{
		map[string]interface{}{"name": "VM1", "urn": "old"},
	}
	b := []interface{}{
		map[string]interface{}{"name": "VM1", "urn": "new"},
	}
	ignore := map[string]bool{"urn": true}

	if !valuesEqual(a, b, ignore) {
		t.Error("Arrays should be equal after ignoring 'urn' in elements")
	}
}

// --- noDriftMessage tests ---

func TestNoDriftMessage_NoFilter(t *testing.T) {
	msg := noDriftMessage("Job: Test", SeverityInfo)
	if msg != "No drift detected. Job: Test" {
		t.Errorf("Unexpected message: %s", msg)
	}
}

func TestNoDriftMessage_WithFilter(t *testing.T) {
	msg := noDriftMessage("Job: Test", SeverityWarning)
	if msg == "No drift detected. Job: Test" {
		t.Error("Should indicate filtered results")
	}
}

// --- formatValue tests ---

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil", nil, "nil"},
		{"string", "hello", "\"hello\""},
		{"int", 42, "42"},
		{"bool", true, "true"},
		{"map", map[string]interface{}{"a": 1, "b": 2}, "{2 fields}"},
		{"slice", []interface{}{1, 2, 3}, "[3 items]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.input)
			if got != tt.want {
				t.Errorf("formatValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatValue_LongString(t *testing.T) {
	long := "This is a very long string that exceeds the fifty character limit for truncation"
	got := formatValue(long)
	if len(got) > 55 { // 47 chars + "..." + quotes
		t.Errorf("Long string should be truncated, got length %d: %s", len(got), got)
	}
}

// --- tryParseFloat64 tests ---

func TestTryParseFloat64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantVal float64
		wantOk  bool
	}{
		{"int", int(42), 42, true},
		{"int32", int32(42), 42, true},
		{"int64", int64(42), 42, true},
		{"uint", uint(42), 42, true},
		{"float32", float32(3.14), float64(float32(3.14)), true},
		{"float64", float64(3.14), 3.14, true},
		{"string", "not a number", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := tryParseFloat64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("tryParseFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && val != tt.wantVal {
				t.Errorf("tryParseFloat64(%v) = %f, want %f", tt.input, val, tt.wantVal)
			}
		})
	}
}
