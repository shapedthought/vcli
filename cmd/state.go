package cmd

import (
	"fmt"
	"log"

	"github.com/shapedthought/vcli/state"
	"github.com/spf13/cobra"
)

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "State management commands",
	Long: `Commands for viewing and managing vcli state.

State is stored locally and tracks which resources have been applied or snapshotted.
`,
}

var stateHistoryCmd = &cobra.Command{
	Use:   "history [resource-name]",
	Short: "Show audit history for a resource",
	Long: `Display the audit trail of actions taken on a resource.

Each entry shows:
- Action type (snapshotted, adopted, applied)
- Timestamp
- User who performed the action
- Fields changed (for apply actions)

Examples:
  # Show history for a repository
  vcli state history "Default Backup Repository"

  # Show history for a job
  vcli state history "Backup Job 1"
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showResourceHistory(args[0])
	},
}

func showResourceHistory(resourceName string) {
	stateMgr := state.NewManager()

	resource, err := stateMgr.GetResource(resourceName)
	if err != nil {
		log.Fatalf("Resource '%s' not found in state.", resourceName)
	}

	fmt.Printf("History for %s (%s):\n\n", resource.Name, resource.Type)

	if len(resource.History) == 0 {
		fmt.Println("  No history recorded.")
		fmt.Println()
		fmt.Println("Note: History tracking was added in state version 3.")
		fmt.Println("Existing resources will start recording history on the next action.")
		return
	}

	for _, event := range resource.History {
		timestamp := event.Timestamp.Format("2006-01-02 15:04:05")

		if len(event.Fields) > 0 {
			// Show field count for apply events
			fieldInfo := fmt.Sprintf("%d field(s)", len(event.Fields))
			if event.Partial {
				fieldInfo += " [partial]"
			}
			fmt.Printf("  %s - %s by %s (%s)\n", timestamp, event.Action, event.User, fieldInfo)

			// Show first few fields
			maxFields := 3
			for i, field := range event.Fields {
				if i >= maxFields {
					fmt.Printf("      ... and %d more\n", len(event.Fields)-maxFields)
					break
				}
				fmt.Printf("      - %s\n", field)
			}
		} else {
			fmt.Printf("  %s - %s by %s\n", timestamp, event.Action, event.User)
		}
	}

	fmt.Printf("\nShowing %d event(s). Max retained: %d\n", len(resource.History), state.DefaultMaxHistoryEvents)
}

func init() {
	stateCmd.AddCommand(stateHistoryCmd)
	rootCmd.AddCommand(stateCmd)
}
