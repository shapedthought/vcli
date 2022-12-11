/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com
*/
package cmd

import (
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vhttp"
	"github.com/spf13/cobra"
)

var (
	yamlF  bool
	jsonF  bool
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets data from the API",
	Long: `Sends a GET request to a specified endpoint.

Returns an the data in json format.

veeamcli get <command> - use the end of the API request after the version e.g. /v4/ 
It will print to stdout in json by default, --yaml will print out to yaml.

`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()
		
		customGet(profile, args[0])
	},
}

func init() {
	getCmd.Flags().BoolVarP(&yamlF, "yaml", "y", false, "prints output in yaml format")
	getCmd.Flags().BoolVarP(&jsonF, "json", "j", false, "prints output in json format")
	rootCmd.AddCommand(getCmd)

}

func customGet(profile models.Profile, custom string) {
	cust := vhttp.GetData[interface{}](custom, profile)

	if !yamlF {
		utils.PrintJson(&cust)
	} else {
		utils.PrintYaml(&cust)
	}

}
