package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/shapedthought/owlctl/utils"
	"github.com/spf13/cobra"
)

var (
	listFlag    bool
	getFlag     bool
	setFlag     bool
	profileFlag bool
	tableFlag   bool
	jsonFlag    bool
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "List and manage the current profile",
	Long: `List, Get, and Set the current API profile (non-interactive).

Examples:
  # Get current profile (machine-readable)
  owlctl profile -g
  owlctl profile --get

  # Set profile (requires argument)
  owlctl profile -s vbr
  owlctl profile --set ent_man

  # List profiles (JSON by default)
  owlctl profile -l
  owlctl profile --list

  # List profiles (human-readable table)
  owlctl profile -l --table

  # Show current profile details
  owlctl profile -p
  owlctl profile --profile
`,
	Run: func(cmd *cobra.Command, args []string) {
		if getFlag {
			printCurrentProfile()
		} else if listFlag {
			listProfiles()
		} else if setFlag {
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: profile name required")
				fmt.Fprintln(os.Stderr, "Usage: owlctl profile --set <profile-name>")
				os.Exit(1)
			}
			setProfile(args[0])
		} else if profileFlag {
			utils.ReadCurrentProfile()
		} else {
			cmd.Help()
		}
	},
}

func init() {
	profileCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List available profiles (JSON by default)")
	profileCmd.Flags().BoolVarP(&getFlag, "get", "g", false, "Get current profile name")
	profileCmd.Flags().BoolVarP(&setFlag, "set", "s", false, "Set current profile (requires profile name)")
	profileCmd.Flags().BoolVarP(&profileFlag, "profile", "p", false, "Show current profile details")
	profileCmd.Flags().BoolVar(&tableFlag, "table", false, "Output as human-readable table (use with --list)")
	profileCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON (default for --list)")
	rootCmd.AddCommand(profileCmd)
}

// printCurrentProfile outputs just the profile name (machine-readable)
func printCurrentProfile() {
	settings := utils.ReadSettings()
	fmt.Println(settings.SelectedProfile)
}

// listProfiles lists available profiles in JSON or table format
func listProfiles() {
	profilesFile := utils.ReadProfilesFile()

	var names []string
	for name := range profilesFile.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	if tableFlag {
		// Human-readable table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPRODUCT\tPORT\tAPI_VERSION\tAUTH_TYPE")
		for _, name := range names {
			profile := profilesFile.Profiles[name]
			apiVersion := profile.APIVersion
			if apiVersion == "" {
				apiVersion = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
				name,
				profile.Product,
				profile.Port,
				apiVersion,
				profile.AuthType,
			)
		}
		w.Flush()
	} else {
		// JSON format (default)
		jsonOutput, err := json.Marshal(names)
		if err != nil {
			log.Fatalf("Failed to marshal profile names: %v", err)
		}
		fmt.Println(string(jsonOutput))
	}
}

// setProfile sets the current profile (non-interactive)
func setProfile(profileName string) {
	settingsPath := utils.SettingPath()
	settings := utils.ReadSettings()
	profilesFile := utils.ReadProfilesFile()

	// Validate profile exists
	if _, exists := profilesFile.Profiles[profileName]; !exists {
		fmt.Fprintf(os.Stderr, "Error: profile '%s' not found\n", profileName)
		fmt.Fprintln(os.Stderr, "Available profiles:")
		for name := range profilesFile.Profiles {
			fmt.Fprintf(os.Stderr, "  - %s\n", name)
		}
		os.Exit(2)
	}

	// Update settings
	settings.SelectedProfile = profileName

	// Save settings
	data, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		log.Fatalf("Failed to marshal settings: %v", err)
	}

	settingsFile := settingsPath + "settings.json"
	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		log.Fatalf("Failed to write settings file: %v", err)
	}

	// Success - output to stdout for confirmation
	fmt.Printf("Profile set to: %s\n", profileName)
}
