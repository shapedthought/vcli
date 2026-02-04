package utils

import (
	"log"
	"os"

	"github.com/shapedthought/vcli/models"
)

func GetAddress(profile models.Profile, settings models.Settings) string {
	// With v1.0 profiles, credentials are always from environment variables
	api_url := os.Getenv("VCLI_URL")
	if api_url == "" {
		log.Fatal("VCLI_URL environment variable not set")
	}

	return api_url
}