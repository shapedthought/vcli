package cmd

import (
	"fmt"
	"log"

	"github.com/shapedthought/owlctl/config"
	"github.com/shapedthought/owlctl/utils"
	"github.com/spf13/cobra"
)

// contextCmd mirrors the kubectl context model — a familiar UX for switching
// between named Veeam environments. These commands are aliases for the
// equivalent `owlctl instance` operations.
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Switch and inspect the active environment context",
	Long: `Manage the active context (named instance) for owlctl commands.

Contexts map directly to named instances defined in owlctl.yaml. Use
'context use' to switch environments — equivalent to 'instance set',
but more natural for users familiar with kubectl.

Examples:
  owlctl context list               # list all defined contexts
  owlctl context use vbr-prod       # switch active context
  owlctl context current            # show the active context
`,
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available contexts",
	Long: `List all named instances defined in owlctl.yaml. The active context is marked with *.

Examples:
  owlctl context list
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Delegate to instance list — identical output
		instanceListCmd.Run(cmd, args)
	},
}

var contextUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch the active context",
	Long: `Set a named instance as the active context. All subsequent commands will
use this environment without needing --instance.

The instance must be defined in owlctl.yaml. Use 'owlctl context use -'
to clear the active context (equivalent to 'owlctl instance unset').

Examples:
  owlctl context use vbr-prod
  owlctl context use azure-prod
  owlctl context use -              # clear active context
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Special case: "-" clears the active context (kubectl convention)
		if name == "-" {
			settings, err := utils.TryReadSettings()
			if err != nil {
				log.Fatalf("Failed to read settings.json: %v", err)
			}
			if settings.DefaultInstance == "" {
				fmt.Println("No active context is set.")
				return
			}
			prev := settings.DefaultInstance
			settings.DefaultInstance = ""
			if err := utils.WriteSettings(settings); err != nil {
				log.Fatalf("Failed to save settings: %v", err)
			}
			fmt.Printf("Switched away from context %q. No active context.\n", prev)
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load owlctl.yaml: %v", err)
		}
		if _, err := cfg.GetInstance(name); err != nil {
			log.Fatalf("Cannot switch context: %v", err)
		}

		settings, err := utils.TryReadSettings()
		if err != nil {
			log.Fatalf("Failed to read settings.json: %v", err)
		}
		settings.DefaultInstance = name
		if err := utils.WriteSettings(settings); err != nil {
			log.Fatalf("Failed to save settings: %v", err)
		}

		fmt.Printf("Switched to context %q.\n", name)
	},
}

var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active context",
	Long: `Print the name of the currently active context.

Examples:
  owlctl context current
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		settings, _ := utils.TryReadSettings()
		if settings.DefaultInstance == "" {
			fmt.Println("(none)")
		} else {
			fmt.Println(settings.DefaultInstance)
		}
	},
}

func init() {
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	rootCmd.AddCommand(contextCmd)
}
