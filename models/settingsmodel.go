package models

type Settings struct {
	SelectedProfile string `json:"selectedProfile"`
	ApiNotSecure    bool   `json:"apiNotSecure"`
	CredsFileMode 	bool   `json:"credsFileMode"`
}
