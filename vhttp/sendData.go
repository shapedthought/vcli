package vhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/shapedthought/vcli/auth"
	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"gopkg.in/yaml.v2"
)

func SendData(api_url string, filename string, endPoint string, method string, profile models.Profile, settings models.Settings) {

		// profile := utils.GetProfile()
		// settings := utils.ReadSettings()

		// var api_url string

		// if settings.CredsFileMode {
		// 	if len(profile.Address) > 0 {
		// 		api_url = profile.Address
		// 	} else {
		// 		log.Fatal("Profile Address not set")
		// 	}
		// } else {
		// 	api_url = os.Getenv("VCLI_URL")
		// 	if api_url == "" {
		// 		log.Fatal("VCLI_URL environment variable not set")
		// 	}
		// }

		var data interface{}

		client := Client(settings.ApiNotSecure)
		apibit := "/api/"

		if profile.Name == "vb365" && profile.Name != "ent_man" {
			apibit = "/"
		} else if profile.Name == "ent_man" {
			apibit = "/api"
		}
	
		connstring := fmt.Sprintf("https://%v:%v%v%v/%v", api_url, profile.Port, apibit, profile.APIVersion, endPoint)

		// var sendData []byte
		var j []byte
		var err error

		// if usi {
		// 	st, err := io.ReadAll(os.Stdin)
		// 	utils.IsErr(err)
		// 	str := string(st)
		// 	fmt.Printf("%#v", str)
		// 	str = strings.TrimSuffix(str, "\n")
		// 	j = []byte(str)
		// 	err = json.Unmarshal(j, &data)
		// 	utils.IsErr(err)
		// } else if strings.Contains(filename, ".") {
		// 	j, err = os.ReadFile(filename)
		// 	fmt.Printf("%#v", string(j))
		// 	utils.IsErr(err)
		// 	err = json.Unmarshal(j, &data)
		// 	utils.IsErr(err)
		// }
		if strings.Contains(filename, ".") {
			j, err = os.ReadFile(filename)
			utils.IsErr(err)
			if strings.Contains(filename, "yaml") || strings.Contains(filename, "yml") {
				err = yaml.Unmarshal(j, &data)
				utils.IsErr(err)
				// sendData, err = json.Marshal(data)
				// utils.IsErr(err)
			} else {
				err = json.Unmarshal(j, &data)
				utils.IsErr(err)
			}
		}

		sendData, err := json.Marshal(data)
		utils.IsErr(err)

		var r *http.Request
		// var err error

		if strings.Contains(filename, ".") {
			r, err = http.NewRequest(method, connstring, bytes.NewReader(sendData))
			utils.IsErr(err)
		} else {
			r, err = http.NewRequest(method, connstring, nil)
			utils.IsErr(err)
		}
		
		utils.IsErr(err)

		r.Header.Add("accept", profile.Headers.Accept)
		r.Header.Add("Content-Type", "application/json")

		// Get authentication token using TokenManager
		token, err := auth.GetTokenForRequest(profile, settings)
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		if profile.Name == "ent_man" {
			r.Header.Add("x-RestSvcSessionId", token)
		} else {
			r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
			r.Header.Add("Authorization", "Bearer "+token)
		}

		res, err := client.Do(r)
		if err != nil {
			fmt.Printf("Error sending HTTP request %v\n", err)
			return
		}

		defer res.Body.Close()

		fmt.Println("Status Code:", res.StatusCode)
		fmt.Println("Status:", res.Status)

}
