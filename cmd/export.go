package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	exportOutput     string
	exportDirectory  string
	exportAll        bool
	exportSimplified bool
	exportAsOverlay  bool
	exportBasePath   string
)

var exportCmd = &cobra.Command{
	Use:   "export [job-id]",
	Short: "Export VBR jobs to declarative YAML format",
	Long: `Export existing VBR backup jobs to declarative YAML configuration files.

Examples:
  # Export single job to stdout
  vcli export 57b3baab-6237-41bf-add7-db63d41d984c

  # Export single job to file
  vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o backup.yaml

  # Export as overlay (minimal patch with only changed fields)
  vcli export 57b3baab-6237-41bf-add7-db63d41d984c --as-overlay -o overlay.yaml

  # Export as overlay with specific base
  vcli export 57b3baab-6237-41bf-add7-db63d41d984c --as-overlay --base base/defaults.yaml -o overlay.yaml

  # Export all jobs to current directory
  vcli export --all

  # Export all jobs to specific directory
  vcli export --all -d ./configs/
`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name != "vbr" {
			log.Fatal("This command only works with VBR at the moment.")
		}

		if exportAll {
			exportAllJobs(profile)
		} else {
			if len(args) == 0 {
				log.Fatal("Job ID is required. Use --all to export all jobs.")
			}
			exportSingleJob(args[0], profile)
		}
	},
}

func exportSingleJob(jobID string, profile models.Profile) {
	endpoint := fmt.Sprintf("jobs/%s", jobID)

	var yamlContent []byte
	var err error

	if exportSimplified || exportAsOverlay {
		// Simplified and overlay modes use typed struct (VSphere-specific fields)
		vbrJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)
		if exportAsOverlay {
			yamlContent, err = convertJobToYAMLOverlay(&vbrJob, exportBasePath)
		} else {
			yamlContent, err = convertJobToYAMLSimplified(&vbrJob)
		}
	} else {
		// Full mode: use raw JSON to preserve all fields regardless of job type
		rawData := vhttp.GetData[json.RawMessage](endpoint, profile)

		var meta struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		}
		if jsonErr := json.Unmarshal(rawData, &meta); jsonErr != nil {
			log.Fatalf("Failed to parse job data: %v", jsonErr)
		}

		yamlContent, err = convertResourceToYAML(meta.Name, rawData, ResourceExportConfig{
			Kind:         "VBRJob",
			IgnoreFields: jobIgnoreFields,
			HeaderLines: []string{
				"# VBR Job Configuration (Full Export)",
				"# Exported from VBR",
				fmt.Sprintf("# Job ID: %s", meta.ID),
				"# API Version: 1.3-rev1",
				"#",
				"# This export contains the complete job configuration.",
				"# All fields from the VBR API are preserved.",
			},
		})
	}

	if err != nil {
		log.Fatalf("Failed to convert job to YAML: %v", err)
	}

	writeExportOutput(yamlContent, exportOutput, "job")
}

func exportAllJobs(profile models.Profile) {
	// Determine output directory
	outputDir := exportDirectory
	if outputDir == "" {
		outputDir = "."
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	successCount := 0
	failedCount := 0

	if exportSimplified || exportAsOverlay {
		// Simplified and overlay modes use typed struct (VSphere-specific fields)
		type JobList struct {
			Data []models.VbrJobGet `json:"data"`
		}
		jobs := vhttp.GetData[JobList]("jobs", profile)

		if len(jobs.Data) == 0 {
			fmt.Println("No jobs found")
			return
		}

		fmt.Printf("Exporting %d jobs...\n", len(jobs.Data))

		for i, job := range jobs.Data {
			endpoint := fmt.Sprintf("jobs/%s", job.ID)
			fullJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)

			var yamlContent []byte
			var err error
			if exportAsOverlay {
				yamlContent, err = convertJobToYAMLOverlay(&fullJob, exportBasePath)
			} else {
				yamlContent, err = convertJobToYAMLSimplified(&fullJob)
			}
			if err != nil {
				fmt.Printf("Warning: Failed to convert job %s: %v\n", job.Name, err)
				failedCount++
				continue
			}

			if writeErr := writeExportAllOutput(yamlContent, outputDir, job.Name, i, len(jobs.Data)); writeErr != nil {
				fmt.Printf("Warning: %v\n", writeErr)
				failedCount++
				continue
			}
			successCount++
		}
	} else {
		// Full mode: use raw JSON to preserve all fields regardless of job type
		type JobList struct {
			Data []json.RawMessage `json:"data"`
		}
		jobs := vhttp.GetData[JobList]("jobs", profile)

		if len(jobs.Data) == 0 {
			fmt.Println("No jobs found")
			return
		}

		fmt.Printf("Exporting %d jobs...\n", len(jobs.Data))

		for i, rawJob := range jobs.Data {
			var meta struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}
			if err := json.Unmarshal(rawJob, &meta); err != nil {
				fmt.Printf("Warning: Failed to parse job %d: %v\n", i, err)
				failedCount++
				continue
			}

			// Fetch full job details as raw JSON
			endpoint := fmt.Sprintf("jobs/%s", meta.ID)
			fullRaw := vhttp.GetData[json.RawMessage](endpoint, profile)

			yamlContent, err := convertResourceToYAML(meta.Name, fullRaw, ResourceExportConfig{
				Kind:         "VBRJob",
				IgnoreFields: jobIgnoreFields,
				HeaderLines: []string{
					"# VBR Job Configuration (Full Export)",
					"# Exported from VBR",
					fmt.Sprintf("# Job ID: %s", meta.ID),
					"# API Version: 1.3-rev1",
					"#",
					"# This export contains the complete job configuration.",
					"# All fields from the VBR API are preserved.",
				},
			})
			if err != nil {
				fmt.Printf("Warning: Failed to convert job %s: %v\n", meta.Name, err)
				failedCount++
				continue
			}

			if writeErr := writeExportAllOutput(yamlContent, outputDir, meta.Name, i, len(jobs.Data)); writeErr != nil {
				fmt.Printf("Warning: %v\n", writeErr)
				failedCount++
				continue
			}
			successCount++
		}
	}

	fmt.Printf("\nExport complete: %d successful, %d failed\n", successCount, failedCount)
}

func convertJobToYAMLSimplified(vbrJob *models.VbrJobGet) ([]byte, error) {
	// Create resolver for name resolution
	resolver := resources.NewResolver()

	// Build simplified spec
	spec := resources.VBRJobSpec{
		Type:        vbrJob.Type,
		Description: vbrJob.Description,
		IsDisabled:  vbrJob.IsDisabled,
	}

	// Convert VMs
	for _, include := range vbrJob.VirtualMachines.Includes {
		spec.Objects = append(spec.Objects, resources.JobObject{
			Type:     include.Type,
			Name:     include.Name,
			HostName: include.HostName,
		})
	}

	// Resolve repository name
	repoName, err := resolver.ResolveRepositoryName(vbrJob.Storage.BackupRepositoryID)
	if err != nil {
		// If we can't resolve, use the ID
		repoName = vbrJob.Storage.BackupRepositoryID
	}
	spec.Repository = repoName

	// Convert schedule
	if vbrJob.Schedule.RunAutomatically {
		spec.Schedule = &resources.JobSchedule{
			Enabled: true,
		}
		if vbrJob.Schedule.Daily.IsEnabled {
			spec.Schedule.Daily = vbrJob.Schedule.Daily.LocalTime
		}
		if vbrJob.Schedule.Retry.IsEnabled {
			spec.Schedule.Retry = &struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
				Times   int  `yaml:"times,omitempty" json:"times,omitempty"`
				Wait    int  `yaml:"wait,omitempty" json:"wait,omitempty"`
			}{
				Enabled: true,
				Times:   vbrJob.Schedule.Retry.RetryCount,
				Wait:    vbrJob.Schedule.Retry.AwaitMinutes,
			}
		}
	}

	// Convert storage settings
	spec.Storage = &resources.JobStorageSettings{
		Compression: vbrJob.Storage.AdvancedSettings.StorageData.CompressionLevel,
		Encryption:  vbrJob.Storage.AdvancedSettings.StorageData.Encryption.IsEnabled,
		Retention: &struct {
			Type     string `yaml:"type,omitempty" json:"type,omitempty"`
			Quantity int    `yaml:"quantity,omitempty" json:"quantity,omitempty"`
		}{
			Type:     vbrJob.Storage.RetentionPolicy.Type,
			Quantity: vbrJob.Storage.RetentionPolicy.Quantity,
		},
	}

	// Create full resource spec
	resourceSpec := resources.ResourceSpec{
		APIVersion: "vcli.veeam.com/v1",
		Kind:       "VBRJob",
		Metadata: resources.Metadata{
			Name: vbrJob.Name,
		},
		Spec: make(map[string]interface{}),
	}

	// Convert spec to map
	specBytes, err := yaml.Marshal(spec)
	if err != nil {
		return nil, err
	}

	var specMap map[string]interface{}
	if err := yaml.Unmarshal(specBytes, &specMap); err != nil {
		return nil, err
	}
	resourceSpec.Spec = specMap

	// Add header comment
	header := fmt.Sprintf("# VBR Job Configuration\n# Exported from VBR\n# Job ID: %s\n\n", vbrJob.ID)

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(resourceSpec)
	if err != nil {
		return nil, err
	}

	// Combine header and YAML
	result := []byte(header)
	result = append(result, yamlBytes...)

	return result, nil
}

func convertJobToYAMLOverlay(vbrJob *models.VbrJobGet, basePath string) ([]byte, error) {
	var baseSpec map[string]interface{}

	// Load base if provided
	if basePath != "" {
		baseData, err := os.ReadFile(basePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read base file: %w", err)
		}

		var baseResource resources.ResourceSpec
		if err := yaml.Unmarshal(baseData, &baseResource); err != nil {
			return nil, fmt.Errorf("failed to parse base file: %w", err)
		}
		baseSpec = baseResource.Spec
	}

	// Convert job to map
	jobBytes, err := yaml.Marshal(vbrJob)
	if err != nil {
		return nil, err
	}

	var jobMap map[string]interface{}
	if err := yaml.Unmarshal(jobBytes, &jobMap); err != nil {
		return nil, err
	}

	// Calculate overlay (differences from base)
	var overlaySpec map[string]interface{}
	if baseSpec != nil {
		// Diff against provided base
		overlaySpec = calculateDiff(baseSpec, jobMap)
	} else {
		// No base provided - create minimal overlay with commonly changed fields
		overlaySpec = extractCommonFields(jobMap)
	}

	// Create resource spec
	resourceSpec := resources.ResourceSpec{
		APIVersion: "vcli.veeam.com/v1",
		Kind:       "VBRJob",
		Metadata: resources.Metadata{
			Name: vbrJob.Name,
		},
		Spec: overlaySpec,
	}

	// Add header comment
	header := "# VBR Job Overlay\n"
	if basePath != "" {
		header += fmt.Sprintf("# Base: %s\n", basePath)
	}
	header += fmt.Sprintf("# Job ID: %s\n", vbrJob.ID)
	header += "#\n# This overlay contains only the fields that differ from the base.\n# Apply with: vcli job apply base.yaml -o this-file.yaml\n\n"

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(resourceSpec)
	if err != nil {
		return nil, err
	}

	// Combine header and YAML
	result := []byte(header)
	result = append(result, yamlBytes...)

	return result, nil
}

// calculateDiff compares job against base and returns only differences
func calculateDiff(base, job map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	for key, jobValue := range job {
		// Skip read-only fields
		if key == "id" {
			continue
		}

		baseValue, existsInBase := base[key]

		if !existsInBase {
			// Field only in job, include it
			diff[key] = jobValue
			continue
		}

		// Both exist - check if different
		if !deepEqual(baseValue, jobValue) {
			// Different - add to diff
			switch jv := jobValue.(type) {
			case map[string]interface{}:
				// Recursively diff nested objects
				if bv, ok := baseValue.(map[string]interface{}); ok {
					nestedDiff := calculateDiff(bv, jv)
					if len(nestedDiff) > 0 {
						diff[key] = nestedDiff
					}
				} else {
					diff[key] = jobValue
				}
			default:
				diff[key] = jobValue
			}
		}
	}

	return diff
}

// extractCommonFields extracts commonly changed fields when no base provided
// This is intentionally minimal - it only extracts the most frequently modified fields
// (description, retention policy, and daily schedule). If you need more fields in the
// overlay, provide a base template with --base to get a full diff calculation.
func extractCommonFields(job map[string]interface{}) map[string]interface{} {
	overlay := make(map[string]interface{})

	// Include description if present
	if desc, ok := job["description"].(string); ok && desc != "" {
		overlay["description"] = desc
	}

	// Include storage retention if present
	if storage, ok := job["storage"].(map[string]interface{}); ok {
		if retentionPolicy, ok := storage["retentionPolicy"].(map[string]interface{}); ok {
			overlay["storage"] = map[string]interface{}{
				"retentionPolicy": retentionPolicy,
			}
		}
	}

	// Include schedule if present
	if schedule, ok := job["schedule"].(map[string]interface{}); ok {
		if daily, ok := schedule["daily"].(map[string]interface{}); ok {
			overlay["schedule"] = map[string]interface{}{
				"daily": daily,
			}
		}
	}

	return overlay
}

// deepEqual compares two values for deep equality
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func sanitizeFilename(name string) string {
	// Replace invalid filename characters with hyphens
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := reg.ReplaceAllString(name, "-")

	// Replace spaces with hyphens
	sanitized = strings.ReplaceAll(sanitized, " ", "-")

	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	sanitized = reg.ReplaceAllString(sanitized, "-")

	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")

	return sanitized
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().StringVarP(&exportDirectory, "directory", "d", "", "Output directory for bulk export")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all jobs")
	exportCmd.Flags().BoolVar(&exportSimplified, "simplified", false, "Export simplified format (legacy)")
	exportCmd.Flags().BoolVar(&exportAsOverlay, "as-overlay", false, "Export as overlay (minimal patch)")
	exportCmd.Flags().StringVar(&exportBasePath, "base", "", "Base template to diff against (for overlay export)")
	rootCmd.AddCommand(exportCmd)
}
