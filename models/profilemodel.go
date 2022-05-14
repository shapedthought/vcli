package models

type Profile struct {
	Name       string  `json:"name"`
	Headers    Headers `json:"headers"`
	URL        string  `json:"url"`
	Port       string  `json:"port"`
	APIVersion string  `json:"api_version"`
}

type Headers struct {
	Accept      string `json:"accept"`
	ContentType string `json:"Content-type"`
	XAPIVersion string `json:"x-api-version"`
}
