package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shapedthought/vcli/models"
)

// Authenticator handles API authentication
type Authenticator struct {
	client *http.Client
	debug  bool
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(insecure bool, debug bool) *Authenticator {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	return &Authenticator{
		client: &http.Client{Transport: tr},
		debug:  debug,
	}
}

// AuthResult contains authentication response
type AuthResult struct {
	Token      string
	ExpiresIn  int
	TokenType  string
	IsBasicAuth bool
}

// Authenticate performs OAuth or Basic Auth login
func (a *Authenticator) Authenticate(profile models.Profile, username, password, apiURL string) (*AuthResult, error) {
	// Build connection string using new Endpoints structure
	connstring := fmt.Sprintf("https://%s:%d%s", apiURL, profile.Port, profile.Endpoints.Auth)

	if a.debug {
		fmt.Printf("DEBUG: Authenticating to %s\n", connstring)
	}

	var r *http.Request
	var err error

	if profile.AuthType == "basic" {
		// Enterprise Manager uses Basic Auth
		r, err = http.NewRequest("POST", connstring, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		r.Header.Add("accept", profile.Headers.Accept)
		r.SetBasicAuth(username, password)
	} else {
		// Other products use OAuth
		data := url.Values{}
		data.Add("grant_type", "password")
		data.Add("username", username)
		data.Add("password", password)

		r, err = http.NewRequest("POST", connstring, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		r.Header.Add("accept", profile.Headers.Accept)
		r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
		r.Header.Add("Content-Type", profile.Headers.ContentType)
	}

	res, err := a.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		return nil, fmt.Errorf("authentication failed: invalid credentials (401)")
	}

	if res.StatusCode != 200 && res.StatusCode != 201 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("authentication failed with status %d: %s", res.StatusCode, string(body))
	}

	if profile.AuthType == "basic" {
		// Extract session token from header
		token := res.Header.Get("X-RestSvcSessionId")
		if token == "" {
			return nil, fmt.Errorf("no session token in response")
		}

		return &AuthResult{
			Token:       token,
			ExpiresIn:   3600, // Enterprise Manager doesn't provide expiry, default to 1 hour
			TokenType:   "session",
			IsBasicAuth: true,
		}, nil
	} else {
		// Parse OAuth response
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var oauthResp models.SendHeader
		if err := json.Unmarshal(body, &oauthResp); err != nil {
			return nil, fmt.Errorf("failed to parse OAuth response: %w", err)
		}

		if oauthResp.AccessToken == "" {
			return nil, fmt.Errorf("no access token in response")
		}

		return &AuthResult{
			Token:       oauthResp.AccessToken,
			ExpiresIn:   oauthResp.ExpiresIn,
			TokenType:   oauthResp.TokenType,
			IsBasicAuth: false,
		}, nil
	}
}

// ExtractTokenInfo parses token from OAuth response body (for legacy compatibility)
func ExtractTokenInfo(responseBody []byte, isBasicAuth bool) (string, int, error) {
	if isBasicAuth {
		var authModel models.BasicAuthModel
		if err := json.Unmarshal(responseBody, &authModel); err != nil {
			return "", 0, fmt.Errorf("failed to parse basic auth response: %w", err)
		}
		return authModel.Token, 3600, nil
	}

	var oauthResp models.SendHeader
	if err := json.Unmarshal(responseBody, &oauthResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse OAuth response: %w", err)
	}

	return oauthResp.AccessToken, oauthResp.ExpiresIn, nil
}

// FormatTokenForHeaders formats token for headers.json (legacy format)
func FormatTokenForHeaders(token string, expiresIn int, tokenType string, isBasicAuth bool) []byte {
	if isBasicAuth {
		aum := models.BasicAuthModel{
			Token:       token,
			ContentType: "application/json",
		}
		tokenBytes := new(bytes.Buffer)
		json.NewEncoder(tokenBytes).Encode(aum)
		return tokenBytes.Bytes()
	}

	header := models.SendHeader{
		AccessToken: token,
		TokenType:   tokenType,
		ExpiresIn:   expiresIn,
	}
	tokenBytes := new(bytes.Buffer)
	json.NewEncoder(tokenBytes).Encode(header)
	return tokenBytes.Bytes()
}
