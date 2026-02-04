package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/pterm/pterm"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
)

var (
	listFlag    bool
	getFlag     bool
	setFlag     bool
	profileFlag bool
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "List and manage the current profile",
	Long:  `List, Get, and Set the current API profile`,
	Run: func(cmd *cobra.Command, args []string) {
		if getFlag {
			printCurrentProfile()
		}
		if listFlag {
			listProfiles()
		}
		if setFlag {
			setProfile()
		}
		if profileFlag {
			utils.ReadCurrentProfile()
		}
	},
}

func init() {
	profileCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "lists the available profiles")
	profileCmd.Flags().BoolVarP(&getFlag, "get", "g", false, "gets the current profile name")
	profileCmd.Flags().BoolVarP(&setFlag, "set", "s", false, "sets the current profile")
	profileCmd.Flags().BoolVarP(&profileFlag, "profile", "p", false, "shows the current profile settings")
	rootCmd.AddCommand(profileCmd)

}

func printCurrentProfile() {
	settings := utils.ReadSettings()

	fmt.Printf("Current profile: %v\n", settings.SelectedProfile)
}

func listProfiles() {
	profilesFile := utils.ReadProfilesFile()

	var names []string

	for name := range profilesFile.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	fmt.Println("Profiles available")
	for _, n := range names {
		fmt.Println(n)
	}

}

func setProfile() {
	settingsPath := utils.SettingPath()

	settings := utils.ReadSettings()

	profilesFile := utils.ReadProfilesFile()

	var names []string

	for name := range profilesFile.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	pterm.DefaultInteractiveSelect.DefaultText = "Select Profile"

	result, _ := pterm.DefaultInteractiveSelect.
		WithOptions(names).
		Show()

	settings.SelectedProfile = result

	file, _ := json.Marshal(settings)

	settingsFile := settingsPath + "settings.json"

	_ = os.WriteFile(settingsFile, file, 0644)

	fmt.Printf("Profile set to: %v\n", result)

}
