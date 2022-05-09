package models

type TokenLoginSpec struct {
	Username  string `json:"Username"`
	Password  string `json:"Password"`
	GrantType string `json:"grant_type"`
}

type TokenModel struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Issued       string `json:".issued"`
	Expires      string `json:".expires"`
	UserName     string `json:"username"`
	RoleName     string `json:"rolename"`
	MFA          bool   `json:"mfa_enabled"`
}

type CredSpec struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
}
