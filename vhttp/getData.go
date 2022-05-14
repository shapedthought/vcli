package vhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
)

func GetData[T any](url string, profile models.Profile) T {
	headers := utils.ReadHeader()
	creds := utils.ReadCreds()
	settings := utils.ReadSettings()

	if utils.CheckTime(headers.Expires) {
		log.Fatal("API key has expired, please login again")
	}

	// creates a new client instance
	client := Client(settings.ApiNotSecure)

	// example https://192.168.0.123:9194/api/v1/jobs
	connstring := fmt.Sprintf("https://%v:%v/api/%v/%v", creds.Server, profile.Port, profile.APIVersion, url)

	// fmt.Printf("connstring: %v\n", connstring)

	r, err := http.NewRequest("GET", connstring, nil)
	utils.IsErr(err)
	r.Header.Add("accept", profile.Headers.Accept)
	r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
	r.Header.Add("Authorization", "Bearer "+headers.AccessToken)

	res, err := client.Do(r)
	utils.IsErr(err)

	if res.StatusCode != 200 {
		log.Fatalf("Error %v", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	utils.IsErr(err)

	var udata T

	if err := json.Unmarshal(body, &udata); err != nil {
		log.Fatalf("Could not unmarshal - %v", err)
	}

	utils.UpdateTime()

	return udata

}
