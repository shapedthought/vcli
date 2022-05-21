package vhttp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
)

func ApiLogin() {
	settings := utils.ReadSettings()
	creds := utils.ReadCreds()
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

	data := url.Values{}
	data.Add("grant_type", "password")
	data.Add("username", creds.Username)
	data.Add("password", creds.Password)

	connstring := fmt.Sprintf("https://%s%s", creds.Server, profile.URL)

	// fmt.Println(connstring)
	// fmt.Println(profile)

	r, err := http.NewRequest("POST", connstring, strings.NewReader(data.Encode()))
	r.Header.Add("accept", profile.Headers.Accept)
	r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
	r.Header.Add("Content-Type", profile.Headers.ContentType)
	utils.IsErr(err)

	res, err := client.Do(r)
	if res.StatusCode == 401 {
		log.Fatalf("Not Authorised: %v", res.StatusCode)
	}
	utils.IsErr(err)

	if res.StatusCode != 200 {
		log.Fatalf("Error %v", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	utils.IsErr(err)

	var token models.TokenModel

	if err := json.Unmarshal(body, &token); err != nil {
		log.Fatalf("Could not unmarshal - %v", err)
	}

	if err := ioutil.WriteFile("headers.json", body, 0644); err != nil {
		log.Fatalf("Could not save headers file - %v", err)
	}

	fmt.Println("Login OK")
}
