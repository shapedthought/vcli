package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shapedthought/owlctl/models"
)

// newMockOAuthServer returns a test server simulating VBR OAuth endpoint.
// Accepts username=admin / password=pass; rejects all else with 401.
func newMockOAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		if r.FormValue("username") != "admin" || r.FormValue("password") != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.SendHeader{
			AccessToken: "mock-oauth-token-12345",
			TokenType:   "Bearer",
			ExpiresIn:   900,
		})
	}))
}

// newMockBasicAuthServer returns a test server simulating Enterprise Manager basic auth.
func newMockBasicAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("X-RestSvcSessionId", "session-token-abcdef")
		w.WriteHeader(http.StatusCreated)
	}))
}

// callOAuth performs an OAuth POST directly against a test server URL.
// This bypasses Authenticator.Authenticate's HTTPS URL construction so we
// can test against httptest (HTTP) servers.
func callOAuth(t *testing.T, serverURL, username, password string, profile models.Profile) (*AuthResult, error) {
	t.Helper()

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)

	connstring := serverURL + profile.Endpoints.Auth
	req, err := http.NewRequest(http.MethodPost, connstring, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", profile.Headers.Accept)
	req.Header.Set("Content-Type", profile.Headers.ContentType)
	req.Header.Set("x-api-version", profile.Headers.XAPIVersion)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: invalid credentials (401)")
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status %d", res.StatusCode)
	}

	var resp models.SendHeader
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}
	if resp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}
	return &AuthResult{Token: resp.AccessToken, ExpiresIn: resp.ExpiresIn, TokenType: resp.TokenType}, nil
}

// callBasicAuth performs a Basic Auth POST directly against a test server URL.
func callBasicAuth(t *testing.T, serverURL, username, password string, profile models.Profile) (*AuthResult, error) {
	t.Helper()

	connstring := serverURL + profile.Endpoints.Auth
	req, err := http.NewRequest(http.MethodPost, connstring, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", profile.Headers.Accept)
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: invalid credentials (401)")
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("authentication failed with status %d", res.StatusCode)
	}

	token := res.Header.Get("X-RestSvcSessionId")
	if token == "" {
		return nil, fmt.Errorf("no session token in response")
	}
	return &AuthResult{Token: token, ExpiresIn: 3600, TokenType: "session", IsBasicAuth: true}, nil
}

// ---- OAuth Authentication flow -------------------------------------------------
// Note: these tests exercise the OAuth/BasicAuth HTTP flows via callOAuth/callBasicAuth
// helper functions backed by mock servers. They do not call Authenticator.Authenticate
// directly (which builds https:// URLs incompatible with httptest servers).
// Coverage of Authenticator.Authenticate itself requires integration tests (#77).

func TestOAuthFlow_Success(t *testing.T) {
	srv := newMockOAuthServer(t)
	defer srv.Close()

	profile := models.Profile{
		AuthType: "oauth",
		Endpoints: models.Endpoints{Auth: ""},
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
			XAPIVersion: "1.1-rev0",
		},
	}

	result, err := callOAuth(t, srv.URL, "admin", "pass", profile)
	if err != nil {
		t.Fatalf("OAuth authentication returned error: %v", err)
	}
	if result.Token != "mock-oauth-token-12345" {
		t.Errorf("token = %q, want %q", result.Token, "mock-oauth-token-12345")
	}
	if result.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", result.ExpiresIn)
	}
	if result.IsBasicAuth {
		t.Error("expected IsBasicAuth=false for OAuth")
	}
}

func TestOAuthFlow_InvalidCredentials(t *testing.T) {
	srv := newMockOAuthServer(t)
	defer srv.Close()

	profile := models.Profile{
		AuthType:  "oauth",
		Endpoints: models.Endpoints{Auth: ""},
		Headers: models.Headers{
			Accept:      "application/json",
			ContentType: "application/x-www-form-urlencoded",
		},
	}

	_, err := callOAuth(t, srv.URL, "wrong", "creds", profile)
	if err == nil {
		t.Error("expected error for invalid credentials, got nil")
	}
}

// ---- Basic Auth (Enterprise Manager) flow -------------------------------------

func TestBasicAuthFlow_Success(t *testing.T) {
	srv := newMockBasicAuthServer(t)
	defer srv.Close()

	profile := models.Profile{
		AuthType:  "basic",
		Endpoints: models.Endpoints{Auth: ""},
		Headers:   models.Headers{Accept: "application/json"},
	}

	result, err := callBasicAuth(t, srv.URL, "admin", "pass", profile)
	if err != nil {
		t.Fatalf("Basic auth returned error: %v", err)
	}
	if result.Token != "session-token-abcdef" {
		t.Errorf("token = %q, want %q", result.Token, "session-token-abcdef")
	}
	if !result.IsBasicAuth {
		t.Error("expected IsBasicAuth=true for Enterprise Manager")
	}
}

func TestBasicAuthFlow_InvalidCredentials(t *testing.T) {
	srv := newMockBasicAuthServer(t)
	defer srv.Close()

	profile := models.Profile{
		AuthType:  "basic",
		Endpoints: models.Endpoints{Auth: ""},
		Headers:   models.Headers{Accept: "application/json"},
	}

	_, err := callBasicAuth(t, srv.URL, "wrong", "creds", profile)
	if err == nil {
		t.Error("expected error for invalid credentials, got nil")
	}
}

// ---- ExtractTokenInfo ----------------------------------------------------------

func TestExtractTokenInfo_OAuth(t *testing.T) {
	body, _ := json.Marshal(models.SendHeader{
		AccessToken: "extracted-token",
		ExpiresIn:   600,
		TokenType:   "Bearer",
	})

	token, expiresIn, err := ExtractTokenInfo(body, false)
	if err != nil {
		t.Fatalf("ExtractTokenInfo returned error: %v", err)
	}
	if token != "extracted-token" {
		t.Errorf("token = %q, want %q", token, "extracted-token")
	}
	if expiresIn != 600 {
		t.Errorf("expiresIn = %d, want 600", expiresIn)
	}
}

func TestExtractTokenInfo_BasicAuth(t *testing.T) {
	body, _ := json.Marshal(models.BasicAuthModel{
		Token:       "session-token-xyz",
		ContentType: "application/json",
	})

	token, expiresIn, err := ExtractTokenInfo(body, true)
	if err != nil {
		t.Fatalf("ExtractTokenInfo returned error: %v", err)
	}
	if token != "session-token-xyz" {
		t.Errorf("token = %q, want %q", token, "session-token-xyz")
	}
	if expiresIn != 3600 {
		t.Errorf("expiresIn = %d, want 3600 (default for basic auth)", expiresIn)
	}
}

func TestExtractTokenInfo_InvalidJSON(t *testing.T) {
	_, _, err := ExtractTokenInfo([]byte("not json {{{"), false)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ---- FormatTokenForHeaders -----------------------------------------------------

func TestFormatTokenForHeaders_OAuth(t *testing.T) {
	data := FormatTokenForHeaders("bearer-token", 900, "Bearer", false)

	var h models.SendHeader
	if err := json.Unmarshal(data, &h); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if h.AccessToken != "bearer-token" {
		t.Errorf("AccessToken = %q, want %q", h.AccessToken, "bearer-token")
	}
	if h.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", h.ExpiresIn)
	}
	if h.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", h.TokenType, "Bearer")
	}
}

func TestFormatTokenForHeaders_BasicAuth(t *testing.T) {
	data := FormatTokenForHeaders("session-xyz", 3600, "session", true)

	var h models.BasicAuthModel
	if err := json.Unmarshal(data, &h); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if h.Token != "session-xyz" {
		t.Errorf("Token = %q, want %q", h.Token, "session-xyz")
	}
}
