package vhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
)

func ApiLogout() {
	settings := utils.ReadSettings()
	profiles := utils.ReadProfiles()

	var profile models.Profile

	check := false
	for _, v := range profiles {
		if v.Name == settings.SelectedProfile {
			profile = v
			check = true
		}
	}

	if !check {
		log.Fatalf("Error with selected profile %v", settings.SelectedProfile)
	}

	var url string

	switch settings.SelectedProfile {
	case "vbr":
		url = "oauth2/logout"
	case "vone":
		url = "revoke"
	case "vb365":
		fmt.Println("logout is automatic")
		return
	default:
		fmt.Println("profile not implemented")
		return
	}

	fmt.Println("Are you sure? Y/n")
	var confirm string
	fmt.Scanln(&confirm)

	lg := false
	var d struct{}

	if confirm == "Y" {
		lg = PostData(url, profile, d)
	} else {
		fmt.Println("Logout Cancelled")
	}

	if lg {
		fmt.Println("Logout successful")
	} else {
		fmt.Println("Logout could not be completed")
	}

}

func CheckEnv(name string, data string) {
	if data == "" {
		log.Fatalf("Cannot find environment variable: %v", name)
	}
}

func ApiLogin() {
	settings := utils.ReadSettings()
	profiles := utils.ReadProfiles()

	var profile models.Profile

	check := false
	for _, v := range profiles {
		if v.Name == settings.SelectedProfile {
			profile = v
			check = true
		}
	}

	if !check {
		log.Fatalf("Error with selected profile %v", settings.SelectedProfile)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	// fmt.Printf("username %s, password %s\n", creds.Username, creds.Password)
	username := os.Getenv("VCLI_USERNAME")
	password := os.Getenv("VCLI_PASSWORD")
	vcliUrl := os.Getenv("VCLI_URL")
	CheckEnv("VCLI_USERNAME", username)
	CheckEnv("VCLI_PASSWORD", password)
	CheckEnv("VCLI_URL", vcliUrl)

	data := url.Values{}
	data.Add("grant_type", "password")
	data.Add("username", username)
	data.Add("password", password)

	connstring := fmt.Sprintf("https://%s%s", vcliUrl, profile.URL)
	fmt.Println(connstring)
	var r *http.Request
	var err error

	if profile.Name == "ent_man" {
		r, err = http.NewRequest("POST", connstring, nil)
		utils.IsErr(err)
		r.Header.Add("accept", profile.Headers.Accept)
		r.SetBasicAuth(username, password)
	} else {
		r, err = http.NewRequest("POST", connstring, strings.NewReader(data.Encode()))
		utils.IsErr(err)
		r.Header.Add("accept", profile.Headers.Accept)
		r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
		r.Header.Add("Content-Type", profile.Headers.ContentType)
	}

	res, err := client.Do(r)
	if res.StatusCode == 401 {
		log.Fatalf("Not Authorized: %v", res.StatusCode)
	}
	utils.IsErr(err)

	if res.StatusCode != 200 && res.StatusCode != 201 {
		log.Fatalf("Error %v", res.StatusCode)
	}

	defer res.Body.Close()

	var writeData []byte

	if profile.Name == "ent_man" {
		token := res.Header.Get("X-RestSvcSessionId")

		aum := models.BasicAuthModel {
			Token: token,
			ContentType: "application/json",
		}
		tokenBytes := new(bytes.Buffer)
		json.NewEncoder(tokenBytes).Encode(aum)
		writeData = tokenBytes.Bytes()
	} else {
		writeData, err = io.ReadAll(res.Body)
		utils.IsErr(err)
	}

	settingsPath := utils.SettingPath()

	headersFile := settingsPath + "headers.json"

	if err := os.WriteFile(headersFile, writeData, 0644); err != nil {
		log.Fatalf("Could not save headers file - %v", err)
	}

	fmt.Println("Login OK")
}
