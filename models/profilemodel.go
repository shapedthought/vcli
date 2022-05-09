package models

type Profile struct {
	Name       string `json:"name"`
	Header     Header `json:"header"`
	URL        string `json:"url"`
	APIVersion string `json:"api_version"`
}

type Header struct {
	Accept      string `json:"accept"`
	ContentType string `json:"Content-type"`
	XAPIVersion string `json:"x-api-version"`
}
