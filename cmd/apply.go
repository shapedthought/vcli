package cmd

import (
	"fmt"
	"log"

	"github.com/shapedthought/vcli/config"
	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/utils"
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
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
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
		fmt.Println("\n=== Dry Run Mode ===")
		fmt.Printf("Resource: %s (%s)\n", finalSpec.Metadata.Name, finalSpec.Kind)
		fmt.Printf("Configuration to be applied:\n")

		// Display key fields
		if desc, ok := finalSpec.Spec["description"].(string); ok {
			fmt.Printf("  Description: %s\n", desc)
		}
		if repo, ok := finalSpec.Spec["repository"].(string); ok {
			fmt.Printf("  Repository: %s\n", repo)
		}
		if typ, ok := finalSpec.Spec["type"].(string); ok {
			fmt.Printf("  Type: %s\n", typ)
		}

		// Show storage settings
		if storage, ok := finalSpec.Spec["storage"].(map[string]interface{}); ok {
			fmt.Println("  Storage:")
			if comp, ok := storage["compression"].(string); ok {
				fmt.Printf("    Compression: %s\n", comp)
			}
			if ret, ok := storage["retention"].(map[string]interface{}); ok {
				if typ, ok := ret["type"].(string); ok {
					fmt.Printf("    Retention Type: %s\n", typ)
				}
				if qty, ok := ret["quantity"].(int); ok {
					fmt.Printf("    Retention Quantity: %d\n", qty)
				}
			}
		}

		// Show VMs
		if objects, ok := finalSpec.Spec["objects"].([]interface{}); ok {
			fmt.Printf("  Objects (%d):\n", len(objects))
			for _, obj := range objects {
				if objMap, ok := obj.(map[string]interface{}); ok {
					name := objMap["name"]
					typ := objMap["type"]
					fmt.Printf("    - %s (%s)\n", name, typ)
				}
			}
		}

		fmt.Println("\n=== End Dry Run ===")
		fmt.Println("Note: This is a preview. No changes have been made.")
		fmt.Println("Remove --dry-run flag to apply these changes.")
		return
	}

	// TODO: Implement actual job creation/update
	// This will be implemented in subsequent commits
	fmt.Printf("\nApply functionality not yet implemented.\n")
	fmt.Printf("Resource to apply: %s (%s)\n", finalSpec.Metadata.Name, finalSpec.Kind)
	fmt.Println("This will be implemented in the next phase.")
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

func init() {
	applyCmd.Flags().StringVarP(&overlayFile, "overlay", "o", "", "Overlay file to merge with base configuration")
	applyCmd.Flags().StringVar(&environment, "env", "", "Environment to use (looks up overlay from vcli.yaml)")
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying them")

	jobsCmd.AddCommand(applyCmd)
}
