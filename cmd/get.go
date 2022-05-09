/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com

*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vhttp"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets data from the API",
	Long: `Sends a GET request to a specified endpoint.

Prints a table of the specified resource. 

Examples:
# gets all the configured jobs
veeamcli get jobs`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()
		if args[0] == "jobs" {
			if profile.Name == "vbr" {
				getVbrJobs(args[0], profile)
			}
		}
	},
}

func init() {

	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getVbrJobs(url string, profile models.Profile) {
	getData := vhttp.GetData[models.VbrJob](url, profile)

	d, err := json.MarshalIndent(getData, "", "    ")
	utils.IsErr(err)

	fmt.Println(d)

}
