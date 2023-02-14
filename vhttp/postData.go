package vhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
)

func PostData[T any](url string, profile models.Profile, data T) bool {

	headers := utils.ReadHeader[models.SendHeader]()
	creds := utils.ReadCreds()
	settings := utils.ReadSettings()

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
	r.Header.Add("content-type", "application/json")
	r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
	r.Header.Add("Authorization", "Bearer "+headers.AccessToken)

	res, err := client.Do(r)

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	var js map[string]interface{}

	err = json.NewDecoder(res.Body).Decode(&js)
	utils.IsErr(err)

	if res.StatusCode != 201 && res.StatusCode != 200 {
		fmt.Println(js)
		fmt.Printf("Error code: %v\n", res.StatusCode)
		return false
	} else {
		return true
	}

}
