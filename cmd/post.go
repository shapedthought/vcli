/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/shapedthought/veeamcli/utils"
	"github.com/spf13/cobra"
)

var filename string

// postCmd represents the post command
var postCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a resource",
	Long: `Creates a resource in the selected profile using a supplied yaml file.

The yaml file must conform to the specifications layed out in the Veeam API docuementation.
Use a json to yaml converter, or "get" an existing job and modify the file. 

VBR Commands:
veeamcli create job -f job.yaml
	`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name == "vbr" {
			switch args[0] {
			case "job":
				vbrCreateJob(profile, filename)
			}
		}
	},
}

func init() {
	postCmd.Flags().StringVarP(&filename, "file", "f", "", "payload in yaml format")
	postCmd.MarkFlagRequired("filename")
	rootCmd.AddCommand(postCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// postCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// postCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
