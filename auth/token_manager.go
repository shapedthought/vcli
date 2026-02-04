package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/99designs/keyring"
	"github.com/shapedthought/vcli/models"
	"golang.org/x/term"
)

const (
	KeyringService = "vcli"
	TokenEnvVar    = "VCLI_TOKEN"
)

// TokenManager handles token storage and retrieval
type TokenManager struct {
	keyring keyring.Keyring
	debug   bool
}

// TokenInfo holds token metadata
type TokenInfo struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	Profile   string    `json:"profile"`
}

// NewTokenManager creates a new token manager
func NewTokenManager(debug bool) (*TokenManager, error) {
	// Try to open system keychain
	// Use file backend as fallback for systems without keychain
	kr, err := keyring.Open(keyring.Config{
		ServiceName:      KeyringService,
		KeychainName:     "vcli",
		FileDir:          "~/.vcli",
		FilePasswordFunc: func(prompt string) (string, error) {
			return "vcli-fallback-key", nil
		},
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,      // macOS
			keyring.WinCredBackend,       // Windows
			keyring.SecretServiceBackend, // Linux
			keyring.FileBackend,          // Fallback
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return &TokenManager{
		keyring: kr,
		debug:   debug,
	}, nil
}

// GetToken retrieves token using hybrid approach
// Priority: 1) VCLI_TOKEN env var, 2) keychain, 3) auto-authenticate
// profileName is used for keychain storage/retrieval
func (tm *TokenManager) GetToken(profileName string, profile models.Profile, username, password, apiURL string, insecure bool) (string, error) {
	if tm.debug {
		fmt.Fprintln(os.Stderr, "DEBUG: Token resolution started")
	}

	// 1. Check explicit token env var (highest priority)
	if token := os.Getenv(TokenEnvVar); token != "" {
		if tm.debug {
			fmt.Fprintln(os.Stderr, "DEBUG: Using token from VCLI_TOKEN env var")
		}
		// Validate token format and expiry if possible
		if isValidTokenFormat(token) {
			return token, nil
		}
		if tm.debug {
			fmt.Fprintln(os.Stderr, "DEBUG: VCLI_TOKEN token invalid, trying next method")
		}
	}

	// 2. Check system keychain
	if token, err := tm.getTokenFromKeychain(profileName); err == nil {
		if tm.debug {
			fmt.Fprintln(os.Stderr, "DEBUG: Using token from system keychain")
		}
		return token, nil
	} else if tm.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Keychain lookup failed: %v\n", err)
	}

	// 3. Auto-authenticate if credentials available
	if username != "" && password != "" && apiURL != "" {
		if tm.debug {
			fmt.Fprintln(os.Stderr, "DEBUG: Auto-authenticating with credentials")
		}
		token, expiresIn, err := tm.AuthenticateWithSettings(profile, username, password, apiURL, insecure)
		if err != nil {
			return "", fmt.Errorf("auto-authentication failed: %w", err)
		}

		// Store in keychain if interactive session (not CI/CD)
		if isInteractiveSession() {
			if tm.debug {
				fmt.Fprintln(os.Stderr, "DEBUG: Storing token in keychain (interactive session)")
			}
			if err := tm.StoreToken(profileName, token, expiresIn); err != nil {
				// Non-fatal: just warn
				if tm.debug {
					fmt.Fprintf(os.Stderr, "DEBUG: Failed to store token in keychain: %v\n", err)
				}
			}
		} else if tm.debug {
			fmt.Fprintln(os.Stderr, "DEBUG: Skipping keychain storage (CI/CD environment)")
		}

		return token, nil
	}

	return "", errors.New("no authentication method available: set VCLI_TOKEN, store token in keychain, or provide VCLI_USERNAME/VCLI_PASSWORD/VCLI_URL")
}

// StoreToken stores a token in the system keychain
func (tm *TokenManager) StoreToken(profileName, token string, expiresIn int) error {
	info := TokenInfo{
		Token:     token,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
		Profile:   profileName,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	item := keyring.Item{
		Key:  profileName,
		Data: data,
	}

	if err := tm.keyring.Set(item); err != nil {
		return fmt.Errorf("failed to store token in keychain: %w", err)
	}

	return nil
}

// getTokenFromKeychain retrieves token from system keychain
func (tm *TokenManager) getTokenFromKeychain(profileName string) (string, error) {
	item, err := tm.keyring.Get(profileName)
	if err != nil {
		return "", fmt.Errorf("token not found in keychain: %w", err)
	}

	var info TokenInfo
	if err := json.Unmarshal(item.Data, &info); err != nil {
		return "", fmt.Errorf("failed to unmarshal token info: %w", err)
	}

	// Check if token is expired
	if time.Now().After(info.ExpiresAt) {
		if tm.debug {
			fmt.Fprintln(os.Stderr, "DEBUG: Keychain token expired")
		}
		// Remove expired token
		tm.keyring.Remove(profileName)
		return "", errors.New("token expired")
	}

	return info.Token, nil
}

// DeleteToken removes a token from keychain
func (tm *TokenManager) DeleteToken(profileName string) error {
	return tm.keyring.Remove(profileName)
}

// authenticate performs OAuth login and returns token
func (tm *TokenManager) authenticate(profile models.Profile, username, password, apiURL string) (string, int, error) {
	// Read settings to get insecure flag (would need to be passed in)
	// For now, we'll create authenticator with insecure=false
	// This will be refactored to accept settings as parameter
	auth := NewAuthenticator(false, tm.debug)

	result, err := auth.Authenticate(profile, username, password, apiURL)
	if err != nil {
		return "", 0, err
	}

	return result.Token, result.ExpiresIn, nil
}

// AuthenticateWithSettings performs OAuth login with settings context
func (tm *TokenManager) AuthenticateWithSettings(profile models.Profile, username, password, apiURL string, insecure bool) (string, int, error) {
	auth := NewAuthenticator(insecure, tm.debug)

	result, err := auth.Authenticate(profile, username, password, apiURL)
	if err != nil {
		return "", 0, err
	}

	return result.Token, result.ExpiresIn, nil
}

// isValidTokenFormat performs basic token format validation
func isValidTokenFormat(token string) bool {
	// JWT tokens start with "eyJ"
	// Basic validation without full JWT parsing
	return len(token) > 20 && (token[:3] == "eyJ" || len(token) > 50)
}

// isInteractiveSession detects if running in an interactive terminal
func isInteractiveSession() bool {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// Check for common CI/CD environment variables
	ciEnvVars := []string{
		"CI",                    // Generic CI indicator
		"GITHUB_ACTIONS",        // GitHub Actions
		"GITLAB_CI",             // GitLab CI
		"JENKINS_HOME",          // Jenkins
		"CIRCLECI",              // CircleCI
		"TRAVIS",                // Travis CI
		"TF_BUILD",              // Azure DevOps
		"CODEBUILD_BUILD_ID",    // AWS CodeBuild
		"BITBUCKET_BUILD_NUMBER", // Bitbucket Pipelines
	}

	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return false
		}
	}

	return true
}

// IsCI returns true if running in a CI/CD environment
func IsCI() bool {
	return !isInteractiveSession()
}

// GetTokenForRequest is a convenience function for API requests that handles
// credential gathering from settings/env and token resolution.
// Returns the token string or an error.
func GetTokenForRequest(profileName string, profile models.Profile, settings models.Settings) (string, error) {
	// Check if explicit token is set (highest priority)
	if token := os.Getenv(TokenEnvVar); token != "" {
		if isValidTokenFormat(token) {
			return token, nil
		}
	}

	// Get credentials from environment variables
	// Note: With v1.0 profiles, credentials are no longer stored in profiles.json
	username := os.Getenv("VCLI_USERNAME")
	apiURL := os.Getenv("VCLI_URL")
	password := os.Getenv("VCLI_PASSWORD")

	if username == "" || apiURL == "" {
		return "", errors.New("VCLI_USERNAME or VCLI_URL environment variable not set")
	}

	if password == "" {
		return "", errors.New("VCLI_PASSWORD environment variable not set")
	}

	// Create token manager and get token
	tm, err := NewTokenManager(false) // debug=false for API requests
	if err != nil {
		return "", fmt.Errorf("failed to initialize token manager: %w", err)
	}

	// Get token (tries keychain → auto-auth)
	token, err := tm.GetToken(profileName, profile, username, password, apiURL, settings.ApiNotSecure)
	if err != nil {
		return "", fmt.Errorf("failed to get authentication token: %w", err)
	}

	return token, nil
}
