package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/shapedthought/owlctl/config"
	"github.com/shapedthought/owlctl/models"
	"github.com/shapedthought/owlctl/resources"
	"github.com/shapedthought/owlctl/utils"
)

// GroupApplyResult tracks the outcome of applying a single spec within a group
type GroupApplyResult struct {
	SpecPath     string
	ResourceName string
	Action       string // "created", "updated", "would-create", "would-update"
	Error        error
}

// GroupDiffConfig defines resource-specific parameters for group diff operations
type GroupDiffConfig struct {
	// Kind is the expected resource kind string used in YAML specs (e.g., "VBRRepository")
	Kind string
	// DisplayName is the human-readable singular name for output (e.g., "repository")
	DisplayName string
	// PluralName is the human-readable plural name (e.g., "repositories"). If empty, DisplayName+"s" is used.
	PluralName string
	// FetchCurrent retrieves the current resource from VBR by name.
	// Returns (rawJSON, resourceID, error). If not found, returns (nil, "", nil).
	FetchCurrent func(name string, profile models.Profile) (json.RawMessage, string, error)
	// IgnoreFields are fields to exclude from drift detection
	IgnoreFields map[string]bool
	// SeverityMap classifies drift fields by severity
	SeverityMap SeverityMap
	// RemediateCmd is the remediation command template shown in summary (e.g., "owlctl repo apply --group %s")
	RemediateCmd string
}

// pluralDisplayName returns the plural form of the display name
func (dcfg GroupDiffConfig) pluralDisplayName() string {
	if dcfg.PluralName != "" {
		return dcfg.PluralName
	}
	return dcfg.DisplayName + "s"
}

// applyGroupResource applies all specs in a named group using the generic resource apply path.
// It loads the group config, resolves profile/overlay, merges each spec, and applies via applyResourceSpec.
func applyGroupResource(group string, applyCfg ResourceApplyConfig, dryRun bool) {
	settings := utils.ReadSettings()
	profile := utils.GetCurrentProfile()

	if settings.SelectedProfile != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load owlctl.yaml: %v", err)
	}
	cfg.WarnDeprecatedFields()

	groupCfg, err := cfg.GetGroup(group)
	if err != nil {
		log.Fatalf("Group error: %v", err)
	}

	if len(groupCfg.Specs) == 0 {
		log.Fatalf("Group %q has no specs defined", group)
	}

	// Resolve paths relative to owlctl.yaml
	profilePath := ""
	if groupCfg.Profile != "" {
		profilePath = cfg.ResolvePath(groupCfg.Profile)
	}
	overlayPath := ""
	if groupCfg.Overlay != "" {
		overlayPath = cfg.ResolvePath(groupCfg.Overlay)
	}

	fmt.Printf("Applying group: %s (%d specs)\n", group, len(groupCfg.Specs))
	if profilePath != "" {
		fmt.Printf("  Profile: %s\n", groupCfg.Profile)
	}
	if overlayPath != "" {
		fmt.Printf("  Overlay: %s\n", groupCfg.Overlay)
	}
	fmt.Println()

	// Pre-load profile and overlay once to avoid repeated disk I/O
	var profileSpec, overlaySpec *resources.ResourceSpec
	opts := resources.DefaultMergeOptions()

	if profilePath != "" {
		p, err := resources.LoadResourceSpec(profilePath)
		if err != nil {
			log.Fatalf("Failed to load profile %s: %v", profilePath, err)
		}
		profileSpec = &p
	}
	if overlayPath != "" {
		o, err := resources.LoadResourceSpec(overlayPath)
		if err != nil {
			log.Fatalf("Failed to load overlay %s: %v", overlayPath, err)
		}
		overlaySpec = &o
	}

	var results []GroupApplyResult

	for _, specRelPath := range groupCfg.Specs {
		specPath := cfg.ResolvePath(specRelPath)
		result := GroupApplyResult{SpecPath: specRelPath}

		// Load spec
		spec, err := resources.LoadResourceSpec(specPath)
		if err != nil {
			result.Error = fmt.Errorf("failed to load spec: %w", err)
			results = append(results, result)
			continue
		}

		// Merge with cached profile/overlay
		mergedSpec, err := resources.ApplyGroupMergeFromSpecs(spec, profileSpec, overlaySpec, opts)
		if err != nil {
			result.Error = fmt.Errorf("merge failed: %w", err)
			results = append(results, result)
			continue
		}

		result.ResourceName = mergedSpec.Metadata.Name

		// Validate kind matches expected resource type
		if mergedSpec.Kind != applyCfg.Kind {
			result.Error = fmt.Errorf("unsupported kind: %s (expected %s)", mergedSpec.Kind, applyCfg.Kind)
			results = append(results, result)
			continue
		}

		// Apply via generic resource apply
		applyResult := applyResourceSpec(mergedSpec, applyCfg, profile, dryRun)
		if applyResult.Error != nil {
			result.Error = applyResult.Error
		} else {
			result.Action = applyResult.Action
		}

		results = append(results, result)
	}

	printGroupApplySummary(group, results)

	// Determine exit code
	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}

	if successCount == 0 {
		os.Exit(ExitError)
	} else if successCount < len(results) {
		os.Exit(ExitPartialApply)
	}
	// All succeeded — exit 0 (default)
}

// diffGroupResource compares merged group specs (profile+spec+overlay) against live VBR state.
// Unlike state-based diff, group diff does NOT require state.json — the group definition
// IS the source of truth.
func diffGroupResource(group string, dcfg GroupDiffConfig) {
	loadSeverityOverrides()
	settings := utils.ReadSettings()
	profile := utils.GetCurrentProfile()

	if settings.SelectedProfile != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load owlctl.yaml: %v", err)
	}
	cfg.WarnDeprecatedFields()

	groupCfg, err := cfg.GetGroup(group)
	if err != nil {
		log.Fatalf("Group error: %v", err)
	}

	if len(groupCfg.Specs) == 0 {
		log.Fatalf("Group %q has no specs defined", group)
	}

	// Resolve paths
	profilePath := ""
	if groupCfg.Profile != "" {
		profilePath = cfg.ResolvePath(groupCfg.Profile)
	}
	overlayPath := ""
	if groupCfg.Overlay != "" {
		overlayPath = cfg.ResolvePath(groupCfg.Overlay)
	}

	fmt.Printf("Checking drift for group: %s (%d specs)\n", group, len(groupCfg.Specs))
	if profilePath != "" {
		fmt.Printf("  Profile: %s\n", groupCfg.Profile)
	}
	if overlayPath != "" {
		fmt.Printf("  Overlay: %s\n", groupCfg.Overlay)
	}
	fmt.Println()

	// Pre-load profile and overlay once to avoid repeated disk I/O
	var profileSpec, overlaySpec *resources.ResourceSpec
	opts := resources.DefaultMergeOptions()

	if profilePath != "" {
		p, err := resources.LoadResourceSpec(profilePath)
		if err != nil {
			log.Fatalf("Failed to load profile %s: %v", profilePath, err)
		}
		profileSpec = &p
	}
	if overlayPath != "" {
		o, err := resources.LoadResourceSpec(overlayPath)
		if err != nil {
			log.Fatalf("Failed to load overlay %s: %v", overlayPath, err)
		}
		overlaySpec = &o
	}

	minSev := parseSeverityFlag()
	cleanCount := 0
	driftedCount := 0
	notFoundCount := 0
	errorCount := 0
	var allDrifts []Drift

	for _, specRelPath := range groupCfg.Specs {
		specPath := cfg.ResolvePath(specRelPath)

		// Load spec
		spec, err := resources.LoadResourceSpec(specPath)
		if err != nil {
			fmt.Printf("  %s: Failed to load spec: %v\n", specRelPath, err)
			errorCount++
			continue
		}

		// Compute desired state from group merge using cached profile/overlay
		desiredSpec, err := resources.ApplyGroupMergeFromSpecs(spec, profileSpec, overlaySpec, opts)
		if err != nil {
			fmt.Printf("  %s: Failed to merge: %v\n", specRelPath, err)
			errorCount++
			continue
		}

		resourceName := desiredSpec.Metadata.Name

		// Validate kind matches expected resource type
		if desiredSpec.Kind != dcfg.Kind {
			fmt.Printf("  %s: Unsupported kind: %s (expected %s)\n", specRelPath, desiredSpec.Kind, dcfg.Kind)
			errorCount++
			continue
		}

		// Fetch current from VBR
		currentRaw, _, err := dcfg.FetchCurrent(resourceName, profile)
		if err != nil {
			fmt.Printf("  %s: Failed to fetch current: %v\n", resourceName, err)
			errorCount++
			continue
		}

		if currentRaw == nil {
			fmt.Printf("  %s: Not found in VBR (would be created by apply)\n", resourceName)
			notFoundCount++
			continue
		}

		// Convert current to map for comparison
		var currentMap map[string]interface{}
		if err := json.Unmarshal(currentRaw, &currentMap); err != nil {
			fmt.Printf("  %s: Failed to unmarshal current data: %v\n", resourceName, err)
			errorCount++
			continue
		}

		// Compare merged desired spec against live VBR
		drifts := detectDrift(desiredSpec.Spec, currentMap, dcfg.IgnoreFields)
		drifts = classifyDrifts(drifts, dcfg.SeverityMap)
		drifts = filterDriftsBySeverity(drifts, minSev)

		if len(drifts) > 0 {
			maxSev := getMaxSeverity(drifts)
			fmt.Printf("  %s %s: %d drifts detected\n", maxSev, resourceName, len(drifts))
			allDrifts = append(allDrifts, drifts...)
			driftedCount++
		} else {
			fmt.Printf("  %s: No drift\n", resourceName)
			cleanCount++
		}
	}

	// Summary
	if driftedCount > 0 {
		fmt.Println()
		printSecuritySummary(allDrifts)
	}

	plural := dcfg.pluralDisplayName()
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d %s clean\n", cleanCount, plural)
	if driftedCount > 0 {
		fmt.Printf("  - %d %s drifted — remediate with: %s\n", driftedCount, plural, fmt.Sprintf(dcfg.RemediateCmd, group))
	}
	if notFoundCount > 0 {
		fmt.Printf("  - %d %s not found in VBR (would be created by apply)\n", notFoundCount, plural)
	}
	if errorCount > 0 {
		fmt.Printf("  - %d specs failed to evaluate (see errors above)\n", errorCount)
	}

	if errorCount > 0 {
		os.Exit(ExitError)
	}
	if driftedCount > 0 {
		os.Exit(exitCodeForDrifts(allDrifts))
	}
	os.Exit(0)
}

// printGroupApplySummary prints a summary table after group apply
func printGroupApplySummary(group string, results []GroupApplyResult) {
	fmt.Printf("\n=== Group Apply Summary: %s ===\n", group)
	fmt.Printf("%-40s %-20s %-10s\n", "SPEC", "RESOURCE", "STATUS")
	fmt.Printf("%-40s %-20s %-10s\n", "----", "--------", "------")

	for _, r := range results {
		status := r.Action
		if r.Error != nil {
			status = fmt.Sprintf("FAILED: %v", r.Error)
		}
		name := r.ResourceName
		if name == "" {
			name = "(unknown)"
		}
		fmt.Printf("%-40s %-20s %s\n", r.SpecPath, name, status)
	}

	// Count results
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		} else {
			failCount++
		}
	}

	fmt.Printf("\nTotal: %d specs, %d succeeded, %d failed\n", len(results), successCount, failCount)
}
