package cmd

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

// Severity represents the security severity level of a drift
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

// Drift represents a single configuration drift between state and live VBR
type Drift struct {
	Path     string
	Action   string // "modified", "added", "removed"
	State    interface{}
	VBR      interface{}
	Severity Severity
}

// SeverityMap maps field path segments to severity levels.
// Fields not listed default to INFO.
type SeverityMap map[string]Severity

// GetSeverity returns the severity for a drift path.
// Checks full path first, then last segment. Defaults to INFO.
func (sm SeverityMap) GetSeverity(path string) Severity {
	if s, ok := sm[path]; ok {
		return s
	}
	parts := strings.Split(path, ".")
	if len(parts) > 0 {
		if s, ok := sm[parts[len(parts)-1]]; ok {
			return s
		}
	}
	return SeverityInfo
}

// severityRank returns numeric rank for severity comparison (higher = more severe)
func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 3
	case SeverityWarning:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// --- Shared severity filter flags ---

var (
	severityFilter string
	securityOnly   bool
)

// addSeverityFlags registers --severity and --security-only flags on a diff command
func addSeverityFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&severityFilter, "severity", "", "Minimum severity to show (critical, warning, info)")
	cmd.Flags().BoolVar(&securityOnly, "security-only", false, "Show only WARNING and CRITICAL drifts (alias for --severity warning)")
}

// parseSeverityFlag returns the minimum severity from flags.
// --severity takes precedence over --security-only.
func parseSeverityFlag() Severity {
	if severityFilter != "" {
		switch strings.ToLower(severityFilter) {
		case "critical":
			return SeverityCritical
		case "warning":
			return SeverityWarning
		case "info":
			return SeverityInfo
		default:
			log.Fatalf("Invalid severity: %s (use critical, warning, or info)", severityFilter)
		}
	}
	if securityOnly {
		return SeverityWarning
	}
	return SeverityInfo
}

// --- Per-resource ignore fields ---

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

// sobrIgnoreFields defines read-only or frequently changing fields to ignore during SOBR drift detection
var sobrIgnoreFields = map[string]bool{
	"id":       true,
	"uniqueId": true,
	"status":   true,
}

// encryptionIgnoreFields defines read-only or frequently changing fields to ignore during encryption password drift detection
var encryptionIgnoreFields = map[string]bool{
	"id":               true,
	"uniqueId":         true,
	"modificationTime": true,
}

// kmsIgnoreFields defines read-only fields to ignore during KMS server drift detection
var kmsIgnoreFields = map[string]bool{
	"id": true,
}

// --- Per-resource severity maps ---

// jobSeverityMap classifies job drift fields by severity
var jobSeverityMap = SeverityMap{
	// CRITICAL — changes that directly affect data protection
	"isDisabled":         SeverityCritical,
	"retentionPolicy":    SeverityCritical,
	"retainCycles":       SeverityCritical,
	"gfsPolicy":          SeverityCritical,
	"backupRepositoryId": SeverityCritical,
	// CRITICAL — full dotted paths for nested fields
	"storage.advancedSettings.storageData.encryption.isEnabled": SeverityCritical,
	"storage.retentionPolicy.quantity":                          SeverityCritical,
	"storage.backupRepositoryId":                                SeverityCritical,
	// WARNING — changes that weaken defense-in-depth
	"guestProcessing": SeverityWarning,
	"schedule":        SeverityWarning,
	"encryption":      SeverityWarning,
	"healthCheck":     SeverityWarning,
	// WARNING — full dotted paths for nested fields
	"storage.retentionPolicy.type":                 SeverityWarning,
	"guestProcessing.appAwareProcessing.isEnabled": SeverityWarning,
	"schedule.daily.isEnabled":                     SeverityWarning,
	"schedule.runAutomatically":                    SeverityWarning,
}

// repoSeverityMap classifies repository drift fields by severity
var repoSeverityMap = SeverityMap{
	// CRITICAL — repository type changes and immutability
	"type": SeverityCritical,
	"repository.makeRecentBackupsImmutableDays": SeverityCritical,
	"makeRecentBackupsImmutableDays":            SeverityCritical, // fallback for last segment match

	// CRITICAL — security-relevant advanced settings
	"repository.advancedSettings.decompressBeforeStoring": SeverityCritical,
	"decompressBeforeStoring":                             SeverityCritical,
	"repository.advancedSettings.perVmBackup":             SeverityCritical,
	"perVmBackup":                                         SeverityCritical,

	// WARNING — operational changes with security relevance
	"path":         SeverityWarning,
	"maxTaskCount": SeverityWarning,
	"repository.advancedSettings.alignDataBlocks": SeverityWarning,
	"alignDataBlocks":                             SeverityWarning,
	"repository.readWriteLimitEnabled":            SeverityWarning,
	"readWriteLimitEnabled":                       SeverityWarning,
	"repository.readWriteRate":                    SeverityWarning,
	"readWriteRate":                               SeverityWarning,
	"repository.taskLimitEnabled":                 SeverityWarning,
	"taskLimitEnabled":                            SeverityWarning,
}

// sobrSeverityMap classifies SOBR drift fields by severity
var sobrSeverityMap = SeverityMap{
	// CRITICAL — SOBR availability and immutability
	"isEnabled":                    SeverityCritical,
	"immutabilityMode":             SeverityCritical,
	"type":                         SeverityCritical,
	"enforceStrictPlacementPolicy": SeverityCritical,

	// CRITICAL — capacity tier encryption
	"capacityTier.encryption":          SeverityCritical,
	"capacityTier.encryption.isEnabled": SeverityCritical,
	"encryption":                       SeverityCritical, // fallback for last segment match

	// WARNING — backup health monitoring
	"capacityTier.backupHealth":          SeverityWarning,
	"capacityTier.backupHealth.isEnabled": SeverityWarning,
	"backupHealth":                       SeverityWarning, // fallback for last segment match

	// WARNING — policy and tier changes
	"movePolicyEnabled":  SeverityWarning,
	"copyPolicyEnabled":  SeverityWarning,
	"daysCount":          SeverityWarning,
	"performanceExtents": SeverityWarning,
	"extents":            SeverityWarning,
}

// encryptionSeverityMap classifies encryption password drift fields by severity
var encryptionSeverityMap = SeverityMap{
	"hint": SeverityWarning,
}

// kmsSeverityMap classifies KMS server drift fields by severity
var kmsSeverityMap = SeverityMap{
	// CRITICAL — KMS type changes
	"type": SeverityCritical,

	// WARNING — server hostname and description changes
	"name":        SeverityWarning, // Server hostname/address
	"description": SeverityWarning,
}

// --- Drift detection ---

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
		if !valuesEqual(stateValue, vbrValue, ignore) {
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

// --- Severity classification and filtering ---

// classifyDrifts assigns severity to each drift based on the resource-specific severity map
func classifyDrifts(drifts []Drift, sm SeverityMap) []Drift {
	for i := range drifts {
		drifts[i].Severity = sm.GetSeverity(drifts[i].Path)
	}
	return drifts
}

// filterDriftsBySeverity returns only drifts at or above the minimum severity level
func filterDriftsBySeverity(drifts []Drift, minSeverity Severity) []Drift {
	if minSeverity == SeverityInfo {
		return drifts
	}
	minRank := severityRank(minSeverity)
	var filtered []Drift
	for _, d := range drifts {
		if severityRank(d.Severity) >= minRank {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// getMaxSeverity returns the highest severity among the given drifts
func getMaxSeverity(drifts []Drift) Severity {
	max := SeverityInfo
	for _, d := range drifts {
		if severityRank(d.Severity) > severityRank(max) {
			max = d.Severity
		}
	}
	return max
}

// exitCodeForDrifts returns the appropriate exit code:
//
//	ExitSuccess (0) = no drift
//	ExitDriftWarning (3) = INFO/WARNING drift
//	ExitDriftCritical (4) = CRITICAL drift
func exitCodeForDrifts(drifts []Drift) int {
	if len(drifts) == 0 {
		return ExitSuccess
	}
	if getMaxSeverity(drifts) == SeverityCritical {
		return ExitDriftCritical
	}
	return ExitDriftWarning
}

// noDriftMessage returns the appropriate "no drift" message, accounting for severity filtering
func noDriftMessage(resourceDesc string, minSev Severity) string {
	if minSev != SeverityInfo {
		return fmt.Sprintf("No %s or higher drift detected. %s (lower severity drifts may exist.)", minSev, resourceDesc)
	}
	return fmt.Sprintf("No drift detected. %s", resourceDesc)
}

// --- Drift printing ---

// printDriftWithSeverity prints a single drift entry with its severity label
func printDriftWithSeverity(drift Drift) {
	sev := string(drift.Severity)

	switch drift.Action {
	case "modified":
		stateStr := formatValue(drift.State)
		vbrStr := formatValue(drift.VBR)
		fmt.Printf("  %s ~ %s: %s (state) -> %s (VBR)\n", sev, drift.Path, stateStr, vbrStr)
	case "removed":
		fmt.Printf("  %s - %s: Removed from VBR\n", sev, drift.Path)
	case "added":
		vbrStr := formatValue(drift.VBR)
		fmt.Printf("  %s + %s: Added in VBR (value: %s)\n", sev, drift.Path, vbrStr)
	}
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
// Filters ignored fields from nested maps and arrays before comparison.
func valuesEqual(a, b interface{}, ignore map[string]bool) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}

	// Check if both are slices/arrays
	aSlice, aIsSlice := a.([]interface{})
	bSlice, bIsSlice := b.([]interface{})

	if aIsSlice && bIsSlice {
		// Filter ignored fields from array elements before comparing
		cleanedA := filterIgnoredFieldsFromArray(aSlice, ignore)
		cleanedB := filterIgnoredFieldsFromArray(bSlice, ignore)
		return reflect.DeepEqual(cleanedA, cleanedB)
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
		// Filter ignored fields from maps before comparing
		cleanedA := filterIgnoredFieldsFromMap(aMap, ignore)
		cleanedB := filterIgnoredFieldsFromMap(bMap, ignore)
		return reflect.DeepEqual(cleanedA, cleanedB)
	}

	if aIsMap && b == nil {
		return len(aMap) == 0
	}

	if a == nil && bIsMap {
		return len(bMap) == 0
	}

	// Handle numeric comparisons (int, float64, etc.)
	// JSON unmarshaling can produce different numeric types
	aNum, aIsNum := tryParseFloat64(a)
	bNum, bIsNum := tryParseFloat64(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	// Fall back to reflect.DeepEqual for other types
	return reflect.DeepEqual(a, b)
}

// filterIgnoredFieldsFromArray recursively filters ignored fields from array elements
func filterIgnoredFieldsFromArray(arr []interface{}, ignore map[string]bool) []interface{} {
	if arr == nil {
		return nil
	}

	cleaned := make([]interface{}, len(arr))
	for i, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// Recursively filter maps inside arrays
			cleaned[i] = filterIgnoredFieldsFromMap(itemMap, ignore)
		} else if itemArr, ok := item.([]interface{}); ok {
			// Recursively filter nested arrays
			cleaned[i] = filterIgnoredFieldsFromArray(itemArr, ignore)
		} else {
			cleaned[i] = item
		}
	}
	return cleaned
}

// filterIgnoredFieldsFromMap recursively filters ignored fields from a map
func filterIgnoredFieldsFromMap(m map[string]interface{}, ignore map[string]bool) map[string]interface{} {
	if m == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range m {
		// Skip ignored fields
		if ignore[k] {
			continue
		}

		// Recursively handle nested structures
		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[k] = filterIgnoredFieldsFromMap(nestedMap, ignore)
		} else if nestedArr, ok := v.([]interface{}); ok {
			result[k] = filterIgnoredFieldsFromArray(nestedArr, ignore)
		} else {
			result[k] = v
		}
	}
	return result
}

// tryParseFloat64 attempts to convert a value to float64
// Returns (value, true) if successful, (0, false) otherwise
func tryParseFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}
