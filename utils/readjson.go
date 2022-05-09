package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/shapedthought/veeamcli/models"
	"gopkg.in/yaml.v2"
)

func ReadCurrentProfile() {
	settings := ReadSettings()

	profiles := ReadProfiles()

	for _, v := range profiles {
		if v.Name == settings.SelectedProfile {
			d, err := json.MarshalIndent(v, "", "    ")
			IsErr(err)
			fmt.Println(string(d))
		}
	}
}

func GetProfile() models.Profile {
	profiles := ReadProfiles()
	settings := ReadSettings()

	var profile models.Profile
	for _, v := range profiles {
		if v.Name == settings.SelectedProfile {
			profile = v
		}
	}

	return profile
}

func ReadProfiles() []models.Profile {
	var profiles []models.Profile

	j, err := os.Open("profiles.json")
	IsErr(err)

	b, err := ioutil.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &profiles)
	IsErr(err)

	return profiles
}

func ReadSettings() models.Settings {
	var settings models.Settings

	j, err := os.Open("settings.json")
	IsErr(err)

	b, err := ioutil.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &settings)
	IsErr(err)

	return settings
}

func ReadHeader() models.SendHeader {
	var headers models.SendHeader

	j, err := os.Open("headers.json")
	IsErr(err)

	b, err := ioutil.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &headers)
	IsErr(err)

	return headers
}

func ReadCreds() models.CredSpec {
	var creds models.CredSpec
	yml, err := os.Open("creds.yaml")
	IsErr(err)

	b, err := ioutil.ReadAll(yml)
	IsErr(err)

	err = yaml.Unmarshal(b, &creds)
	IsErr(err)

	return creds
}
