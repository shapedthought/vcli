package vhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
)

type ReadHeaders interface {
	models.BasicAuthModel |  models.SendHeader
}

func GetData[T any](url string, profile models.Profile) T {
	
	// creds := utils.ReadCreds()
	settings := utils.ReadSettings()

	env_url := os.Getenv("VCLI_URL")
	if env_url == "" {
		log.Fatal("VCLI_URL environment variable not set")
	}

	// creates a new client instance
	client := Client(settings.ApiNotSecure)

	apibit := "/api/"

	if profile.Name == "vb365" && profile.Name != "ent_man" {
		apibit = "/"
	} else if profile.Name == "ent_man" {
		apibit = "/api"
	}

	connstring := fmt.Sprintf("https://%v:%v%v%v/%v", env_url, profile.Port, apibit, profile.APIVersion, url)

	r, err := http.NewRequest("GET", connstring, nil)
	utils.IsErr(err)
	r.Header.Add("accept", profile.Headers.Accept)
	if profile.Name == "ent_man" {
		headers := utils.ReadHeader[models.BasicAuthModel]()
		r.Header.Add("x-RestSvcSessionId", headers.Token)
	} else {
		headers := utils.ReadHeader[models.SendHeader]()
		r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
		r.Header.Add("Authorization", "Bearer "+headers.AccessToken)
	}
	

	res, err := client.Do(r)
	utils.IsErr(err)

	if res.StatusCode != 200 {
		log.Fatalf("Error %v", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	utils.IsErr(err)

	var udata T

	if err := json.Unmarshal(body, &udata); err != nil {
		log.Fatalf("Could not unmarshal - %v", err)
	}

	return udata

}
