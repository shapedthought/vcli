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

var planJSON bool

var planCmd = &cobra.Command{
	Use:   "plan [file]",
	Short: "Preview changes without applying them",
	Long: `Plan shows what changes would be made by vcli apply without actually executing them.

This is useful for:
- Previewing changes before applying
- Validating YAML configuration
- CI/CD pipelines (non-destructive checks)

Exit codes:
  0 - No changes needed
  2 - Changes would be applied

Examples:
  # Preview changes
  vcli plan backup-job.yaml

  # JSON output for CI/CD
  vcli plan backup-job.yaml --json
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name != "vbr" {
			log.Fatal("This command only works with VBR at the moment.")
		}

		exitCode := planResource(args[0])
		os.Exit(exitCode)
	},
}

func planResource(filename string) int {
	// Load and validate configuration file
	resourceSpec, err := loadResourceSpec(filename)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate resource spec
	if err := validateResourceSpec(resourceSpec); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create state manager (read-only, no lock needed)
	stateManager := state.NewManager()

	// Load current state
	currentState, err := stateManager.LoadState()
	if err != nil {
		log.Fatalf("Failed to load state: %v", err)
	}

	// Create resource from spec
	resource, err := createResourceFromSpec(resourceSpec)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	if !planJSON {
		fmt.Printf("Planning: %s (%s)\n\n", resource.Name(), resourceSpec.Kind)
	}

	// Determine action (create vs update)
	var currentResource resources.Resource
	var action string

	if stateResource, found := currentState.GetResource(resource.Name()); found {
		// Resource exists in state - fetch current from VBR
		currentResource, err = resource.Fetch(stateResource.ID)
		if err != nil {
			log.Fatalf("Failed to fetch current resource: %v", err)
		}
		resource.SetID(stateResource.ID)
		action = "update"
	} else {
		// Resource doesn't exist in state - it's a create
		action = "create"
	}

	// Calculate diff
	changes, err := resource.Diff(currentResource)
	if err != nil {
		log.Fatalf("Failed to calculate diff: %v", err)
	}

	// Handle JSON output
	if planJSON {
		outputPlanJSON(resource.Name(), resourceSpec.Kind, action, changes)
		if len(changes) == 0 {
			return 0
		}
		return 2
	}

	// Display changes
	if len(changes) == 0 {
		fmt.Println("No changes needed. Resource is up to date.")
		return 0
	}

	displayChanges(changes)

	// Show plan summary
	createCount := 0
	updateCount := 0
	if action == "create" {
		createCount = 1
	} else {
		updateCount = 1
	}

	fmt.Printf("\nPlan: %d to create, %d to update, 0 to delete\n", createCount, updateCount)
	fmt.Printf("\nNote: Run 'vcli apply %s' to execute these changes\n", filename)

	return 2
}

func outputPlanJSON(name, kind, action string, changes []resources.Change) {
	output := map[string]interface{}{
		"resource": name,
		"type":     kind,
		"action":   action,
		"changes":  changes,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonBytes))
}

func init() {
	planCmd.Flags().BoolVar(&planJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(planCmd)
}
