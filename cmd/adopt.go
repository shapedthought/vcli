package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os/user"
	"time"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
)

// adoptResource is the shared logic for all adopt subcommands.
// It loads a YAML spec, fetches the matching resource from VBR, compares them,
// saves to state with origin "applied", and reports any mismatches.
func adoptResource(
	specFile string,
	kind string,
	fetchCurrent func(name string, profile models.Profile) (json.RawMessage, string, error),
	ignoreFields map[string]bool,
	severityMap SeverityMap,
) {
	// 1. Load YAML spec
	spec, err := resources.LoadResourceSpec(specFile)
	if err != nil {
		log.Fatalf("Failed to load spec file: %v", err)
	}

	// 2. Validate kind
	if spec.Kind != kind {
		log.Fatalf("Expected kind %s, got %s", kind, spec.Kind)
	}

	// 3. Get VBR profile
	profile := utils.GetProfile()
	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// 4. Fetch current from VBR
	rawJSON, resourceID, err := fetchCurrent(spec.Metadata.Name, profile)
	if err != nil {
		log.Fatalf("Failed to fetch resource from VBR: %v", err)
	}

	// 5. Unmarshal both spec and VBR data to maps for comparison
	specMap := spec.Spec

	var vbrMap map[string]interface{}
	if err := json.Unmarshal(rawJSON, &vbrMap); err != nil {
		log.Fatalf("Failed to unmarshal VBR data: %v", err)
	}

	// 6. Compare using existing drift detection
	drifts := detectDrift(specMap, vbrMap, ignoreFields)
	drifts = classifyDrifts(drifts, severityMap)

	// 7. Save to state with origin "applied"
	if err := saveAdoptedResourceToState(kind, spec.Metadata.Name, resourceID, spec); err != nil {
		log.Fatalf("Failed to save adopted resource to state: %v", err)
	}

	stateMgr := state.NewManager()

	// 8. Print results
	if len(drifts) == 0 {
		fmt.Printf("Adopted %s: %s\n", kindDisplayName(kind), spec.Metadata.Name)
		fmt.Printf("  Origin: applied\n")
		fmt.Printf("  State updated: %s\n", stateMgr.GetStatePath())
	} else {
		fmt.Println("Warning: spec does not match current VBR configuration.")
		for _, drift := range drifts {
			printAdoptDrift(drift)
		}
		fmt.Printf("\nAdopted with %d differences. The next diff will show these as drift.\n", len(drifts))
		fmt.Printf("  Origin: applied\n")
		fmt.Printf("  State updated: %s\n", stateMgr.GetStatePath())
	}
}

// saveAdoptedResourceToState saves the spec's data to state with origin "applied".
// This stores the spec's desired state (not VBR's current data) so that subsequent
// diff commands will correctly detect VBR divergence from the spec.
func saveAdoptedResourceToState(kind, name, id string, spec resources.ResourceSpec) error {
	stateMgr := state.NewManager()

	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	resource := &state.Resource{
		Type:          kind,
		ID:            id,
		Name:          name,
		LastApplied:   time.Now(),
		LastAppliedBy: currentUser,
		Origin:        "applied",
		Spec:          spec.Spec,
	}

	if err := stateMgr.UpdateResource(resource); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

// printAdoptDrift prints a single drift entry in the adopt output format
func printAdoptDrift(drift Drift) {
	sev := string(drift.Severity)

	switch drift.Action {
	case "modified":
		specStr := formatValue(drift.State)
		vbrStr := formatValue(drift.VBR)
		fmt.Printf("  %s ~ %s: %s (spec) vs %s (VBR)\n", sev, drift.Path, specStr, vbrStr)
	case "removed":
		specStr := formatValue(drift.State)
		fmt.Printf("  %s - %s: %s (spec) not in VBR\n", sev, drift.Path, specStr)
	case "added":
		vbrStr := formatValue(drift.VBR)
		fmt.Printf("  %s + %s: %s in VBR but not in spec\n", sev, drift.Path, vbrStr)
	}
}

// kindDisplayName returns a human-readable name for a resource kind
func kindDisplayName(kind string) string {
	switch kind {
	case "VBRJob":
		return "job"
	case "VBRRepository":
		return "repository"
	case "VBRScaleOutRepository":
		return "scale-out repository"
	case "VBREncryptionPassword":
		return "encryption password"
	case "VBRKmsServer":
		return "KMS server"
	default:
		return kind
	}
}

// --- Job Adopt command ---

var jobAdoptCmd = &cobra.Command{
	Use:   "adopt [spec-file]",
	Short: "Adopt a job spec into state without modifying VBR",
	Long: `Adopt loads a YAML spec, fetches the matching job from VBR, compares them,
saves to state with origin "applied", and reports any mismatches.

This is a read-only operation — VBR is not modified.

Examples:
  # Adopt a job spec
  vcli job adopt backup-job.yaml
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		adoptResource(args[0], "VBRJob", fetchCurrentJob, jobIgnoreFields, jobSeverityMap)
	},
}

// fetchCurrentJob fetches the current job from VBR by name and returns raw JSON, job ID, and error
func fetchCurrentJob(name string, profile models.Profile) (json.RawMessage, string, error) {
	job, found := findJobByName(name, profile)
	if !found {
		return nil, "", fmt.Errorf("job '%s' not found in VBR", name)
	}

	// Fetch full job details as raw JSON
	endpoint := fmt.Sprintf("jobs/%s", job.ID)
	rawData := vhttp.GetData[json.RawMessage](endpoint, profile)

	return rawData, job.ID, nil
}

func init() {
	jobsCmd.AddCommand(jobAdoptCmd)
}
