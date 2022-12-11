package vhttp

import (
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
	case "vmb365":
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

	// fmt.Println(connstring)
	// fmt.Println(profile)

	r, err := http.NewRequest("POST", connstring, strings.NewReader(data.Encode()))
	r.Header.Add("accept", profile.Headers.Accept)
	r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
	r.Header.Add("Content-Type", profile.Headers.ContentType)
	utils.IsErr(err)

	res, err := client.Do(r)
	if res.StatusCode == 401 {
		log.Fatalf("Not Authorized: %v", res.StatusCode)
	}
	utils.IsErr(err)

	if res.StatusCode != 200 {
		log.Fatalf("Error %v", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	utils.IsErr(err)

	var token models.TokenModel

	if err := json.Unmarshal(body, &token); err != nil {
		log.Fatalf("Could not unmarshal - %v", err)
	}

	if err := os.WriteFile("headers.json", body, 0644); err != nil {
		log.Fatalf("Could not save headers file - %v", err)
	}

	fmt.Println("Login OK")
}
