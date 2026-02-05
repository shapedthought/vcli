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
		// Try simplified format first (storage.compression)
		if comp, ok := storage["compression"].(string); ok {
			fmt.Printf("    Compression: %s\n", comp)
		} else {
			// Try full export format (storage.advancedSettings.storageData.compressionLevel)
			if advSettings, ok := storage["advancedSettings"].(map[string]interface{}); ok {
				if storageData, ok := advSettings["storageData"].(map[string]interface{}); ok {
					if comp, ok := storageData["compressionLevel"].(string); ok {
						fmt.Printf("    Compression: %s\n", comp)
					}
				}
			}
		}

		// Try simplified format first (storage.encryption)
		if enc, ok := storage["encryption"].(bool); ok {
			fmt.Printf("    Encryption:  %v\n", enc)
		} else {
			// Try full export format (storage.advancedSettings.storageData.encryption.isEnabled)
			if advSettings, ok := storage["advancedSettings"].(map[string]interface{}); ok {
				if storageData, ok := advSettings["storageData"].(map[string]interface{}); ok {
					if encryption, ok := storageData["encryption"].(map[string]interface{}); ok {
						if isEnabled, ok := encryption["isEnabled"].(bool); ok {
							fmt.Printf("    Encryption:  %v", isEnabled)
							// Show encryption type if available
							if encType, ok := encryption["encryptionType"].(string); ok {
								fmt.Printf(" (%s)", encType)
							}
							fmt.Println()
						}
					}
				}
			}
		}

		// Try simplified format first (storage.retention)
		if ret, ok := storage["retention"].(map[string]interface{}); ok {
			retType := ret["type"]
			retQty := ret["quantity"]
			fmt.Printf("    Retention:   %v %s\n", retQty, retType)
		} else {
			// Try full export format (storage.retentionPolicy)
			if ret, ok := storage["retentionPolicy"].(map[string]interface{}); ok {
				retType := ret["type"]
				retQty := ret["quantity"]
				fmt.Printf("    Retention:   %v %s\n", retQty, retType)
			}
		}
	}

	// Show schedule if present
	if schedule, ok := finalSpec.Spec["schedule"].(map[string]interface{}); ok {
		fmt.Println("\n  Schedule Settings:")

		// Try simplified format first (schedule.enabled)
		if enabled, ok := schedule["enabled"].(bool); ok {
			fmt.Printf("    Enabled: %v\n", enabled)
		} else {
			// Try full export format (schedule.runAutomatically or schedule.daily.isEnabled)
			if runAuto, ok := schedule["runAutomatically"].(bool); ok {
				fmt.Printf("    Run Automatically: %v\n", runAuto)
			}
		}

		// Try simplified format first (schedule.daily as string)
		if daily, ok := schedule["daily"].(string); ok {
			fmt.Printf("    Daily:   %s\n", daily)
		} else {
			// Try full export format (schedule.daily as object)
			if dailyObj, ok := schedule["daily"].(map[string]interface{}); ok {
				if isEnabled, ok := dailyObj["isEnabled"].(bool); ok && isEnabled {
					if localTime, ok := dailyObj["localTime"].(string); ok {
						fmt.Printf("    Daily:   %s\n", localTime)
					}
					if dailyKind, ok := dailyObj["dailyKind"].(string); ok {
						fmt.Printf("    Kind:    %s\n", dailyKind)
					}
				}
			}
		}

		// Try simplified format first (schedule.retry)
		if retry, ok := schedule["retry"].(map[string]interface{}); ok {
			if enabled, ok := retry["enabled"].(bool); ok && enabled {
				times := retry["times"]
				wait := retry["wait"]
				fmt.Printf("    Retry:   %v times, wait %v minutes\n", times, wait)
			} else {
				// Try full export format (retry.isEnabled)
				if isEnabled, ok := retry["isEnabled"].(bool); ok && isEnabled {
					retryCount := retry["retryCount"]
					awaitMinutes := retry["awaitMinutes"]
					fmt.Printf("    Retry:   %v times, wait %v minutes\n", retryCount, awaitMinutes)
				}
			}
		}
	}

	// Show VMs/objects
	fmt.Println("\n  Backup Objects:")

	// Try simplified format first (objects array)
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
	} else {
		// Try full export format (virtualMachines.includes array)
		if virtualMachines, ok := finalSpec.Spec["virtualMachines"].(map[string]interface{}); ok {
			if includes, ok := virtualMachines["includes"].([]interface{}); ok {
				fmt.Printf("    Total: %d object(s)\n", len(includes))
				for i, obj := range includes {
					if objMap, ok := obj.(map[string]interface{}); ok {
						name := objMap["name"]
						typ := objMap["type"]
						hostName := objMap["hostName"]
						if hostName != nil && hostName != "" {
							fmt.Printf("    %d. %s (%s) on %s\n", i+1, name, typ, hostName)
						} else {
							fmt.Printf("    %d. %s (%s)\n", i+1, name, typ)
						}
					}
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
