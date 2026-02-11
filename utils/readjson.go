package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/shapedthought/owlctl/models"
	"gopkg.in/yaml.v2"
)

func ReadCurrentProfile() {
	settings := ReadSettings()
	profile := GetProfile(settings.SelectedProfile)

	d, err := json.MarshalIndent(profile, "", "    ")
	IsErr(err)
	fmt.Println(string(d))
}

// GetProfile returns the profile with the given name
func GetProfile(profileName string) models.Profile {
	profilesFile := ReadProfilesFile()

	profile, exists := profilesFile.Profiles[profileName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Profile '%s' not found in profiles.json\n", profileName)
		os.Exit(1)
	}

	// Apply instance port override if set
	if profilePortOverride != nil {
		profile.Port = *profilePortOverride
	}

	return profile
}

// GetCurrentProfile returns the currently selected profile from settings
func GetCurrentProfile() models.Profile {
	settings := ReadSettings()
	return GetProfile(settings.SelectedProfile)
}

func SettingPath() string {
	settingsPath := os.Getenv("OWLCTL_SETTINGS_PATH")
	if settingsPath != "" {
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(settingsPath, "\\") {
				settingsPath = settingsPath + "\\"
			}
		} else {
			if !strings.HasSuffix(settingsPath, "/") {
				settingsPath = settingsPath + "/"
			}
		}
	}
	return settingsPath
}

// ReadProfilesFile reads the profiles.json file (new v1.0 format)
func ReadProfilesFile() models.ProfilesFile {
	var profilesFile models.ProfilesFile

	settingsPath := SettingPath()
	profileFile := settingsPath + "profiles.json"

	j, err := os.Open(profileFile)
	IsErr(err)
	defer j.Close()

	b, err := io.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &profilesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid profiles.json format. Expected v1.0 format.\n")
		fmt.Fprintf(os.Stderr, "Please run 'owlctl init' to regenerate profiles.json\n")
		os.Exit(1)
	}

	// Validate version
	if profilesFile.Version != "1.0" {
		fmt.Fprintf(os.Stderr, "WARNING: Unsupported profiles.json version: %s (expected 1.0)\n", profilesFile.Version)
	}

	return profilesFile
}

// SaveProfilesFile writes the profiles file to disk
func SaveProfilesFile(profilesFile models.ProfilesFile) error {
	settingsPath := SettingPath()
	profileFile := settingsPath + "profiles.json"

	data, err := json.MarshalIndent(profilesFile, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	// Use restrictive permissions (0600) to protect infrastructure information
	if err := os.WriteFile(profileFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write profiles file: %w", err)
	}

	return nil
}

// settingsOverride allows ActivateInstance to change the effective profile/insecure
// without touching settings.json on disk.
var settingsOverride *models.Settings

// OverrideSettings sets a process-level settings override.
// Subsequent calls to ReadSettings() return this value instead of reading from disk.
func OverrideSettings(s models.Settings) { settingsOverride = &s }

// ClearSettingsOverride removes the settings override, reverting to disk reads.
func ClearSettingsOverride() { settingsOverride = nil }

// profilePortOverride allows ActivateInstance to override the product profile's default port.
var profilePortOverride *int

// OverrideProfilePort sets a process-level port override applied by GetProfile.
func OverrideProfilePort(port int) { profilePortOverride = &port }

// ClearProfilePortOverride removes the port override, reverting to the product default.
func ClearProfilePortOverride() { profilePortOverride = nil }

func ReadSettings() models.Settings {
	if settingsOverride != nil {
		return *settingsOverride
	}

	var settings models.Settings

	// get the settings path if there
	settingsPath := SettingPath()

	settingsFile := settingsPath + "settings.json"

	j, err := os.Open(settingsFile)
	IsErr(err)

	b, err := io.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &settings)
	IsErr(err)

	return settings
}

func ReadHeader[T models.SendHeader | models.BasicAuthModel]() T {
	var headers T

	settingsPath := SettingPath()

	headersFile := settingsPath + "headers.json"

	j, err := os.Open(headersFile)
	IsErr(err)

	b, err := io.ReadAll(j)
	IsErr(err)

	err = json.Unmarshal(b, &headers)
	IsErr(err)

	return headers
}

func ReadCreds() models.CredSpec {
	var creds models.CredSpec
	yml, err := os.Open("creds.yaml")
	IsErr(err)

	b, err := io.ReadAll(yml)
	IsErr(err)

	err = yaml.Unmarshal(b, &creds)
	IsErr(err)

	return creds
}
