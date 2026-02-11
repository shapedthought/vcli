package cmd

import (
	"fmt"
	"log"

	"github.com/shapedthought/owlctl/config"
	"github.com/spf13/cobra"
)

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage named instances",
	Long: `Instance commands for listing and inspecting named server connections
defined in owlctl.yaml.

Instances define named server connections with product type, URL, port,
credential references, and TLS settings. They replace the simpler "targets"
and support multi-server automation.

Example owlctl.yaml:
  instances:
    vbr-prod:
      product: vbr
      url: https://vbr-prod.example.com
      credentialRef: PROD
      description: Production VBR server
    vbr-dr:
      product: vbr
      url: https://vbr-dr.example.com
      credentialRef: DR
      description: DR site VBR server

Commands:
  owlctl instance list              List all defined instances
  owlctl instance show <name>       Show details of a specific instance
`,
}

var instanceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances defined in owlctl.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}

		names := cfg.ListInstances()
		if len(names) == 0 {
			fmt.Println("No instances defined in owlctl.yaml.")
			return
		}

		fmt.Printf("%-20s %-10s %-40s %-15s %-30s\n", "NAME", "PRODUCT", "URL", "CREDENTIAL REF", "DESCRIPTION")
		fmt.Printf("%-20s %-10s %-40s %-15s %-30s\n", "----", "-------", "---", "--------------", "-----------")

		for _, name := range names {
			inst, _ := cfg.GetInstance(name)

			desc := inst.Description
			if len(desc) > 28 {
				desc = desc[:25] + "..."
			}

			credRef := inst.CredentialRef
			if credRef == "" {
				credRef = "(default)"
			}

			url := inst.URL
			if len(url) > 38 {
				url = url[:35] + "..."
			}

			fmt.Printf("%-20s %-10s %-40s %-15s %-30s\n", name, inst.Product, url, credRef, desc)
		}
	},
}

var instanceShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a specific instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}

		name := args[0]
		inst, err := cfg.GetInstance(name)
		if err != nil {
			log.Fatalf("Instance error: %v", err)
		}

		fmt.Printf("Instance: %s\n", name)
		if inst.Description != "" {
			fmt.Printf("Description: %s\n", inst.Description)
		}
		fmt.Println()
		fmt.Printf("Product: %s\n", inst.Product)
		fmt.Printf("URL: %s\n", inst.URL)
		if inst.Port != 0 {
			fmt.Printf("Port: %d\n", inst.Port)
		} else {
			fmt.Println("Port: (product default)")
		}
		if inst.Insecure != nil {
			fmt.Printf("Insecure: %v\n", *inst.Insecure)
		} else {
			fmt.Println("Insecure: (use global setting)")
		}
		if inst.CredentialRef != "" {
			fmt.Printf("Credential Ref: %s\n", inst.CredentialRef)
			fmt.Printf("  Username env: OWLCTL_%s_USERNAME\n", inst.CredentialRef)
			fmt.Printf("  Password env: OWLCTL_%s_PASSWORD\n", inst.CredentialRef)
		} else {
			fmt.Println("Credential Ref: (default — OWLCTL_USERNAME / OWLCTL_PASSWORD)")
		}
	},
}

func init() {
	instanceCmd.AddCommand(instanceListCmd)
	instanceCmd.AddCommand(instanceShowCmd)
	rootCmd.AddCommand(instanceCmd)
}
