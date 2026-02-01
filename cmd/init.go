package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize vcli",
	Long: `Initializing the vcli creates several files
	
	settings.json - a file with base settings
	profiles.json - a file with api profiles
	`,
	Run: func(cmd *cobra.Command, args []string) {
		initApp()
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
		APIVersion: "v7",
		Username: "",
		Address: "",
	}

	aws := models.Profile{
		Name: "aws",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.4-rev0",
		},
		URL:        ":11005/api/v1/token",
		Port:       "11005",
		APIVersion: "v1",
		Username: "",
		Address: "",
	}

	vbr := models.Profile{
		Name: "vbr",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.3-rev1",
		},
		URL:        ":9419/api/oauth2/token",
		Port:       "9419",
		APIVersion: "v1",
		Username: "",
		Address: "",
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
		APIVersion: "v5",
		Username: "",
		Address: "",
	}

	gcp := models.Profile{
		Name: "gcp",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.2-rev0",
		},
		URL:        ":13140/api/v1/token",
		Port:       "13140",
		APIVersion: "v1",
		Username: "",
		Address: "",
	}

	vone := models.Profile{
		Name: "vone",
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.0-rev2",
		},
		URL:        ":1239/api/token",
		Port:       "1239",
		APIVersion: "v2.1",
		Username: "",
		Address: "",
	}

	ent_man := models.Profile{
		Name: "ent_man",
		Headers: models.Headers {
			Accept: "application/json",
			ContentType: "application/json",
			XAPIVersion: "",
		},
		URL: ":9398/api/sessionMngr/?v=latest",
		Port: "9398",
		APIVersion: "",
		Username: "",
		Address: "",
	}

	ps := [7]models.Profile{vbm365, aws, vbr, azure, gcp, vone, ent_man}

	settingsPath := utils.SettingPath()

	profilePath := settingsPath + "profiles"

	utils.SaveJson(&ps, profilePath)

	pterm.DefaultInteractiveConfirm.DefaultText = "Allow insecure TLS?"

	insecure, _ := pterm.DefaultInteractiveConfirm.Show()

	pterm.DefaultInteractiveConfirm.DefaultText = "Use Creds file mode?"

	credsMode, _ := pterm.DefaultInteractiveConfirm.Show()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecure,
		CredsFileMode: credsMode,
	}

	if credsMode {
		fmt.Println("Remember to update the profile.json file with the usernames and API addresses.")
		fmt.Printf("Settings file location: %v", settingsPath)
		fmt.Println("VCLI_PASSWORD will still need to be set as an environmental variable")
	} else {
		fmt.Println("Initialized, ensure all environment variables are set.")
	}

	settingsFilePath := settingsPath + "settings"

	utils.SaveJson(&settings, settingsFilePath)

}
