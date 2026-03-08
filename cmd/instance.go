package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/shapedthought/owlctl/config"
	"github.com/shapedthought/owlctl/utils"
	"github.com/spf13/cobra"
)

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage named instances",
	Long: `Instance commands for managing named server connections defined in owlctl.yaml.

Instances define named server connections with product type, URL, port,
credential references, and TLS settings. They replace the simpler "targets"
and support multi-server automation.

Commands:
  owlctl instance add <name>        Add or update an instance in owlctl.yaml
  owlctl instance remove <name>     Remove an instance from owlctl.yaml
  owlctl instance set <name>        Set the default instance (persisted to settings.json)
  owlctl instance get               Show the current default instance
  owlctl instance unset             Clear the default instance
  owlctl instance list              List all defined instances (* marks the default)
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

		settings := utils.ReadSettings()
		defaultInstance := settings.DefaultInstance

		fmt.Printf("%-2s %-20s %-10s %-40s %-15s %-30s\n", "", "NAME", "PRODUCT", "URL", "CREDENTIAL REF", "DESCRIPTION")
		fmt.Printf("%-2s %-20s %-10s %-40s %-15s %-30s\n", "", "----", "-------", "---", "--------------", "-----------")

		for _, name := range names {
			inst, err := cfg.GetInstance(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: instance %q: %v\n", name, err)
				continue
			}

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

			marker := ""
			if name == defaultInstance {
				marker = "*"
			}

			fmt.Printf("%-2s %-20s %-10s %-40s %-15s %-30s\n", marker, name, inst.Product, url, credRef, desc)
		}

		if defaultInstance != "" {
			fmt.Println("\n* = default instance (set via 'owlctl instance set')")
		}
	},
}

var instanceSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set the default instance in settings.json",
	Long: `Set a named instance as the default so --instance is not needed on every command.

The instance must exist in owlctl.yaml. Use 'owlctl instance unset' to clear.

Examples:
  owlctl instance set vbr-prod
  owlctl instance get
  owlctl instance unset
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}

		if _, err := cfg.GetInstance(name); err != nil {
			log.Fatalf("Cannot set default: %v", err)
		}

		settings := utils.ReadSettings()
		settings.DefaultInstance = name
		if err := utils.WriteSettings(settings); err != nil {
			log.Fatalf("Failed to save settings: %v", err)
		}

		fmt.Printf("Default instance set to %q.\n", name)
		fmt.Println("Run 'owlctl instance unset' to clear.")
	},
}

var instanceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show the current default instance from settings.json",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		settings := utils.ReadSettings()
		if settings.DefaultInstance == "" {
			fmt.Println("(none)")
		} else {
			fmt.Println(settings.DefaultInstance)
		}
	},
}

var instanceUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Clear the default instance from settings.json",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		settings := utils.ReadSettings()
		if settings.DefaultInstance == "" {
			fmt.Println("No default instance is set.")
			return
		}
		prev := settings.DefaultInstance
		settings.DefaultInstance = ""
		if err := utils.WriteSettings(settings); err != nil {
			log.Fatalf("Failed to save settings: %v", err)
		}
		fmt.Printf("Default instance %q cleared.\n", prev)
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

var (
	instanceAddProduct       string
	instanceAddURL           string
	instanceAddPort          int
	instanceAddCredentialRef string
	instanceAddDescription   string
	instanceAddInsecure      bool
	instanceAddForce         bool
)

var validProducts = []string{"vbr", "ent_man", "vb365", "vone", "aws", "azure", "gcp"}

var instanceAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update an instance in owlctl.yaml",
	Long: `Add a named instance to owlctl.yaml, or update it if it already exists.

If owlctl.yaml does not exist it will be created. If OWLCTL_CONFIG is set,
the file at that path is used instead.

Examples:
  owlctl instance add vbr-prod --url vbr-prod.example.com --product vbr
  owlctl instance add vbr-prod --url vbr-prod.example.com --product vbr --credential-ref PROD --description "Production VBR"
  owlctl instance add vbr-dr   --url vbr-dr.example.com   --product vbr --insecure
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if instanceAddURL == "" {
			log.Fatal("--url is required")
		}
		if instanceAddProduct == "" {
			log.Fatal("--product is required")
		}

		validProduct := false
		for _, p := range validProducts {
			if instanceAddProduct == p {
				validProduct = true
				break
			}
		}
		if !validProduct {
			log.Fatalf("invalid product %q — must be one of: %v", instanceAddProduct, validProducts)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}

		if _, exists := cfg.Instances[name]; exists && !instanceAddForce {
			log.Fatalf("Instance %q already exists. Use --force to overwrite.", name)
		}

		inst := config.InstanceConfig{
			Product:       instanceAddProduct,
			URL:           instanceAddURL,
			CredentialRef: instanceAddCredentialRef,
			Description:   instanceAddDescription,
		}
		if instanceAddPort != 0 {
			inst.Port = instanceAddPort
		}
		if cmd.Flags().Changed("insecure") {
			inst.Insecure = &instanceAddInsecure
		}

		cfg.AddInstance(name, inst)

		if err := cfg.Save(); err != nil {
			log.Fatalf("Failed to save owlctl.yaml: %v", err)
		}

		fmt.Printf("Instance %q saved to owlctl.yaml.\n", name)
		if inst.CredentialRef != "" {
			fmt.Printf("Set credentials: OWLCTL_%s_USERNAME / OWLCTL_%s_PASSWORD\n", inst.CredentialRef, inst.CredentialRef)
		} else {
			fmt.Println("No credential ref set — will use OWLCTL_USERNAME / OWLCTL_PASSWORD.")
		}
	},
}

var instanceRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an instance from owlctl.yaml",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}

		if err := cfg.RemoveInstance(name); err != nil {
			log.Fatalf("%v", err)
		}

		if err := cfg.Save(); err != nil {
			log.Fatalf("Failed to save owlctl.yaml: %v", err)
		}

		fmt.Printf("Instance %q removed from owlctl.yaml.\n", name)
	},
}

func init() {
	instanceAddCmd.Flags().StringVar(&instanceAddURL, "url", "", "Server hostname or IP (required)")
	instanceAddCmd.Flags().StringVar(&instanceAddProduct, "product", "", "Veeam product: vbr, ent_man, vb365, vone, aws, azure, gcp (required)")
	instanceAddCmd.Flags().IntVar(&instanceAddPort, "port", 0, "Port override (default: product default)")
	instanceAddCmd.Flags().StringVar(&instanceAddCredentialRef, "credential-ref", "", "Credential ref (reads OWLCTL_{REF}_USERNAME / _PASSWORD)")
	instanceAddCmd.Flags().StringVar(&instanceAddDescription, "description", "", "Human-readable description")
	instanceAddCmd.Flags().BoolVar(&instanceAddInsecure, "insecure", false, "Skip TLS verification for this instance")
	instanceAddCmd.Flags().BoolVar(&instanceAddForce, "force", false, "Overwrite if instance already exists")

	instanceCmd.AddCommand(instanceAddCmd)
	instanceCmd.AddCommand(instanceRemoveCmd)
	instanceCmd.AddCommand(instanceSetCmd)
	instanceCmd.AddCommand(instanceGetCmd)
	instanceCmd.AddCommand(instanceUnsetCmd)
	instanceCmd.AddCommand(instanceListCmd)
	instanceCmd.AddCommand(instanceShowCmd)
	rootCmd.AddCommand(instanceCmd)
}
