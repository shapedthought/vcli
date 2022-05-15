/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com

*/
package cmd

import (
	"fmt"

	"github.com/shapedthought/veeamcli/utils"
	"github.com/spf13/cobra"
)

var (
	save  bool
	yamlF bool
	jsonF bool
	nameF string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets data from the API",
	Long: `Sends a GET request to a specified endpoint.

Prints a table of the specified resource. 

Examples:
veeamcli get jobStates
veeamcli get jobs 
veeamcli get proxies
veeamcli get repos
veeamcli get sessions - last 20
veeamcli get backupObjects
veeamcli get inventory - gets VMware inventory
`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name == "vbr" {
			switch args[0] {
			case "jobStates":
				getVbrJobStates(profile)
			case "jobs":
				getVbrJobs(profile)
			case "proxies":
				getProxies(profile)
			case "repos":
				getRepos(profile)
			case "sobr":
				getSobr(profile)
			case "sessions":
				getSessions(profile)
			case "backupObjects":
				getBackupObject(profile)
			case "inventory":
				getInventory(profile)
			case "all":
				getAll(profile)
			default:
				fmt.Println("type not found")
			}
		}
	},
}

func init() {
	getCmd.Flags().BoolVarP(&save, "save", "s", false, "saves the data to a yaml file")
	getCmd.Flags().BoolVarP(&yamlF, "yaml", "y", false, "prints output in yaml format")
	getCmd.Flags().BoolVarP(&jsonF, "json", "j", false, "prints output in json format")
	getCmd.Flags().StringVarP(&nameF, "nameFilter", "f", "", "filters items based on supplied string value")
	rootCmd.AddCommand(getCmd)

}
