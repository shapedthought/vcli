package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
)

var diffJSON bool

var diffCmd = &cobra.Command{
	Use:   "diff [resource-name]",
	Short: "Detect configuration drift",
	Long: `Diff detects drift between state file (last applied configuration) and actual VBR configuration.

This helps identify:
- Manual changes made in VBR UI
- Changes made by other tools
- Out-of-band modifications

Exit codes:
  0 - No drift detected
  3 - Drift detected

Examples:
  # Check all resources for drift
  vcli diff

  # Check specific resource
  vcli diff prod-db-backup

  # JSON output for automation
  vcli diff --json
`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name != "vbr" {
			log.Fatal("This command only works with VBR at the moment.")
		}

		var resourceName string
		if len(args) > 0 {
			resourceName = args[0]
		}

		exitCode := checkDrift(resourceName)
		os.Exit(exitCode)
	},
}

func checkDrift(resourceName string) int {
	// Create state manager (read-only, no lock needed)
	stateManager := state.NewManager()

	// Load current state
	currentState, err := stateManager.LoadState()
	if err != nil {
		log.Fatalf("Failed to load state: %v", err)
	}

	if len(currentState.Resources) == 0 {
		fmt.Println("No resources in state. Use 'vcli apply' to manage resources.")
		return 0
	}

	// Filter resources if specific name provided
	resourcesToCheck := currentState.Resources
	if resourceName != "" {
		found := false
		for _, r := range currentState.Resources {
			if r.Name == resourceName {
				resourcesToCheck = []state.Resource{r}
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("Resource %q not found in state", resourceName)
		}
	}

	if !diffJSON {
		if resourceName == "" {
			fmt.Printf("Checking drift for %d resource(s)...\n\n", len(resourcesToCheck))
		}
	}

	driftDetected := false
	driftResults := []map[string]interface{}{}

	for _, stateRes := range resourcesToCheck {
		// Create resource instance based on type
		var resource resources.Resource
		var err error

		switch stateRes.Type {
		case "vbr_job":
			resource, err = resources.NewVBRJobResource(stateRes.Name, stateRes.Spec)
			if err != nil {
				fmt.Printf("Warning: Failed to create resource %s: %v\n", stateRes.Name, err)
				continue
			}
			resource.SetID(stateRes.ID)
		default:
			fmt.Printf("Warning: Unknown resource type %s for %s\n", stateRes.Type, stateRes.Name)
			continue
		}

		// Fetch current configuration from VBR
		currentResource, err := resource.Fetch(stateRes.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch %s: %v\n", stateRes.Name, err)
			continue
		}

		// Calculate drift
		changes, err := resource.Diff(currentResource)
		if err != nil {
			fmt.Printf("Warning: Failed to calculate drift for %s: %v\n", stateRes.Name, err)
			continue
		}

		// Record drift
		hasDrift := len(changes) > 0

		if diffJSON {
			driftResults = append(driftResults, map[string]interface{}{
				"resource": stateRes.Name,
				"type":     stateRes.Type,
				"drift":    hasDrift,
				"changes":  changes,
			})
		} else {
			if hasDrift {
				driftDetected = true
				fmt.Printf("✗ %s: Configuration drift detected\n", stateRes.Name)
				fmt.Printf("\n  Drift in %s:\n", stateRes.Name)
				for _, change := range changes {
					switch change.Action {
					case "add":
						fmt.Printf("    + %s: %v (added in VBR)\n", change.Path, formatValue(change.New))
					case "modify":
						fmt.Printf("    ~ %s: %v (state) → %v (VBR)\n", change.Path, formatValue(change.Old), formatValue(change.New))
					case "delete":
						fmt.Printf("    - %s: %v (removed from VBR)\n", change.Path, formatValue(change.Old))
					}
				}
				fmt.Printf("\n  Possible causes:\n")
				fmt.Printf("    • Manual changes in VBR UI\n")
				fmt.Printf("    • Another tool modified the job\n")
				fmt.Printf("    • Job was modified via API\n")
				fmt.Printf("\n  To fix: vcli apply <config-file>.yaml\n\n")
			} else {
				fmt.Printf("✓ %s: No drift detected\n", stateRes.Name)
			}
		}

		if hasDrift {
			driftDetected = true
		}
	}

	// Handle JSON output
	if diffJSON {
		output := map[string]interface{}{
			"resources": driftResults,
			"drift":     driftDetected,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}

		fmt.Println(string(jsonBytes))
	} else {
		// Summary
		if resourceName == "" {
			syncCount := 0
			driftCount := 0
			for _, result := range resourcesToCheck {
				// Recalculate for summary
				var resource resources.Resource
				var err error

				switch result.Type {
				case "vbr_job":
					resource, err = resources.NewVBRJobResource(result.Name, result.Spec)
					if err != nil {
						continue
					}
					resource.SetID(result.ID)
				default:
					continue
				}

				currentResource, err := resource.Fetch(result.ID)
				if err != nil {
					continue
				}

				changes, err := resource.Diff(currentResource)
				if err != nil {
					continue
				}

				if len(changes) > 0 {
					driftCount++
				} else {
					syncCount++
				}
			}

			fmt.Printf("\nSummary: %d resource(s) with drift, %d resource(s) in sync\n", driftCount, syncCount)
		}
	}

	if driftDetected {
		return 3
	}
	return 0
}

func init() {
	diffCmd.Flags().BoolVar(&diffJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(diffCmd)
}
