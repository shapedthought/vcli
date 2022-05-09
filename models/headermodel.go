package models

type SendHeader struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Issued       string `json:".issued"`
	Expires      string `json:".expires"`
}
