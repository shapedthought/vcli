package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/shapedthought/owlctl/config"
	"github.com/spf13/cobra"
)

var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Manage named VBR server targets",
	Long: `Target commands for listing named VBR server connections defined in owlctl.yaml.

Targets let you switch between VBR servers without changing environment variables.
Use --target <name> on any command to override OWLCTL_URL.

Example owlctl.yaml:
  targets:
    primary:
      url: https://vbr-prod.example.com
      description: Production VBR server
    dr:
      url: https://vbr-dr.example.com
      description: Disaster recovery site

Usage:
  owlctl target list              List all defined targets
  owlctl get jobs --target dr     Run command against DR server
  owlctl login --target primary   Authenticate against production
`,
}

var targetListJSON bool

var targetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all targets defined in owlctl.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}

		names := cfg.ListTargets()
		if len(names) == 0 {
			fmt.Println("No targets defined in owlctl.yaml.")
			fmt.Println()
			fmt.Println("Add targets to your owlctl.yaml:")
			fmt.Println()
			fmt.Println("  targets:")
			fmt.Println("    primary:")
			fmt.Println("      url: https://vbr-prod.example.com")
			fmt.Println("      description: Production VBR server")
			return
		}

		if targetListJSON {
			out := make(map[string]config.TargetConfig)
			for _, name := range names {
				target, _ := cfg.GetTarget(name)
				out[name] = target
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				log.Fatalf("Failed to marshal JSON: %v", err)
			}
			fmt.Println(string(data))
			return
		}

		fmt.Printf("%-20s %-40s %s\n", "NAME", "URL", "DESCRIPTION")
		fmt.Printf("%-20s %-40s %s\n", "----", "---", "-----------")

		for _, name := range names {
			target, _ := cfg.GetTarget(name)

			desc := target.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}

			url := target.URL
			if len(url) > 38 {
				url = url[:35] + "..."
			}

			fmt.Printf("%-20s %-40s %s\n", name, url, desc)
		}
	},
}

func init() {
	targetListCmd.Flags().BoolVar(&targetListJSON, "json", false, "Output as JSON")
	targetCmd.AddCommand(targetListCmd)
	rootCmd.AddCommand(targetCmd)
}
