package cmd

import (
	"fmt"
	"log"

	"github.com/shapedthought/vcli/config"
	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	planOverlayFile string
	planEnvironment string
	planShowYAML    bool
)

var planCmd = &cobra.Command{
	Use:   "plan [config-file]",
	Short: "Preview merged configuration without applying",
	Long: `Plan shows the final merged configuration that would be applied.

The plan command is useful for:
- Previewing overlay merges before applying
- Validating YAML configuration syntax
- Understanding what the final job configuration will look like
- CI/CD pipeline validation (non-destructive)

Examples:
  # Preview base configuration
  vcli job plan backup-job.yaml

  # Preview with overlay
  vcli job plan base-job.yaml -o prod-overlay.yaml

  # Preview with environment overlay
  vcli job plan base-job.yaml --env production

  # Show full YAML output
  vcli job plan base-job.yaml -o prod-overlay.yaml --show-yaml

Note: This command shows the merged configuration but does not compare
against current VBR state. Full drift detection will be available in Phase 2.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		planJob(args[0])
	},
}

func planJob(configFile string) {
	settings := utils.ReadSettings()

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

	// Determine which overlay to use (same logic as apply command)
	var finalSpec resources.ResourceSpec
	var overlayUsed string

	if planOverlayFile != "" {
		// Explicit overlay file provided
		overlayUsed = planOverlayFile
		mergedSpec, err := resources.MergeYAMLFiles(configFile, planOverlayFile, resources.DefaultMergeOptions())
		if err != nil {
			log.Fatalf("Failed to merge with overlay: %v", err)
		}
		finalSpec = mergedSpec
	} else if planEnvironment != "" || needsPlanConfigOverlay() {
		// Try to use overlay from vcli.yaml
		overlayPath, err := getPlanConfiguredOverlay()
		if err != nil {
			// If no overlay configured, just use base spec
			overlayUsed = "none"
			finalSpec = baseSpec
		} else {
			overlayUsed = overlayPath
			mergedSpec, err := resources.MergeYAMLFiles(configFile, overlayPath, resources.DefaultMergeOptions())
			if err != nil {
				log.Fatalf("Failed to merge with configured overlay: %v", err)
			}
			finalSpec = mergedSpec
		}
	} else {
		// No overlay specified
		overlayUsed = "none"
		finalSpec = baseSpec
	}

	// Display plan header
	fmt.Println("╔════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Configuration Plan Preview                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("Resource Name: %s\n", finalSpec.Metadata.Name)
	fmt.Printf("Resource Type: %s\n", finalSpec.Kind)
	fmt.Printf("Base Config:   %s\n", configFile)
	fmt.Printf("Overlay:       %s\n", overlayUsed)
	fmt.Println()

	// Show labels if present
	if len(finalSpec.Metadata.Labels) > 0 {
		fmt.Println("Labels:")
		for k, v := range finalSpec.Metadata.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Println()
	}

	// Display merged configuration details
	fmt.Println("Merged Configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Display key fields
	if desc, ok := finalSpec.Spec["description"].(string); ok {
		fmt.Printf("  Description: %s\n", desc)
	}
	if repo, ok := finalSpec.Spec["repository"].(string); ok {
		fmt.Printf("  Repository:  %s\n", repo)
	}
	if typ, ok := finalSpec.Spec["type"].(string); ok {
		fmt.Printf("  Job Type:    %s\n", typ)
	}
	if disabled, ok := finalSpec.Spec["isDisabled"].(bool); ok {
		fmt.Printf("  Disabled:    %v\n", disabled)
	}

	// Show storage settings
	fmt.Println("\n  Storage Settings:")
	if storage, ok := finalSpec.Spec["storage"].(map[string]interface{}); ok {
		if comp, ok := storage["compression"].(string); ok {
			fmt.Printf("    Compression: %s\n", comp)
		}
		if enc, ok := storage["encryption"].(bool); ok {
			fmt.Printf("    Encryption:  %v\n", enc)
		}
		if ret, ok := storage["retention"].(map[string]interface{}); ok {
			retType := ret["type"]
			retQty := ret["quantity"]
			fmt.Printf("    Retention:   %v %s\n", retQty, retType)
		}
	}

	// Show schedule if present
	if schedule, ok := finalSpec.Spec["schedule"].(map[string]interface{}); ok {
		fmt.Println("\n  Schedule Settings:")
		if enabled, ok := schedule["enabled"].(bool); ok {
			fmt.Printf("    Enabled: %v\n", enabled)
		}
		if daily, ok := schedule["daily"].(string); ok {
			fmt.Printf("    Daily:   %s\n", daily)
		}
		if retry, ok := schedule["retry"].(map[string]interface{}); ok {
			if enabled, ok := retry["enabled"].(bool); ok && enabled {
				times := retry["times"]
				wait := retry["wait"]
				fmt.Printf("    Retry:   %v times, wait %v minutes\n", times, wait)
			}
		}
	}

	// Show VMs/objects
	fmt.Println("\n  Backup Objects:")
	if objects, ok := finalSpec.Spec["objects"].([]interface{}); ok {
		fmt.Printf("    Total: %d object(s)\n", len(objects))
		for i, obj := range objects {
			if objMap, ok := obj.(map[string]interface{}); ok {
				name := objMap["name"]
				typ := objMap["type"]
				hostName := objMap["hostName"]
				if hostName != nil {
					fmt.Printf("    %d. %s (%s) on %s\n", i+1, name, typ, hostName)
				} else {
					fmt.Printf("    %d. %s (%s)\n", i+1, name, typ)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Show full YAML if requested
	if planShowYAML {
		fmt.Println("\nFull YAML Configuration:")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		yamlBytes, err := yaml.Marshal(finalSpec)
		if err != nil {
			log.Fatalf("Failed to marshal YAML: %v", err)
		}
		fmt.Println(string(yamlBytes))
		fmt.Println("─────────────────────────────────────────────────────────────────")
	}

	// Show next steps
	fmt.Println("\nNext Steps:")
	fmt.Printf("  To apply this configuration:\n")
	if overlayUsed != "none" {
		if planOverlayFile != "" {
			fmt.Printf("    vcli job apply %s -o %s\n", configFile, planOverlayFile)
		} else {
			fmt.Printf("    vcli job apply %s --env %s\n", configFile, planEnvironment)
		}
	} else {
		fmt.Printf("    vcli job apply %s\n", configFile)
	}
	fmt.Println()
}

// needsPlanConfigOverlay checks if we should try to use vcli.yaml overlay
func needsPlanConfigOverlay() bool {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}

	if cfg.CurrentEnvironment != "" {
		_, err := cfg.GetEnvironmentOverlay("")
		return err == nil
	}

	return false
}

// getPlanConfiguredOverlay gets the overlay path from vcli.yaml
func getPlanConfiguredOverlay() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load vcli.yaml: %w", err)
	}

	env := planEnvironment
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
	planCmd.Flags().StringVarP(&planOverlayFile, "overlay", "o", "", "Overlay file to merge with base configuration")
	planCmd.Flags().StringVar(&planEnvironment, "env", "", "Environment to use (looks up overlay from vcli.yaml)")
	planCmd.Flags().BoolVar(&planShowYAML, "show-yaml", false, "Display full merged YAML configuration")

	jobsCmd.AddCommand(planCmd)
}
