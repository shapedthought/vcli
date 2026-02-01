package cmd

import (
	"fmt"
	"reflect"
	"strings"
)

// Drift represents a single configuration drift between state and live VBR
type Drift struct {
	Path   string
	Action string // "modified", "added", "removed"
	State  interface{}
	VBR    interface{}
}

// jobIgnoreFields defines read-only or frequently changing fields to ignore during job drift detection
var jobIgnoreFields = map[string]bool{
	"id":               true,
	"lastRun":          true,
	"nextRun":          true,
	"statistics":       true,
	"creationTime":     true,
	"modificationTime": true,
	"targetType":       true,
	"platform":         true,
	"serverName":       true,
	"isRunning":        true,
	"lastResult":       true,
	"sessionCount":     true,
	"urn":              true,
	"objectId":         true,
	"size":             true,
	"metadata":         true,
}

// repoIgnoreFields defines read-only or frequently changing fields to ignore during repository drift detection
var repoIgnoreFields = map[string]bool{
	"id":       true,
	"uniqueId": true,
}

// repoCriticalPaths defines field paths that should be flagged as CRITICAL severity in drift output
var repoCriticalPaths = map[string]bool{
	"type": true,
}

// sobrIgnoreFields defines read-only or frequently changing fields to ignore during SOBR drift detection
var sobrIgnoreFields = map[string]bool{
	"id":       true,
	"uniqueId": true,
	"status":   true,
}

// sobrCriticalPaths defines field paths flagged as CRITICAL severity for SOBR drift
var sobrCriticalPaths = map[string]bool{
	"isEnabled":                      true,
	"immutabilityMode":               true,
	"daysCount":                      true,
	"movePolicyEnabled":              true,
	"copyPolicyEnabled":              true,
	"performanceExtents":             true,
	"extents":                        true,
	"type":                           true,
	"enforceStrictPlacementPolicy":   true,
}

// encryptionIgnoreFields defines read-only or frequently changing fields to ignore during encryption password drift detection
var encryptionIgnoreFields = map[string]bool{
	"id":               true,
	"uniqueId":         true,
	"modificationTime": true,
}

// encryptionCriticalPaths defines field paths flagged as CRITICAL severity for encryption password drift
var encryptionCriticalPaths = map[string]bool{
	"hint": true,
}

// kmsIgnoreFields defines read-only fields to ignore during KMS server drift detection
var kmsIgnoreFields = map[string]bool{
	"id": true,
}

// kmsCriticalPaths defines field paths flagged as CRITICAL severity for KMS server drift
var kmsCriticalPaths = map[string]bool{
	"type": true,
}

// detectDrift compares state spec against live VBR config, ignoring specified fields
func detectDrift(stateSpec, vbrMap map[string]interface{}, ignore map[string]bool) []Drift {
	var drifts []Drift
	collectDrifts("", stateSpec, vbrMap, &drifts, ignore)
	return drifts
}

func collectDrifts(path string, state, vbr map[string]interface{}, drifts *[]Drift, ignore map[string]bool) {
	// Check all fields in state
	for key, stateValue := range state {
		fullPath := key
		if path != "" {
			fullPath = path + "." + key
		}

		// Skip ignored fields
		if ignore[key] {
			continue
		}

		vbrValue, existsInVBR := vbr[key]

		if !existsInVBR {
			// Field was removed from VBR
			*drifts = append(*drifts, Drift{
				Path:   fullPath,
				Action: "removed",
				State:  stateValue,
				VBR:    nil,
			})
			continue
		}

		// Both exist - check if different (using semantic equality)
		if !valuesEqual(stateValue, vbrValue) {
			// Try recursive comparison for maps
			if stateMap, stateIsMap := stateValue.(map[string]interface{}); stateIsMap {
				if vbrMap, vbrIsMap := vbrValue.(map[string]interface{}); vbrIsMap {
					// Recursively compare nested maps
					collectDrifts(fullPath, stateMap, vbrMap, drifts, ignore)
					continue
				}
			}

			// Values are different
			*drifts = append(*drifts, Drift{
				Path:   fullPath,
				Action: "modified",
				State:  stateValue,
				VBR:    vbrValue,
			})
		}
	}

	// Check for fields added in VBR
	for key, vbrValue := range vbr {
		fullPath := key
		if path != "" {
			fullPath = path + "." + key
		}

		// Skip ignored fields
		if ignore[key] {
			continue
		}

		if _, existsInState := state[key]; !existsInState {
			// Field was added in VBR
			*drifts = append(*drifts, Drift{
				Path:   fullPath,
				Action: "added",
				State:  nil,
				VBR:    vbrValue,
			})
		}
	}
}

// printDrift formats and prints a single drift entry
func printDrift(drift Drift) {
	printDriftWithCritical(drift, nil)
}

// printDriftWithCritical formats and prints a single drift entry, annotating critical paths
func printDriftWithCritical(drift Drift, criticalPaths map[string]bool) {
	severity := ""
	if criticalPaths != nil && isCriticalPath(drift.Path, criticalPaths) {
		severity = "CRITICAL "
	}

	switch drift.Action {
	case "modified":
		stateStr := formatValue(drift.State)
		vbrStr := formatValue(drift.VBR)
		fmt.Printf("  %s~ %s: %s (state) → %s (VBR)\n", severity, drift.Path, stateStr, vbrStr)
	case "removed":
		fmt.Printf("  %s- %s: Removed from VBR\n", severity, drift.Path)
	case "added":
		vbrStr := formatValue(drift.VBR)
		fmt.Printf("  %s+ %s: Added in VBR (value: %s)\n", severity, drift.Path, vbrStr)
	}
}

// isCriticalPath checks if the drift path matches any critical path pattern.
// Matches if the path equals or ends with a critical field name.
func isCriticalPath(path string, criticalPaths map[string]bool) bool {
	// Direct match
	if criticalPaths[path] {
		return true
	}
	// Check if the last segment matches
	parts := strings.Split(path, ".")
	if len(parts) > 0 {
		return criticalPaths[parts[len(parts)-1]]
	}
	return false
}

func formatValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case map[string]interface{}:
		return fmt.Sprintf("{%d fields}", len(v))
	case []interface{}:
		return fmt.Sprintf("[%d items]", len(v))
	case string:
		if len(v) > 50 {
			return fmt.Sprintf("\"%s...\"", v[:47])
		}
		return fmt.Sprintf("\"%s\"", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// valuesEqual compares two values for semantic equality.
// Treats nil and empty arrays/slices as equivalent.
func valuesEqual(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}

	// Check if both are slices/arrays
	aSlice, aIsSlice := a.([]interface{})
	bSlice, bIsSlice := b.([]interface{})

	if aIsSlice && bIsSlice {
		return reflect.DeepEqual(aSlice, bSlice)
	}

	if aIsSlice && b == nil {
		return len(aSlice) == 0
	}

	if a == nil && bIsSlice {
		return len(bSlice) == 0
	}

	// Check if both are maps
	aMap, aIsMap := a.(map[string]interface{})
	bMap, bIsMap := b.(map[string]interface{})

	if aIsMap && bIsMap {
		return reflect.DeepEqual(aMap, bMap)
	}

	if aIsMap && b == nil {
		return len(aMap) == 0
	}

	if a == nil && bIsMap {
		return len(bMap) == 0
	}

	// Fall back to reflect.DeepEqual for other types
	return reflect.DeepEqual(a, b)
}
