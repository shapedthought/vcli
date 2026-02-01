package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	applyDryRun bool
	applyForce  bool
)

var applyCmd = &cobra.Command{
	Use:   "apply [file]",
	Short: "Create or update resources from YAML configuration",
	Long: `Apply creates or updates VBR resources based on declarative YAML configuration.

The apply command:
- Creates new resources if they don't exist
- Updates existing resources when configuration changes
- Is idempotent (no changes if config matches reality)
- Tracks state in .vcli/state.json

Examples:
  # Apply configuration
  vcli apply backup-job.yaml

  # Preview changes without applying
  vcli apply backup-job.yaml --dry-run

  # Skip confirmation prompt
  vcli apply backup-job.yaml --force
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name != "vbr" {
			log.Fatal("This command only works with VBR at the moment.")
		}

		applyResource(args[0])
	},
}

func applyResource(filename string) {
	// Load and validate configuration file
	resourceSpec, err := loadResourceSpec(filename)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate resource spec
	if err := validateResourceSpec(resourceSpec); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create state manager
	stateManager := state.NewManager()

	// Acquire state lock (unless dry-run)
	if !applyDryRun {
		if err := stateManager.Lock(); err != nil {
			log.Fatalf("Failed to acquire state lock: %v", err)
		}
		defer func() {
			if err := stateManager.Unlock(); err != nil {
				fmt.Printf("Warning: Failed to release state lock: %v\n", err)
			}
		}()
	}

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

	fmt.Printf("Applying: %s (%s)\n\n", resource.Name(), resourceSpec.Kind)

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

	// Display changes
	if len(changes) == 0 {
		fmt.Println("No changes detected. Resource is up to date.")
		return
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

	// Stop if dry-run
	if applyDryRun {
		fmt.Println("\nDry-run mode: no changes applied")
		return
	}

	// Prompt for confirmation (unless --force)
	if !applyForce {
		if !promptConfirmation("Apply these changes?") {
			fmt.Println("Apply cancelled")
			return
		}
	}

	fmt.Println()

	// Execute changes
	if action == "create" {
		fmt.Printf("Creating VBR job %q...\n", resource.Name())
		if err := resource.Create(); err != nil {
			log.Fatalf("Failed to create resource: %v", err)
		}
		fmt.Printf("✓ Successfully created job (ID: %s)\n", resource.ID())
	} else {
		fmt.Printf("Updating VBR job %q...\n", resource.Name())
		if err := resource.Update(currentResource); err != nil {
			log.Fatalf("Failed to update resource: %v", err)
		}
		fmt.Printf("✓ Successfully updated job (ID: %s)\n", resource.ID())
	}

	// Update state
	stateResource := state.Resource{
		Type: resource.Type(),
		Name: resource.Name(),
		ID:   resource.ID(),
		Spec: resource.Spec(),
	}

	if err := stateManager.UpsertResource(stateResource); err != nil {
		log.Fatalf("Failed to update state: %v", err)
	}

	fmt.Println("✓ State updated")
}

func loadResourceSpec(filename string) (*resources.ResourceSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec resources.ResourceSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &spec, nil
}

func validateResourceSpec(spec *resources.ResourceSpec) error {
	if spec.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if spec.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if spec.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if spec.Spec == nil {
		return fmt.Errorf("spec is required")
	}
	return nil
}

func createResourceFromSpec(spec *resources.ResourceSpec) (resources.Resource, error) {
	switch spec.Kind {
	case "VBRJob":
		return resources.NewVBRJobResource(spec.Metadata.Name, spec.Spec)
	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", spec.Kind)
	}
}

func displayChanges(changes []resources.Change) {
	fmt.Println("Changes:")
	for _, change := range changes {
		switch change.Action {
		case "add":
			fmt.Printf("  + %s: %v\n", change.Path, formatValue(change.New))
		case "modify":
			fmt.Printf("  ~ %s: %v → %v\n", change.Path, formatValue(change.Old), formatValue(change.New))
		case "delete":
			fmt.Printf("  - %s: %v\n", change.Path, formatValue(change.Old))
		}
	}
}

func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}
		// Show first few items for arrays
		if len(val) <= 3 {
			parts := make([]string, len(val))
			for i, item := range val {
				parts[i] = formatValue(item)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return fmt.Sprintf("[%d items]", len(val))
	case map[string]interface{}:
		return fmt.Sprintf("{%d fields}", len(val))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func promptConfirmation(message string) bool {
	fmt.Printf("\n%s (y/n): ", message)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func init() {
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Preview changes without applying")
	applyCmd.Flags().BoolVar(&applyForce, "force", false, "Skip confirmation prompt")
	rootCmd.AddCommand(applyCmd)
}
