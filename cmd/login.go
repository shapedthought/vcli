package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/shapedthought/vcli/auth"
	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
)

var (
	outputToken bool
	debugAuth   bool
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Veeam API",
	Long: `Authenticate with the Veeam API using credentials from environment variables or profiles.json.

The login command now uses secure token storage:
- Tokens are stored in your system keychain (not plaintext files)
- In CI/CD environments, tokens are generated but not stored
- Use --output-token to print the token for CI/CD workflows

Authentication Methods (in priority order):
1. VCLI_TOKEN environment variable (if set)
2. System keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
3. Auto-authenticate using VCLI_USERNAME/VCLI_PASSWORD/VCLI_URL

Examples:
  # Interactive login (stores in keychain)
  vcli login

  # Output token for CI/CD (doesn't store in keychain)
  export VCLI_TOKEN=$(vcli login --output-token)

  # Debug authentication
  vcli login --debug-auth
`,
	Run: func(cmd *cobra.Command, args []string) {
		loginWithTokenManager()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolVar(&outputToken, "output-token", false, "Output token to stdout (for CI/CD)")
	loginCmd.Flags().BoolVar(&debugAuth, "debug-auth", false, "Enable authentication debug logging")
}

func loginWithTokenManager() {
	settings := utils.ReadSettings()
	profiles := utils.ReadProfiles()

	var profile models.Profile
	check := false
	for _, v := range profiles {
		if v.Name == settings.SelectedProfile {
			profile = v
			check = true
			break
		}
	}

	if !check {
		log.Fatalf("Error with selected profile %v", settings.SelectedProfile)
	}

	// Get credentials
	var username, vcliUrl string
	if settings.CredsFileMode {
		if len(profile.Username) > 0 && len(profile.Address) > 0 {
			username = profile.Username
			vcliUrl = profile.Address
		} else {
			log.Fatal("Username or API address not set in the profile")
		}
	} else {
		username = os.Getenv("VCLI_USERNAME")
		vcliUrl = os.Getenv("VCLI_URL")
	}

	password := os.Getenv("VCLI_PASSWORD")

	// Validate credentials
	if username == "" {
		log.Fatal("VCLI_USERNAME not set")
	}
	if password == "" {
		log.Fatal("VCLI_PASSWORD not set")
	}
	if vcliUrl == "" {
		log.Fatal("VCLI_URL not set")
	}

	// Create token manager
	tm, err := auth.NewTokenManager(debugAuth)
	if err != nil {
		if debugAuth {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to initialize token manager: %v\n", err)
			fmt.Fprintln(os.Stderr, "Falling back to direct authentication")
		}
		// Fallback to direct authentication
		authenticateDirect(profile, username, password, vcliUrl, settings.ApiNotSecure)
		return
	}

	// Authenticate
	token, expiresIn, err := tm.AuthenticateWithSettings(profile, username, password, vcliUrl, settings.ApiNotSecure)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	if outputToken {
		// Just print the token for scripting
		fmt.Println(token)
		return
	}

	// Store token in keychain if interactive (not CI/CD)
	if auth.IsCI() {
		if debugAuth {
			fmt.Fprintln(os.Stderr, "DEBUG: CI/CD environment detected, skipping keychain storage")
		}
		fmt.Println("Login OK (CI/CD mode - token not stored)")
		fmt.Fprintln(os.Stderr, "Note: In CI/CD, tokens are auto-generated on each command")
		fmt.Fprintln(os.Stderr, "Use 'vcli login --output-token' to capture token for reuse")
	} else {
		// Store in keychain for interactive sessions
		if err := tm.StoreToken(profile.Name, token, expiresIn); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to store token in keychain: %v\n", err)
			fmt.Fprintln(os.Stderr, "Token will need to be re-generated on next command")
		} else {
			if debugAuth {
				fmt.Fprintf(os.Stderr, "DEBUG: Token stored in system keychain for profile '%s'\n", profile.Name)
			}
		}
		fmt.Println("Login OK")
	}

	// For backward compatibility, also write to headers.json (will be removed in future version)
	if !outputToken {
		if err := writeLegacyHeadersFile(token, expiresIn, profile.Name == "ent_man"); err != nil {
			// Non-fatal warning
			if debugAuth {
				fmt.Fprintf(os.Stderr, "WARNING: Failed to write legacy headers.json: %v\n", err)
			}
		} else if debugAuth {
			fmt.Fprintln(os.Stderr, "DEBUG: Legacy headers.json written for backward compatibility")
		}
	}
}

// authenticateDirect is fallback when keyring fails
func authenticateDirect(profile models.Profile, username, password, apiURL string, insecure bool) {
	authenticator := auth.NewAuthenticator(insecure, debugAuth)

	result, err := authenticator.Authenticate(profile, username, password, apiURL)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	if outputToken {
		fmt.Println(result.Token)
		return
	}

	// Write to headers.json (legacy)
	if err := writeLegacyHeadersFile(result.Token, result.ExpiresIn, result.IsBasicAuth); err != nil {
		log.Fatalf("Failed to write token: %v", err)
	}

	fmt.Println("Login OK")
}

// writeLegacyHeadersFile writes headers.json for backward compatibility
// This will be removed in a future version once all code uses TokenManager
func writeLegacyHeadersFile(token string, expiresIn int, isBasicAuth bool) error {
	settingsPath := utils.SettingPath()
	headersFile := settingsPath + "headers.json"

	var data []byte
	if isBasicAuth {
		authModel := models.BasicAuthModel{
			Token:       token,
			ContentType: "application/json",
		}
		var err error
		data, err = json.Marshal(authModel)
		if err != nil {
			return fmt.Errorf("failed to marshal basic auth: %w", err)
		}
	} else {
		header := models.SendHeader{
			AccessToken: token,
			TokenType:   "bearer",
			ExpiresIn:   expiresIn,
		}
		var err error
		data, err = json.Marshal(header)
		if err != nil {
			return fmt.Errorf("failed to marshal OAuth header: %w", err)
		}
	}

	if err := os.WriteFile(headersFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write headers file: %w", err)
	}

	return nil
}
