package vhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
)

func PostData[T any](url string, profile models.Profile, data T) bool {
	headers := utils.ReadHeader()
	creds := utils.ReadCreds()
	settings := utils.ReadSettings()

	if utils.CheckTime(headers.Expires, profile.Name) {
		log.Fatal("API key has expired, please login again")
	}

	// creates a new client instance
	client := Client(settings.ApiNotSecure)

	// example https://192.168.0.123:9194/api/v1/jobs
	connstring := fmt.Sprintf("https://%v:%v/api/%v/%v", creds.Server, profile.Port, profile.APIVersion, url)

	// fmt.Printf("connstring: %v\n", connstring)
	b, err := json.Marshal(data)
	utils.IsErr(err)

	r, err := http.NewRequest("POST", connstring, bytes.NewReader(b))
	utils.IsErr(err)
	r.Header.Add("accept", profile.Headers.Accept)
	r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
	r.Header.Add("Authorization", "Bearer "+headers.AccessToken)

	res, err := client.Do(r)
	utils.IsErr(err)

	if res.StatusCode != 201 {
		log.Fatalf("Error %v", res.StatusCode)
		log.Fatal(res.Status)
	} else {
		return true
	}

	return false

}
