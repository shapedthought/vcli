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
	"github.com/shapedthought/owlctl/vhttp"
	"github.com/spf13/cobra"
)

var (
	overlayFile string
	environment string
	dryRun      bool
	groupName   string
)

// GroupApplyResult tracks the outcome of applying a single spec within a group
type GroupApplyResult struct {
	SpecPath string
	JobName  string
	Action   string // "created", "updated", "would-create", "would-update"
	Error    error
}

var applyCmd = &cobra.Command{
	Use:   "apply [config-file]",
	Short: "Apply a declarative job configuration",
	Long: `Apply a declarative job configuration with optional overlay.

The apply command creates or updates VBR jobs based on YAML configuration files.
It supports overlay files for environment-specific customization.

Examples:
  # Apply a job configuration
  owlctl job apply backup-job.yaml

  # Apply with overlay file
  owlctl job apply base-job.yaml -o prod-overlay.yaml

  # Apply with environment (uses overlay from owlctl.yaml)
  owlctl job apply base-job.yaml --env production

  # Dry run to preview changes
  owlctl job apply base-job.yaml -o prod-overlay.yaml --dry-run

  # Apply all specs in a group (from owlctl.yaml)
  owlctl job apply --group sql-tier

  # Dry run a group
  owlctl job apply --group sql-tier --dry-run

Overlay Resolution:
  1. If -o/--overlay is specified, use that overlay file
  2. If --env is specified, use overlay from owlctl.yaml for that environment
  3. If owlctl.yaml exists and has currentEnvironment set, use that overlay
  4. Otherwise, apply base configuration without overlay
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if groupName != "" {
			// Validate mutual exclusivity
			if len(args) > 0 {
				log.Fatal("Cannot use --group with a positional config file argument")
			}
			if overlayFile != "" {
				log.Fatal("Cannot use --group with --overlay (group defines its own overlay)")
			}
			if environment != "" {
				log.Fatal("Cannot use --group with --env (group defines its own overlay)")
			}
			applyGroup(groupName)
		} else if len(args) > 0 {
			applyJob(args[0])
		} else {
			log.Fatal("Provide a config file or use --group")
		}
	},
}

func applyJob(configFile string) {
	settings := utils.ReadSettings()
	profile := utils.GetCurrentProfile()

	if settings.SelectedProfile != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// Load base configuration
	baseSpec, err := resources.LoadResourceSpec(configFile)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// Validate resource type
	if baseSpec.Kind != "VBRJob" {
		log.Fatalf("Unsupported resource kind: %s (only VBRJob is supported)", baseSpec.Kind)
	}

	// Determine which overlay to use
	var finalSpec resources.ResourceSpec
	if overlayFile != "" {
		// Explicit overlay file provided
		fmt.Printf("Applying overlay: %s\n", overlayFile)
		mergedSpec, err := resources.MergeYAMLFiles(configFile, overlayFile, resources.DefaultMergeOptions())
		if err != nil {
			log.Fatalf("Failed to merge with overlay: %v", err)
		}
		finalSpec = mergedSpec
	} else if environment != "" || needsConfigOverlay() {
		// Try to use overlay from owlctl.yaml
		overlayPath, err := getConfiguredOverlay()
		if err != nil {
			// If no overlay configured, just use base spec
			fmt.Printf("No overlay configured, applying base configuration\n")
			finalSpec = baseSpec
		} else {
			fmt.Printf("Applying environment overlay: %s\n", overlayPath)
			mergedSpec, err := resources.MergeYAMLFiles(configFile, overlayPath, resources.DefaultMergeOptions())
			if err != nil {
				log.Fatalf("Failed to merge with configured overlay: %v", err)
			}
			finalSpec = mergedSpec
		}
	} else {
		// No overlay specified
		fmt.Println("Applying base configuration (no overlay)")
		finalSpec = baseSpec
	}

	// Display merged configuration in dry-run mode
	if dryRun {
		fmt.Println("\n╔════════════════════════════════════════════════════════════════════╗")
		fmt.Println("║                         Dry Run Mode                               ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("Resource: %s (%s)\n", finalSpec.Metadata.Name, finalSpec.Kind)
		fmt.Println()

		// Fetch current job from VBR to show diff
		currentJob, exists := findJobByName(finalSpec.Metadata.Name, profile)

		if !exists {
			// Job doesn't exist - will be created
			fmt.Println("⚠ Job not found in VBR - will be created")
			fmt.Println()
			showNewJobSummary(finalSpec)
		} else {
			// Job exists - show diff
			fmt.Printf("✓ Current job found in VBR (ID: %s)\n", currentJob.ID)
			fmt.Println()
			showJobDiff(finalSpec, currentJob)
		}

		fmt.Println("\n=== End Dry Run ===")
		fmt.Println("Note: This is a preview. No changes have been made.")
		fmt.Println("Remove --dry-run flag to apply these changes.")
		return
	}

	// Apply the job configuration
	if err := applyVBRJob(finalSpec, profile); err != nil {
		log.Fatalf("Failed to apply job: %v", err)
	}

	fmt.Printf("\n✓ Successfully applied job: %s\n", finalSpec.Metadata.Name)
}

// applyGroup applies all specs in a named group from owlctl.yaml
func applyGroup(group string) {
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

	var results []GroupApplyResult

	for _, specRelPath := range groupCfg.Specs {
		specPath := cfg.ResolvePath(specRelPath)
		result := GroupApplyResult{SpecPath: specRelPath}

		// Merge with group profile/overlay
		mergedSpec, err := resources.ApplyGroupMerge(specPath, profilePath, overlayPath, resources.DefaultMergeOptions())
		if err != nil {
			result.Error = fmt.Errorf("merge failed: %w", err)
			results = append(results, result)
			continue
		}

		result.JobName = mergedSpec.Metadata.Name

		// Validate kind
		if mergedSpec.Kind != resources.KindVBRJob {
			result.Error = fmt.Errorf("unsupported kind: %s (expected VBRJob)", mergedSpec.Kind)
			results = append(results, result)
			continue
		}

		if dryRun {
			// Show dry-run preview for this spec
			fmt.Printf("--- %s ---\n", specRelPath)
			fmt.Printf("Resource: %s (%s)\n", mergedSpec.Metadata.Name, mergedSpec.Kind)

			currentJob, exists := findJobByName(mergedSpec.Metadata.Name, profile)
			if !exists {
				fmt.Println("  Would be created (not found in VBR)")
				result.Action = "would-create"
			} else {
				fmt.Printf("  Found in VBR (ID: %s) — would be updated\n", currentJob.ID)
				result.Action = "would-update"
			}
			fmt.Println()
		} else {
			// Apply the job
			if err := applyVBRJob(mergedSpec, profile); err != nil {
				result.Error = err
			} else {
				// Determine action based on whether job existed
				_, exists := findJobByName(mergedSpec.Metadata.Name, profile)
				if exists {
					result.Action = "updated"
				} else {
					result.Action = "created"
				}
			}
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

// printGroupApplySummary prints a summary table after group apply
func printGroupApplySummary(group string, results []GroupApplyResult) {
	fmt.Printf("\n=== Group Apply Summary: %s ===\n", group)
	fmt.Printf("%-40s %-20s %-10s\n", "SPEC", "JOB NAME", "STATUS")
	fmt.Printf("%-40s %-20s %-10s\n", "----", "--------", "------")

	for _, r := range results {
		status := r.Action
		if r.Error != nil {
			status = fmt.Sprintf("FAILED: %v", r.Error)
		}
		jobName := r.JobName
		if jobName == "" {
			jobName = "(unknown)"
		}
		fmt.Printf("%-40s %-20s %s\n", r.SpecPath, jobName, status)
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

// needsConfigOverlay checks if we should try to use owlctl.yaml overlay
func needsConfigOverlay() bool {
	// Try to load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}

	// Check if there's a current environment with an overlay
	if cfg.CurrentEnvironment != "" {
		cfg.WarnDeprecatedFields()
		_, err := cfg.GetEnvironmentOverlay("")
		return err == nil
	}

	return false
}

// getConfiguredOverlay gets the overlay path from owlctl.yaml
func getConfiguredOverlay() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load owlctl.yaml: %w", err)
	}

	cfg.WarnDeprecatedFields()

	// Use explicit environment if specified
	env := environment
	if env == "" {
		env = cfg.CurrentEnvironment
	}

	if env == "" {
		return "", fmt.Errorf("no environment specified and no currentEnvironment in owlctl.yaml")
	}

	overlayPath, err := cfg.GetEnvironmentOverlay(env)
	if err != nil {
		return "", err
	}

	return overlayPath, nil
}

// applyVBRJob creates or updates a VBR job based on the specification
func applyVBRJob(spec resources.ResourceSpec, profile models.Profile) error {
	// Convert spec to VbrJobPost model
	jobPost, err := specToVBRJob(spec)
	if err != nil {
		return fmt.Errorf("failed to convert spec to VBR job: %w", err)
	}

	// Check if job already exists
	existingJob, exists := findJobByName(jobPost.Name, profile)

	var jobID string
	if exists {
		// Update existing job
		fmt.Printf("Job '%s' already exists (ID: %s), updating...\n", jobPost.Name, existingJob.ID)

		// For PUT, we need to merge our changes with the existing job
		// Convert existing job to VbrJobPost
		mergedJob := mergeJobUpdates(existingJob, jobPost)

		// IMPORTANT: VBR requires the ID in the request body for PUT (as well as in URL)
		// Convert to VbrJobGet which has the ID field for the PUT request
		jobForPut := models.VbrJobGet{
			ID:              existingJob.ID, // Include ID in body for PUT
			Type:            mergedJob.Type,
			Name:            mergedJob.Name,
			Description:     mergedJob.Description,
			IsDisabled:      mergedJob.IsDisabled,
			IsHighPriority:  mergedJob.IsHighPriority,
			VirtualMachines: mergedJob.VirtualMachines,
			Storage:         mergedJob.Storage,
			GuestProcessing: mergedJob.GuestProcessing,
			Schedule:        mergedJob.Schedule,
		}

		// Debug: Print what we're sending (only when OWLCTL_DEBUG is set)
		if os.Getenv("OWLCTL_DEBUG") != "" {
			debugBytes, _ := json.MarshalIndent(jobForPut, "", "  ")
			_ = os.WriteFile("/tmp/owlctl-debug-request.json", debugBytes, 0600)
			fmt.Fprintf(os.Stderr, "\nDEBUG: Full request saved to /tmp/owlctl-debug-request.json\n")
			fmt.Fprintf(os.Stderr, "Job ID in body: %s\n", jobForPut.ID)
		}

		endpoint := fmt.Sprintf("jobs/%s", existingJob.ID)
		vhttp.PutData(endpoint, jobForPut, profile)
		fmt.Printf("Updated job: %s\n", jobPost.Name)
		jobID = existingJob.ID
	} else {
		// Create new job
		fmt.Printf("Creating new job: %s\n", jobPost.Name)
		result := vhttp.PostData[models.VbrJobGet]("jobs", jobPost, profile)
		fmt.Printf("Created job with ID: %s\n", result.ID)
		jobID = result.ID
	}

	// Update state after successful apply (using shared helper)
	// Note: Legacy job apply doesn't track field changes like generic applyResource
	if err := updateResourceState(spec, jobID, "VBRJob", nil); err != nil {
		// Log warning but don't fail the apply
		fmt.Printf("Warning: Failed to update state: %v\n", err)
	}

	return nil
}

// mergeJobUpdates merges the desired changes into the existing job
func mergeJobUpdates(existing models.VbrJobGet, desired models.VbrJobPost) models.VbrJobPost {
	// Start with the existing job (convert to VbrJobPost)
	merged := models.VbrJobPost{
		Type:            existing.Type,
		Name:            existing.Name,
		Description:     existing.Description,
		IsDisabled:      existing.IsDisabled,
		IsHighPriority:  existing.IsHighPriority,
		VirtualMachines: existing.VirtualMachines,
		Storage:         existing.Storage,
		GuestProcessing: existing.GuestProcessing,
		Schedule:        existing.Schedule,
	}

	// Apply desired changes on top
	if desired.Description != "" {
		merged.Description = desired.Description
	}
	if desired.Type != "" {
		merged.Type = desired.Type
	}
	// Merge storage settings if provided
	if desired.Storage.RetentionPolicy.Quantity > 0 {
		merged.Storage.RetentionPolicy.Quantity = desired.Storage.RetentionPolicy.Quantity
	}
	// Merge schedule if provided
	if desired.Schedule.Daily.LocalTime != "" {
		merged.Schedule.Daily.LocalTime = desired.Schedule.Daily.LocalTime
	}

	// Apply same data cleaning to merged job
	// For credentials, if both old and new formats are empty, use agent management credentials
	hasOldCreds := merged.GuestProcessing.GuestCredentials.CredsType != ""
	hasNewCreds := merged.GuestProcessing.GuestCredentials.Credentials != nil
	if !hasOldCreds && !hasNewCreds {
		merged.GuestProcessing.GuestCredentials.UseAgentManagementCredentials = true
	}
	// If credentials exist but useAgentManagementCredentials is not set, explicitly set it to false
	if (hasOldCreds || hasNewCreds) && !merged.GuestProcessing.GuestCredentials.UseAgentManagementCredentials {
		// Keep it as false (default value)
	}
	cleanVMExcludes(&merged.VirtualMachines)

	return merged
}

// specToVBRJob converts a ResourceSpec to VbrJobPost model
func specToVBRJob(spec resources.ResourceSpec) (models.VbrJobPost, error) {
	// Remove read-only fields that shouldn't be in POST/PUT requests
	cleanedSpec := make(map[string]interface{})
	for k, v := range spec.Spec {
		// Filter out read-only fields
		if k == "id" {
			continue // ID should not be in request body
		}
		cleanedSpec[k] = v
	}

	// Marshal cleaned spec to JSON then unmarshal to VbrJobPost
	specBytes, err := json.Marshal(cleanedSpec)
	if err != nil {
		return models.VbrJobPost{}, fmt.Errorf("failed to marshal spec: %w", err)
	}

	var jobPost models.VbrJobPost
	if err := json.Unmarshal(specBytes, &jobPost); err != nil {
		return models.VbrJobPost{}, fmt.Errorf("failed to unmarshal to VbrJobPost: %w", err)
	}

	// Always use name from metadata (authoritative source)
	// Metadata.name overrides spec.name
	jobPost.Name = spec.Metadata.Name

	// Apply defaults for fields required by API v1.3+
	// If no credentials specified, use agent management credentials
	if jobPost.GuestProcessing.GuestCredentials.CredsType == "" && jobPost.GuestProcessing.GuestCredentials.Credentials == nil {
		jobPost.GuestProcessing.GuestCredentials.UseAgentManagementCredentials = true
	}

	// Clean up VirtualMachines.Excludes to remove invalid entries
	// VBR API v1.3 requires Platform field in vmObject, but exports don't always include it
	// If excludes.disks is empty or has no actual disk exclusions, clear it
	cleanVMExcludes(&jobPost.VirtualMachines)

	// Debug: Check if we're accidentally including the ID
	debugBytes, _ := json.MarshalIndent(jobPost, "", "  ")
	if contains := string(debugBytes); len(contains) > 0 {
		// Check for ID field (should not be present in POST/PUT body)
		fmt.Printf("DEBUG: Checking for ID field in request body...\n")
	}

	return jobPost, nil
}

// cleanVMExcludes removes invalid or empty exclude entries
func cleanVMExcludes(vms *models.VirtualMachines) {
	// Clear excludes.disks if empty or contains only default entries
	if len(vms.Excludes.Disks) > 0 {
		hasValidExcludes := false
		for _, disk := range vms.Excludes.Disks {
			if len(disk.Disks) > 0 {
				hasValidExcludes = true
				break
			}
		}
		// If no actual disk exclusions, clear the array
		if !hasValidExcludes {
			vms.Excludes.Disks = nil
		}
	}
}

// findJobByName searches for an existing job by name
func findJobByName(name string, profile models.Profile) (models.VbrJobGet, bool) {
	// Get all jobs
	type JobsResponse struct {
		Data []models.VbrJobGet `json:"data"`
	}

	response := vhttp.GetData[JobsResponse]("jobs", profile)

	// Search for job with matching name
	for _, job := range response.Data {
		if job.Name == name {
			return job, true
		}
	}

	return models.VbrJobGet{}, false
}

func init() {
	applyCmd.Flags().StringVarP(&overlayFile, "overlay", "o", "", "Overlay file to merge with base configuration")
	applyCmd.Flags().StringVar(&environment, "env", "", "Environment to use (looks up overlay from owlctl.yaml)")
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")
	applyCmd.Flags().StringVar(&groupName, "group", "", "Apply all specs in named group (from owlctl.yaml)")

	jobsCmd.AddCommand(applyCmd)
}
