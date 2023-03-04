package utils

import (
	"log"
	"os"

	"github.com/shapedthought/vcli/models"
)

func GetAddress(profile models.Profile, settings models.Settings) string {
	var api_url string

	if settings.CredsFileMode {
		if len(profile.Address) > 0 {
			api_url = profile.Address
		} else {
			log.Fatal("Profile Address not set")
		}
	} else {
		api_url = os.Getenv("VCLI_URL")
		if api_url == "" {
			log.Fatal("VCLI_URL environment variable not set")
		}
	}

	return api_url
}