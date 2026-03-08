package models

type Settings struct {
	SelectedProfile string `json:"selectedProfile"`
	ApiNotSecure    bool   `json:"apiNotSecure"`
	DefaultInstance string `json:"defaultInstance,omitempty"`
	// CredsFileMode removed in v1.0 - credentials always from environment variables
}
