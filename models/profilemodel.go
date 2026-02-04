package models

// ProfilesFile represents the new profiles.json structure (v1.0)
type ProfilesFile struct {
	Version        string             `json:"version"`
	CurrentProfile string             `json:"currentProfile"`
	Profiles       map[string]Profile `json:"profiles"`
}

// Profile represents a single product profile configuration
type Profile struct {
	Product    string    `json:"product"`
	APIVersion string    `json:"apiVersion"`
	Port       int       `json:"port"`
	Endpoints  Endpoints `json:"endpoints"`
	AuthType   string    `json:"authType"` // "oauth" or "basic"
	Headers    Headers   `json:"headers"`
}

// Endpoints defines API endpoint paths
type Endpoints struct {
	Auth      string `json:"auth"`      // Authentication endpoint (e.g., "/api/oauth2/token")
	APIPrefix string `json:"apiPrefix"` // API prefix (e.g., "/api/v1")
}

// Headers defines HTTP headers for API requests
type Headers struct {
	Accept      string `json:"accept"`
	ContentType string `json:"Content-type"`
	XAPIVersion string `json:"x-api-version"`
}
