package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/shapedthought/vcli/config"
	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
)

var (
	overlayFile string
	environment string
	dryRun      bool
)

var applyCmd = &cobra.Command{
	Use:   "apply [config-file]",
	Short: "Apply a declarative job configuration",
	Long: `Apply a declarative job configuration with optional overlay.

The apply command creates or updates VBR jobs based on YAML configuration files.
It supports overlay files for environment-specific customization.

Examples:
  # Apply a job configuration
  vcli job apply backup-job.yaml

  # Apply with overlay file
  vcli job apply base-job.yaml -o prod-overlay.yaml

  # Apply with environment (uses overlay from vcli.yaml)
  vcli job apply base-job.yaml --env production

  # Dry run to preview changes
  vcli job apply base-job.yaml -o prod-overlay.yaml --dry-run

Overlay Resolution:
  1. If -o/--overlay is specified, use that overlay file
  2. If --env is specified, use overlay from vcli.yaml for that environment
  3. If vcli.yaml exists and has currentEnvironment set, use that overlay
  4. Otherwise, apply base configuration without overlay
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		applyJob(args[0])
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
		// Try to use overlay from vcli.yaml
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

// needsConfigOverlay checks if we should try to use vcli.yaml overlay
func needsConfigOverlay() bool {
	// Try to load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}

	// Check if there's a current environment with an overlay
	if cfg.CurrentEnvironment != "" {
		_, err := cfg.GetEnvironmentOverlay("")
		return err == nil
	}

	return false
}

// getConfiguredOverlay gets the overlay path from vcli.yaml
func getConfiguredOverlay() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load vcli.yaml: %w", err)
	}

	// Use explicit environment if specified
	env := environment
	if env == "" {
		env = cfg.CurrentEnvironment
	}

	if env == "" {
		return "", fmt.Errorf("no environment specified and no currentEnvironment in vcli.yaml")
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

		// Debug: Print what we're sending
		debugBytes, _ := json.MarshalIndent(jobForPut, "", "  ")
		_ = os.WriteFile("/tmp/vcli-debug-request.json", debugBytes, 0644)
		fmt.Printf("\nDEBUG: Full request saved to /tmp/vcli-debug-request.json\n")
		fmt.Printf("Job ID in body: %s\n", jobForPut.ID)

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
	applyCmd.Flags().StringVar(&environment, "env", "", "Environment to use (looks up overlay from vcli.yaml)")
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")

	jobsCmd.AddCommand(applyCmd)
}
