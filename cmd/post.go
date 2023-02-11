/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var filename string


// postCmd represents the post command
var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Sends a POST command to API",
	Long: `Sends a POST commands to the selected profile.

Payload needs to be in the JSON format.

Note that vcli does not type check the payload.

Commands:
vcli post jobs/c69eb538-5a07-4bd7-80cb-bdf5142eadd6/start
vcli post jobs -f job.json

	`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()
		settings := utils.ReadSettings()

		env_url := os.Getenv("VCLI_URL")
		if env_url == "" {
			log.Fatal("VCLI_URL environment variable not set")
		}
		var data interface{}

		client := vhttp.Client(settings.ApiNotSecure)
		apibit := "/api/"

		if profile.Name == "vb365" && profile.Name != "ent_man" {
			apibit = "/"
		} else if profile.Name == "ent_man" {
			apibit = "/api"
		}
	
		connstring := fmt.Sprintf("https://%v:%v%v%v/%v", env_url, profile.Port, apibit, profile.APIVersion, args[0])

		// var sendData []byte

		if strings.Contains(filename, ".") {
			j, err := os.ReadFile(filename)
			utils.IsErr(err)
			if strings.Contains(filename, "yaml") || strings.Contains(filename, "yml") {
				err = yaml.Unmarshal(j, &data)
				utils.IsErr(err)
				// sendData, err = json.Marshal(data)
				// utils.IsErr(err)
			} else {
				fmt.Println("JSON ran")
				err = json.Unmarshal(j, &data)
				utils.IsErr(err)
			}
		}

		sendData, err := json.Marshal(data)
		utils.IsErr(err)

		var r *http.Request
		// var err error

		if strings.Contains(filename, ".") {
			r, err = http.NewRequest("POST", connstring, bytes.NewReader(sendData))
			utils.IsErr(err)
		} else {
			r, err = http.NewRequest("POST", connstring, nil)
			utils.IsErr(err)
		}
		
		utils.IsErr(err)

		r.Header.Add("accept", profile.Headers.Accept)
		if profile.Name == "ent_man" {
			headers := utils.ReadHeader[models.BasicAuthModel]()
			r.Header.Add("x-RestSvcSessionId", headers.Token)
			r.Header.Add("Content-Type", "application/json")
		} else {
			headers := utils.ReadHeader[models.SendHeader]()
			r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
			r.Header.Add("Content-Type", "application/json")
			r.Header.Add("Authorization", "Bearer "+headers.AccessToken)
		}

		res, err := client.Do(r)
		if err != nil {
			fmt.Printf("Error sending HTTP request %v\n", err)
			return
		}

		defer res.Body.Close()

		fmt.Println("Status Code:", res.StatusCode)
		fmt.Println("Status:", res.Status)
		body, err := io.ReadAll(res.Body)
		utils.IsErr(err)
		fmt.Printf("Response: %s\n", body)

	},
}

func init() {
	postCmd.Flags().StringVarP(&filename, "file", "f", "", "payload in json or yaml format")
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
