/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initalise veeamcli",
	Long: `Initalising the veeamcli creates several files
	
	settings.json - a file with base settings
	profiles.json - a file with api profiles
	creds.yaml - a file with the credentials of the api
	`,
	Run: func(cmd *cobra.Command, args []string) {
		initApp()
		fmt.Println("Initialized, ensure all environment variables are set.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initApp() {

	vbm365 := models.Profile{
		Name: "vb365",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "",
		},
		URL:        ":4443/v6/Token",
		Port:       "4443",
		APIVersion: "v6",
	}

	aws := models.Profile{
		Name: "aws",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.2-rev0",
		},
		URL:        ":11005/api/v1/token",
		Port:       "11005",
		APIVersion: "v1",
	}

	vbr := models.Profile{
		Name: "vbr",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.0-rev2",
		},
		URL:        ":9419/api/oauth2/token",
		Port:       "9419",
		APIVersion: "v1",
	}

	azure := models.Profile{
		Name: "azure",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "",
		},
		URL:        "/api/oauth2/token",
		Port:       "",
		APIVersion: "v3",
	}

	gcp := models.Profile{
		Name: "gcp",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.0-rev0",
		},
		URL:        ":13140/api/v1/token",
		Port:       "13140",
		APIVersion: "v1",
	}

	vone := models.Profile{
		Name: "vone",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.0-rev2",
		},
		URL:        ":1239/api/token",
		Port:       ":1239",
		APIVersion: "v2",
	}

	ps := [6]models.Profile{vbm365, aws, vbr, azure, gcp, vone}

	settingsPath := utils.SettingPath()

	profilePath := settingsPath + "profiles"

	utils.SaveJson(&ps, profilePath)

	pterm.DefaultInteractiveConfirm.DefaultText = "Allow insecure TLS?"

	result, _ := pterm.DefaultInteractiveConfirm.Show()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    result,
	}

	settingsFilePath := settingsPath + "settings"

	utils.SaveJson(&settings, settingsFilePath)

}
