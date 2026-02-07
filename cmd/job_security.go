package cmd

import (
	"fmt"
	"strconv"

	"github.com/shapedthought/owlctl/state"
)

// enhanceJobDriftSeverity post-processes classified drifts, upgrading or
// downgrading severity based on the direction of value change.
// This keeps the core drift engine generic while adding job-specific intelligence.
func enhanceJobDriftSeverity(drifts []Drift) []Drift {
	for i := range drifts {
		if drifts[i].Action != "modified" {
			continue
		}

		switch drifts[i].Path {
		case "isDisabled":
			// CRITICAL only if job was disabled (true); WARNING if re-enabled (false)
			if toBool(drifts[i].VBR) {
				drifts[i].Severity = SeverityCritical
			} else {
				drifts[i].Severity = SeverityWarning
			}

		case "storage.advancedSettings.storageData.encryption.isEnabled":
			// CRITICAL if encryption disabled; INFO if enabled
			if !toBool(drifts[i].VBR) {
				drifts[i].Severity = SeverityCritical
			} else {
				drifts[i].Severity = SeverityInfo
			}

		case "storage.retentionPolicy.quantity":
			// CRITICAL if retention reduced; WARNING if increased
			stateVal := toFloat64(drifts[i].State)
			vbrVal := toFloat64(drifts[i].VBR)
			if vbrVal < stateVal {
				drifts[i].Severity = SeverityCritical
			} else {
				drifts[i].Severity = SeverityWarning
			}

		case "guestProcessing.appAwareProcessing.isEnabled":
			// WARNING if disabled; INFO if enabled
			if !toBool(drifts[i].VBR) {
				drifts[i].Severity = SeverityWarning
			} else {
				drifts[i].Severity = SeverityInfo
			}

		case "schedule.daily.isEnabled":
			// WARNING if disabled; INFO if enabled
			if !toBool(drifts[i].VBR) {
				drifts[i].Severity = SeverityWarning
			} else {
				drifts[i].Severity = SeverityInfo
			}

		case "schedule.runAutomatically":
			// WARNING if disabled; INFO if enabled
			if !toBool(drifts[i].VBR) {
				drifts[i].Severity = SeverityWarning
			} else {
				drifts[i].Severity = SeverityInfo
			}
		}
	}
	return drifts
}

// checkRepoHardeningDrift cross-references repository changes against state
// to detect when a job is moved off a hardened (LinuxHardened) repository.
func checkRepoHardeningDrift(drifts []Drift, stateSpec map[string]interface{}) []Drift {
	// Find the drift for storage.backupRepositoryId
	var repoDrift *Drift
	var repoDriftIdx int
	for i := range drifts {
		if drifts[i].Path == "storage.backupRepositoryId" {
			repoDrift = &drifts[i]
			repoDriftIdx = i
			break
		}
	}

	if repoDrift == nil || repoDrift.Action != "modified" {
		return drifts
	}

	// Extract old repo ID from state spec
	oldRepoID := extractNestedString(stateSpec, "storage", "backupRepositoryId")
	if oldRepoID == "" {
		return drifts
	}

	// Extract new repo ID from the VBR value
	newRepoID := toString(repoDrift.VBR)
	if newRepoID == "" {
		return drifts
	}

	// Load all VBRRepository resources from state, build ID-keyed lookup
	stateMgr := state.NewManager()
	repos, err := stateMgr.ListResources("VBRRepository")
	if err != nil || len(repos) == 0 {
		return drifts
	}

	repoByID := make(map[string]*state.Resource)
	for _, r := range repos {
		repoByID[r.ID] = r
	}

	oldRepo := repoByID[oldRepoID]
	newRepo := repoByID[newRepoID]

	if oldRepo == nil {
		return drifts
	}

	oldIsHardened := isHardenedRepo(oldRepo.Spec)
	newIsHardened := false
	if newRepo != nil {
		newIsHardened = isHardenedRepo(newRepo.Spec)
	}

	// If moved from hardened to non-hardened, ensure CRITICAL and add synthetic drift
	if oldIsHardened && !newIsHardened {
		drifts[repoDriftIdx].Severity = SeverityCritical

		newRepoName := "unknown"
		if newRepo != nil {
			newRepoName = newRepo.Name
		}

		syntheticDrift := Drift{
			Path:     "storage.backupRepositoryId",
			Action:   "modified",
			State:    fmt.Sprintf("Moved from hardened repository %q to non-hardened %q", oldRepo.Name, newRepoName),
			VBR:      nil,
			Severity: SeverityCritical,
		}
		drifts = append(drifts, syntheticDrift)
	}

	return drifts
}

// printSecuritySummary prints a header line when security-relevant drifts (WARNING+) exist.
func printSecuritySummary(drifts []Drift) {
	criticalCount := 0
	warningCount := 0

	for _, d := range drifts {
		switch d.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityWarning:
			warningCount++
		}
	}

	total := criticalCount + warningCount
	if total == 0 {
		return
	}

	if criticalCount > 0 {
		fmt.Printf("CRITICAL: %d security-relevant changes detected\n\n", total)
	} else {
		fmt.Printf("WARNING: %d security-relevant changes detected\n\n", total)
	}
}

// --- helper functions ---

// toBool converts an interface{} to bool. Handles bool and string representations.
func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		b, _ := strconv.ParseBool(val)
		return b
	default:
		return false
	}
}

// toFloat64 converts an interface{} to float64. Handles float64 and json.Number.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// toString converts an interface{} to string.
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// extractNestedString navigates a nested map and returns the string value at the given key path.
func extractNestedString(m map[string]interface{}, keys ...string) string {
	current := m
	for i, key := range keys {
		val, ok := current[key]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			return toString(val)
		}
		next, ok := val.(map[string]interface{})
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

// isHardenedRepo checks if a repository spec represents a Linux hardened repository.
func isHardenedRepo(spec map[string]interface{}) bool {
	repoType, ok := spec["type"]
	if !ok {
		return false
	}
	return toString(repoType) == "LinuxHardened"
}
