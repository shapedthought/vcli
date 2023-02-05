package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

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

func SettingPath() string {
	settingsPath := os.Getenv("VCLI_SETTINGS_PATH")
	if settingsPath != "" {
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(settingsPath, "\\") {
				settingsPath = settingsPath + "\\"
			}
		} else {
			if !strings.HasSuffix(settingsPath, "/") {
				settingsPath = settingsPath + "/"
			}
		}
	}
	return settingsPath
}

func ReadProfiles() []models.Profile {
	var profiles []models.Profile

	settingsPath := SettingPath()

	profileFile := settingsPath + "profiles.json"

	j, err := os.Open(profileFile)
	IsErr(err)

	b, err := io.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &profiles)
	IsErr(err)

	return profiles
}

func ReadSettings() models.Settings {
	var settings models.Settings

	// get the settings path if there
	settingsPath := SettingPath()

	settingsFile := settingsPath + "settings.json"

	j, err := os.Open(settingsFile)
	IsErr(err)

	b, err := io.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &settings)
	IsErr(err)

	return settings
}

func ReadHeader[T models.SendHeader | models.BasicAuthModel]() T {
	var headers T

	settingsPath := SettingPath()

	headersFile := settingsPath + "headers.json"

	j, err := os.Open(headersFile)
	IsErr(err)

	b, err := io.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &headers)
	IsErr(err)

	return headers
}

func ReadCreds() models.CredSpec {
	var creds models.CredSpec
	yml, err := os.Open("creds.yaml")
	IsErr(err)

	b, err := io.ReadAll(yml)
	IsErr(err)

	err = yaml.Unmarshal(b, &creds)
	IsErr(err)

	return creds
}
