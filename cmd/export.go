package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	exportOutput    string
	exportDirectory string
	exportAll       bool
	exportSimplified bool
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
	// Fetch job from VBR
	endpoint := fmt.Sprintf("jobs/%s", jobID)
	vbrJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)

	// Convert to declarative YAML
	yamlContent, err := convertJobToYAML(&vbrJob)
	if err != nil {
		log.Fatalf("Failed to convert job to YAML: %v", err)
	}

	// Output to file or stdout
	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, yamlContent, 0644); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
		fmt.Printf("Exported job to %s\n", exportOutput)
	} else {
		fmt.Println(string(yamlContent))
	}
}

func exportAllJobs(profile models.Profile) {
	// Fetch all jobs
	type JobList struct {
		Data []models.VbrJobGet `json:"data"`
	}

	jobs := vhttp.GetData[JobList]("jobs", profile)

	if len(jobs.Data) == 0 {
		fmt.Println("No jobs found")
		return
	}

	// Determine output directory
	outputDir := exportDirectory
	if outputDir == "" {
		outputDir = "."
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	successCount := 0
	failedCount := 0

	fmt.Printf("Exporting %d jobs...\n", len(jobs.Data))

	for i, job := range jobs.Data {
		// Fetch full job details
		endpoint := fmt.Sprintf("jobs/%s", job.ID)
		fullJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)

		// Convert to YAML
		yamlContent, err := convertJobToYAML(&fullJob)
		if err != nil {
			fmt.Printf("Warning: Failed to convert job %s: %v\n", job.Name, err)
			failedCount++
			continue
		}

		// Sanitize job name for filename
		filename := sanitizeFilename(job.Name) + ".yaml"
		filepath := filepath.Join(outputDir, filename)

		// Write to file
		if err := os.WriteFile(filepath, yamlContent, 0644); err != nil {
			fmt.Printf("Warning: Failed to write %s: %v\n", filename, err)
			failedCount++
			continue
		}

		successCount++
		fmt.Printf("  [%d/%d] Exported %s\n", i+1, len(jobs.Data), filename)
	}

	fmt.Printf("\nExport complete: %d successful, %d failed\n", successCount, failedCount)
}

func convertJobToYAML(vbrJob *models.VbrJobGet) ([]byte, error) {
	if exportSimplified {
		return convertJobToYAMLSimplified(vbrJob)
	}
	return convertJobToYAMLFull(vbrJob)
}

func convertJobToYAMLFull(vbrJob *models.VbrJobGet) ([]byte, error) {
	// Create full resource spec with complete job object
	resourceSpec := resources.ResourceSpec{
		APIVersion: "vcli.veeam.com/v1",
		Kind:       "VBRJob",
		Metadata: resources.Metadata{
			Name: vbrJob.Name,
		},
		Spec: make(map[string]interface{}),
	}

	// Marshal the entire job to preserve all fields
	jobBytes, err := yaml.Marshal(vbrJob)
	if err != nil {
		return nil, err
	}

	var jobMap map[string]interface{}
	if err := yaml.Unmarshal(jobBytes, &jobMap); err != nil {
		return nil, err
	}
	resourceSpec.Spec = jobMap

	// Add header comment
	header := fmt.Sprintf("# VBR Job Configuration (Full Export)\n# Exported from VBR\n# Job ID: %s\n# API Version: 1.3-rev1\n#\n# This export contains the complete job configuration.\n# All ~300 fields from the VBR API are preserved.\n\n", vbrJob.ID)

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
	rootCmd.AddCommand(exportCmd)
}
