/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/pterm/pterm"
	"github.com/shapedthought/veeamcli/utils"
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// profileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// profileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func printCurrentProfile() {
	settings := utils.ReadSettings()

	fmt.Printf("Current profile: %v\n", settings.SelectedProfile)
}

func listProfiles() {
	profiles := utils.ReadProfiles()

	var names []string

	for _, i := range profiles {
		names = append(names, i.Name)
	}

	sort.Strings(names)

	fmt.Println("Profiles available")
	for _, n := range names {
		fmt.Println(n)
	}

}

func setProfile() {
	settings := utils.ReadSettings()

	profiles := utils.ReadProfiles()

	var names []string

	for _, i := range profiles {
		names = append(names, i.Name)
	}

	sort.Strings(names)

	pterm.DefaultInteractiveSelect.DefaultText = "Select Profile"

	result, _ := pterm.DefaultInteractiveSelect.
		WithOptions(names).
		Show()

	settings.SelectedProfile = result

	file, _ := json.Marshal(settings)

	_ = os.WriteFile("settings.json", file, 0644)

	fmt.Printf("Profile set to: %v\n", result)

}
