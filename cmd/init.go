package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/shapedthought/owlctl/models"
	"github.com/shapedthought/owlctl/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	interactive bool
	insecure    bool
	outputDir   string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize owlctl configuration files",
	Long: `Initialize owlctl by creating settings.json and profiles.json files.

By default, init runs non-interactively and outputs JSON to stdout.
Use --interactive flag for the legacy interactive prompt behavior.

Examples:
  # Non-interactive init (outputs to current directory or OWLCTL_SETTINGS_PATH)
  owlctl init

  # Output to specific directory
  owlctl init --output-dir .owlctl/

  # With configuration flags
  owlctl init --insecure --creds-file

  # Pipe to jq for custom handling
  owlctl init | jq '.settings' > .owlctl/settings.json
  owlctl init | jq '.profiles' > .owlctl/profiles.json

  # Interactive mode (legacy)
  owlctl init --interactive
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
  owlctl init settings
  owlctl init settings --insecure --creds-file
  owlctl init settings --output-dir .owlctl/
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
  owlctl init profiles
  owlctl init profiles --output-dir .owlctl/
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
		cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to write configuration files (default: current dir or OWLCTL_SETTINGS_PATH)")
	}

	// Settings-specific flags
	initCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS verification (sets apiNotSecure: true)")
	initSettingsCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS verification (sets apiNotSecure: true)")
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
	// Use OWLCTL_SETTINGS_PATH or current directory
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
// getProfileNames returns a slice of profile names from ProfilesFile
func getProfileNames(pf models.ProfilesFile) []string {
	names := make([]string, 0, len(pf.Profiles))
	for name := range pf.Profiles {
		names = append(names, name)
	}
	return names
}

func getDefaultProfiles() models.ProfilesFile {
	return models.ProfilesFile{
		Version:        "1.0",
		CurrentProfile: "vbr",
		Profiles: map[string]models.Profile{
			"vb365": {
				Product:    "VeeamBackupFor365",
				APIVersion: "v8",
				Port:       4443,
				Endpoints: models.Endpoints{
					Auth:      "/v8/Token",
					APIPrefix: "/v8",
				},
				AuthType: "oauth",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/x-www-form-urlencoded",
					XAPIVersion: "",
				},
			},
			"aws": {
				Product:    "VeeamBackupForAWS",
				APIVersion: "1.4-rev0",
				Port:       11005,
				Endpoints: models.Endpoints{
					Auth:      "/api/v1/token",
					APIPrefix: "/api/v1",
				},
				AuthType: "oauth",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/x-www-form-urlencoded",
					XAPIVersion: "1.4-rev0",
				},
			},
			"vbr": {
				Product:    "VeeamBackupReplication",
				APIVersion: "1.3-rev1",
				Port:       9419,
				Endpoints: models.Endpoints{
					Auth:      "/api/oauth2/token",
					APIPrefix: "/api/v1",
				},
				AuthType: "oauth",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/x-www-form-urlencoded",
					XAPIVersion: "1.3-rev1",
				},
			},
			"azure": {
				Product:    "VeeamBackupForAzure",
				APIVersion: "v5",
				Port:       443,
				Endpoints: models.Endpoints{
					Auth:      "/api/oauth2/token",
					APIPrefix: "/api/v5",
				},
				AuthType: "oauth",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/x-www-form-urlencoded",
					XAPIVersion: "",
				},
			},
			"gcp": {
				Product:    "VeeamBackupForGCP",
				APIVersion: "1.2-rev0",
				Port:       13140,
				Endpoints: models.Endpoints{
					Auth:      "/api/v1/token",
					APIPrefix: "/api/v1",
				},
				AuthType: "oauth",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/x-www-form-urlencoded",
					XAPIVersion: "1.2-rev0",
				},
			},
			"vone": {
				Product:    "VeeamOne",
				APIVersion: "1.0-rev2",
				Port:       1239,
				Endpoints: models.Endpoints{
					Auth:      "/api/token",
					APIPrefix: "/api/v2.1",
				},
				AuthType: "oauth",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/x-www-form-urlencoded",
					XAPIVersion: "1.0-rev2",
				},
			},
			"ent_man": {
				Product:    "EnterpriseManager",
				APIVersion: "",
				Port:       9398,
				Endpoints: models.Endpoints{
					Auth:      "/api/sessionMngr/?v=latest",
					APIPrefix: "/api",
				},
				AuthType: "basic",
				Headers: models.Headers{
					Accept:      "application/json",
					ContentType: "application/json",
					XAPIVersion: "",
				},
			},
		},
	}
}

// initAppNonInteractive runs init in non-interactive mode (default)
func initAppNonInteractive() {
	basePath := getOutputPath()

	// Create profiles (new v1.0 format)
	profilesFile := getDefaultProfiles()
	profilePath := basePath + "profiles"
	utils.SaveJson(&profilesFile, profilePath)

	// Create settings
	settings := models.Settings{
		SelectedProfile: profilesFile.CurrentProfile,
		ApiNotSecure:    insecure,
	}
	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)

	// Output result as JSON for piping
	result := map[string]interface{}{
		"version":  profilesFile.Version,
		"settings": settings,
		"profiles": profilesFile.Profiles,
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
	fmt.Fprintln(os.Stderr, "\nInitialized successfully (profiles v1.0)")
	fmt.Fprintln(os.Stderr, "Ensure environment variables are set: OWLCTL_USERNAME, OWLCTL_PASSWORD, OWLCTL_URL")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Available profiles: %v\n", getProfileNames(profilesFile))
	fmt.Fprintf(os.Stderr, "Current profile: %s\n", profilesFile.CurrentProfile)
}

// initSettingsNonInteractive initializes only settings.json
func initSettingsNonInteractive() {
	basePath := getOutputPath()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecure,
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

	profilesFile := getDefaultProfiles()
	profilePath := basePath + "profiles"
	utils.SaveJson(&profilesFile, profilePath)

	// Output as JSON
	result := map[string]interface{}{
		"version":  profilesFile.Version,
		"profiles": profilesFile.Profiles,
		"file":     profilePath + ".json",
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
	fmt.Fprintf(os.Stderr, "\nProfiles file created (v%s): %s\n", profilesFile.Version, profilePath+".json")
}

// initAppInteractive runs init in interactive mode (legacy behavior)
func initAppInteractive() {
	// Show deprecation warning if not explicitly in interactive mode
	if !interactive && isInteractiveSession() {
		fmt.Fprintln(os.Stderr, "⚠️  WARNING: Interactive mode is deprecated and will be removed in v0.12.0")
		fmt.Fprintln(os.Stderr, "   Use 'owlctl init --interactive' to explicitly enable interactive mode")
		fmt.Fprintln(os.Stderr, "   Or use non-interactive mode: 'owlctl init --insecure --creds-file'")
		fmt.Fprintln(os.Stderr, "")
	}

	basePath := getOutputPath()

	// Create profiles (new v1.0 format)
	profilesFile := getDefaultProfiles()
	profilePath := basePath + "profiles"
	utils.SaveJson(&profilesFile, profilePath)

	// Interactive prompts
	pterm.DefaultInteractiveConfirm.DefaultText = "Allow insecure TLS?"
	insecureResult, _ := pterm.DefaultInteractiveConfirm.Show()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecureResult,
	}

	fmt.Println("Initialized successfully. Ensure environment variables are set:")
	fmt.Println("  - OWLCTL_USERNAME, OWLCTL_PASSWORD, OWLCTL_URL")

	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)
}

// initSettingsInteractive initializes settings.json interactively
func initSettingsInteractive() {
	basePath := getOutputPath()

	pterm.DefaultInteractiveConfirm.DefaultText = "Allow insecure TLS?"
	insecureResult, _ := pterm.DefaultInteractiveConfirm.Show()

	settings := models.Settings{
		SelectedProfile: "vbr",
		ApiNotSecure:    insecureResult,
	}

	settingsPath := basePath + "settings"
	utils.SaveJson(&settings, settingsPath)

	fmt.Printf("Settings file created: %s.json\n", settingsPath)
}
