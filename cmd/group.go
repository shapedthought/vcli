package cmd

import (
	"fmt"
	"log"

	"github.com/shapedthought/owlctl/config"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups of spec files",
	Long: `Group commands for listing and inspecting named groups defined in owlctl.yaml.

Groups are named collections of spec files with an optional profile (base standard)
and overlay (policy patch), enabling batch policy application and standardisation.

Example owlctl.yaml:
  groups:
    sql-tier:
      description: SQL Server backup group
      profile: profiles/gold.yaml
      overlay: overlays/compliance.yaml
      specs:
        - specs/sql-vm-01.yaml
        - specs/sql-vm-02.yaml

Commands:
  owlctl group list              List all defined groups
  owlctl group show <name>       Show details of a specific group
`,
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups defined in owlctl.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}
		cfg.WarnDeprecatedFields()

		names := cfg.ListGroups()
		if len(names) == 0 {
			fmt.Println("No groups defined in owlctl.yaml.")
			return
		}

		fmt.Printf("%-20s %-40s %-6s %-8s %-8s\n", "NAME", "DESCRIPTION", "SPECS", "PROFILE", "OVERLAY")
		fmt.Printf("%-20s %-40s %-6s %-8s %-8s\n", "----", "-----------", "-----", "-------", "-------")

		for _, name := range names {
			group, _ := cfg.GetGroup(name)

			desc := group.Description
			if len(desc) > 38 {
				desc = desc[:35] + "..."
			}

			profileFlag := "-"
			if group.Profile != "" {
				profileFlag = "yes"
			}
			overlayFlag := "-"
			if group.Overlay != "" {
				overlayFlag = "yes"
			}

			fmt.Printf("%-20s %-40s %-6d %-8s %-8s\n", name, desc, len(group.Specs), profileFlag, overlayFlag)
		}
	},
}

var groupShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a specific group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}
		cfg.WarnDeprecatedFields()

		name := args[0]
		group, err := cfg.GetGroup(name)
		if err != nil {
			log.Fatalf("Group error: %v", err)
		}

		fmt.Printf("Group: %s\n", name)
		if group.Description != "" {
			fmt.Printf("Description: %s\n", group.Description)
		}
		fmt.Println()

		if group.Profile != "" {
			fmt.Printf("Profile: %s\n", group.Profile)
			fmt.Printf("  (resolved: %s)\n", cfg.ResolvePath(group.Profile))
		} else {
			fmt.Println("Profile: (none)")
		}

		if group.Overlay != "" {
			fmt.Printf("Overlay: %s\n", group.Overlay)
			fmt.Printf("  (resolved: %s)\n", cfg.ResolvePath(group.Overlay))
		} else {
			fmt.Println("Overlay: (none)")
		}

		fmt.Printf("\nSpecs (%d):\n", len(group.Specs))
		for i, spec := range group.Specs {
			fmt.Printf("  %d. %s\n", i+1, spec)
		}
	},
}

func init() {
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupShowCmd)
	rootCmd.AddCommand(groupCmd)
}
