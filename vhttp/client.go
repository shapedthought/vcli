package vhttp

import (
	"crypto/tls"
	"net/http"
)

func Client(insecure bool) *http.Client {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	client := &http.Client{Transport: tr}

	return client
}
