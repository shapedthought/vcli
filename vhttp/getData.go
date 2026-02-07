package vhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/shapedthought/owlctl/auth"
	"github.com/shapedthought/owlctl/models"
	"github.com/shapedthought/owlctl/utils"
)

type ReadHeaders interface {
	models.BasicAuthModel |  models.SendHeader
}

func GetData[T any](url string, profile models.Profile) T {
	
	settings := utils.ReadSettings()

	// With v1.0 profiles, credentials are always from environment variables
	api_url := os.Getenv("OWLCTL_URL")
	if api_url == "" {
		log.Fatal("OWLCTL_URL environment variable not set")
	}

	// creates a new client instance
	client := Client(settings.ApiNotSecure)

	// Use APIPrefix from endpoints structure
	connstring := fmt.Sprintf("https://%v:%v%v/%v", api_url, profile.Port, profile.Endpoints.APIPrefix, url)

	r, err := http.NewRequest("GET", connstring, nil)
	utils.IsErr(err)
	r.Header.Add("accept", profile.Headers.Accept)

	// Get authentication token using TokenManager
	token, err := auth.GetTokenForRequest(settings.SelectedProfile, profile, settings)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	if profile.AuthType == "basic" {
		r.Header.Add("x-RestSvcSessionId", token)
	} else {
		r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
		r.Header.Add("Authorization", "Bearer "+token)
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
