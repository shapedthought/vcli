package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	interactive bool
	insecure    bool
	credsFile   bool
	outputDir   string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize vcli configuration files",
	Long: `Initialize vcli by creating settings.json and profiles.json files.

By default, init runs non-interactively and outputs JSON to stdout.
Use --interactive flag for the legacy interactive prompt behavior.

Examples:
  # Non-interactive init (outputs to current directory or VCLI_SETTINGS_PATH)
  vcli init

  # Output to specific directory
  vcli init --output-dir .vcli/

  # With configuration flags
  vcli init --insecure --creds-file

  # Pipe to jq for custom handling
  vcli init | jq '.settings' > .vcli/settings.json
  vcli init | jq '.profiles' > .vcli/profiles.json

  # Interactive mode (legacy)
  vcli init --interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		if interactive {
			initAppInteractive()
		} else {
			initAppNonInteractive()
		}
	},
}

// initSettingsCmd initializes only settings.json
var initSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Initialize only settings.json",
	Long: `Initialize only the settings.json file with configuration options.

Examples:
  vcli init settings
  vcli init settings --insecure --creds-file
  vcli init settings --output-dir .vcli/
`,
	Run: func(cmd *cobra.Command, args []string) {
		if interactive {
			initSettingsInteractive()
		} else {
			initSettingsNonInteractive()
		}
	},
}

// initProfilesCmd initializes only profiles.json
var initProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Initialize only profiles.json",
	Long: `Initialize only the profiles.json file with all product profiles.

Examples:
  vcli init profiles
  vcli init profiles --output-dir .vcli/
`,
	Run: func(cmd *cobra.Command, args []string) {
		initProfilesOnly()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.AddCommand(initSettingsCmd)
	initCmd.AddCommand(initProfilesCmd)

	// Add flags to all init commands
	for _, cmd := range []*cobra.Command{initCmd, initSettingsCmd, initProfilesCmd} {
		cmd.Flags().BoolVar(&interactive, "interactive", false, "Run in interactive mode (legacy behavior)")
		cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to write configuration files (default: current dir or VCLI_SETTINGS_PATH)")
	}

	// Settings-specific flags
	initCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS verification (sets skipTLSVerify: true)")
	initCmd.Flags().BoolVar(&credsFile, "creds-file", false, "Enable credentials file mode")
	initSettingsCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS verification (sets skipTLSVerify: true)")
	initSettingsCmd.Flags().BoolVar(&credsFile, "creds-file", false, "Enable credentials file mode")
}

// isInteractiveSession detects if we're in an interactive terminal
func isInteractiveSession() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// getOutputPath returns the directory where config files should be written
func getOutputPath() string {
	if outputDir != "" {
		// Use explicit --output-dir flag
		return ensureTrailingSlash(outputDir)
	}
	// Use VCLI_SETTINGS_PATH or current directory
	return utils.SettingPath()
}

// ensureTrailingSlash adds trailing slash if missing
func ensureTrailingSlash(path string) string {
	if path == "" {
		return path
	}
	lastChar := path[len(path)-1]
	if lastChar != '/' && lastChar != '\\' {
		if os.PathSeparator == '\\' {
			return path + "\\"
		}
		return path + "/"
	}
	return path
}

// getDefaultProfiles returns the default profile configurations
func getDefaultProfiles() []models.Profile {
	return []models.Profile{
		{
			Name: "vb365",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/x-www-form-urlencoded",
				XAPIVersion: "",
			},
			URL:        ":4443/v6/Token",
			Port:       "4443",
			APIVersion: "v7",
			Username:   "",
			Address:    "",
		},
		{
			Name: "aws",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/x-www-form-urlencoded",
				XAPIVersion: "1.4-rev0",
			},
			URL:        ":11005/api/v1/token",
			Port:       "11005",
			APIVersion: "v1",
			Username:   "",
			Address:    "",
		},
		{
			Name: "vbr",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/x-www-form-urlencoded",
				XAPIVersion: "1.3-rev1",
			},
			URL:        ":9419/api/oauth2/token",
			Port:       "9419",
			APIVersion: "v1",
			Username:   "",
			Address:    "",
		},
		{
			Name: "azure",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/x-www-form-urlencoded",
				XAPIVersion: "",
			},
			URL:        "/api/oauth2/token",
			Port:       "",
			APIVersion: "v5",
			Username:   "",
			Address:    "",
		},
		{
			Name: "gcp",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/x-www-form-urlencoded",
				XAPIVersion: "1.2-rev0",
			},
			URL:        ":13140/api/v1/token",
			Port:       "13140",
			APIVersion: "v1",
			Username:   "",
			Address:    "",
		},
		{
			Name: "vone",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/x-www-form-urlencoded",
				XAPIVersion: "1.0-rev2",
			},
			URL:        ":1239/api/token",
			Port:       "1239",
			APIVersion: "v2.1",
			Username:   "",
			Address:    "",
		},
		{
			Name: "ent_man",
			Headers: models.Headers{
				Accept:      "application/json",
				ContentType: "application/json",
				XAPIVersion: "",
			},
			URL:        ":9398/api/sessionMngr/?v=latest",
			Port:       "9398",
			APIVersion: "",
			Username:   "",
			Address:    "",
		},
	}
}

// initAppNonInteractive runs init in non-interactive mode (default)
func initAppNonInteractive() {
	basePath := getOutputPath()

	// Create profiles
	profiles := getDefaultProfiles()
	profilePath := basePath + "profiles"
	utils.SaveJson(&profiles, profilePath)

	// Create settings
	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecure,
		CredsFileMode:   credsFile,
	}
	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)

	// Output result as JSON for piping
	result := map[string]interface{}{
		"settings": settings,
		"profiles": profiles,
		"files": map[string]string{
			"settings": settingsPath + ".json",
			"profiles": profilePath + ".json",
		},
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))

	// Print helpful message to stderr so it doesn't interfere with JSON piping
	if credsFile {
		fmt.Fprintln(os.Stderr, "\nNote: Credentials file mode enabled.")
		fmt.Fprintf(os.Stderr, "Update profiles.json with usernames and API addresses at: %s\n", profilePath+".json")
		fmt.Fprintln(os.Stderr, "Set VCLI_PASSWORD environment variable for authentication.")
	} else {
		fmt.Fprintln(os.Stderr, "\nInitialized successfully.")
		fmt.Fprintln(os.Stderr, "Ensure environment variables are set: VCLI_USERNAME, VCLI_PASSWORD, VCLI_URL")
	}
}

// initSettingsNonInteractive initializes only settings.json
func initSettingsNonInteractive() {
	basePath := getOutputPath()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecure,
		CredsFileMode:   credsFile,
	}
	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)

	// Output as JSON
	result := map[string]interface{}{
		"settings": settings,
		"file":     settingsPath + ".json",
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
	fmt.Fprintf(os.Stderr, "\nSettings file created: %s\n", settingsPath+".json")
}

// initProfilesOnly initializes only profiles.json
func initProfilesOnly() {
	basePath := getOutputPath()

	profiles := getDefaultProfiles()
	profilePath := basePath + "profiles"
	utils.SaveJson(&profiles, profilePath)

	// Output as JSON
	result := map[string]interface{}{
		"profiles": profiles,
		"file":     profilePath + ".json",
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
	fmt.Fprintf(os.Stderr, "\nProfiles file created: %s\n", profilePath+".json")
}

// initAppInteractive runs init in interactive mode (legacy behavior)
func initAppInteractive() {
	// Show deprecation warning if not explicitly in interactive mode
	if !interactive && isInteractiveSession() {
		fmt.Fprintln(os.Stderr, "⚠️  WARNING: Interactive mode is deprecated and will be removed in v0.12.0")
		fmt.Fprintln(os.Stderr, "   Use 'vcli init --interactive' to explicitly enable interactive mode")
		fmt.Fprintln(os.Stderr, "   Or use non-interactive mode: 'vcli init --insecure --creds-file'")
		fmt.Fprintln(os.Stderr, "")
	}

	basePath := getOutputPath()

	// Create profiles
	profiles := getDefaultProfiles()
	profilePath := basePath + "profiles"
	utils.SaveJson(&profiles, profilePath)

	// Interactive prompts
	pterm.DefaultInteractiveConfirm.DefaultText = "Allow insecure TLS?"
	insecureResult, _ := pterm.DefaultInteractiveConfirm.Show()

	pterm.DefaultInteractiveConfirm.DefaultText = "Use Creds file mode?"
	credsMode, _ := pterm.DefaultInteractiveConfirm.Show()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecureResult,
		CredsFileMode:   credsMode,
	}

	if credsMode {
		fmt.Println("Remember to update the profiles.json file with the usernames and API addresses.")
		fmt.Printf("Settings file location: %v\n", basePath)
		fmt.Println("VCLI_PASSWORD will still need to be set as an environment variable")
	} else {
		fmt.Println("Initialized, ensure all environment variables are set.")
	}

	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)
}

// initSettingsInteractive initializes settings.json interactively
func initSettingsInteractive() {
	basePath := getOutputPath()

	pterm.DefaultInteractiveConfirm.DefaultText = "Allow insecure TLS?"
	insecureResult, _ := pterm.DefaultInteractiveConfirm.Show()

	pterm.DefaultInteractiveConfirm.DefaultText = "Use Creds file mode?"
	credsMode, _ := pterm.DefaultInteractiveConfirm.Show()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecureResult,
		CredsFileMode:   credsMode,
	}

	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)

	fmt.Printf("Settings file created: %s.json\n", settingsPath)
}
