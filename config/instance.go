package config

import (
	"fmt"
	"os"

	"github.com/shapedthought/owlctl/utils"
)

// ResolvedInstance holds the fully-resolved connection parameters for an instance.
type ResolvedInstance struct {
	Name        string
	Product     string
	URL         string
	Username    string
	Password    string
	Port        int
	Insecure    *bool
	KeychainKey string // "instance:<name>" — used as keychain storage key
}

// ResolveInstance reads an instance from the config and resolves credentials from env vars.
// If credentialRef is set, reads OWLCTL_{ref}_USERNAME / OWLCTL_{ref}_PASSWORD.
// Otherwise falls back to OWLCTL_USERNAME / OWLCTL_PASSWORD.
func ResolveInstance(cfg *VCLIConfig, name string) (*ResolvedInstance, error) {
	inst, err := cfg.GetInstance(name)
	if err != nil {
		return nil, err
	}

	resolved := &ResolvedInstance{
		Name:        name,
		Product:     inst.Product,
		URL:         inst.URL,
		Port:        inst.Port,
		Insecure:    inst.Insecure,
		KeychainKey: "instance:" + name,
	}

	// Resolve credentials
	if inst.CredentialRef != "" {
		usernameVar := fmt.Sprintf("OWLCTL_%s_USERNAME", inst.CredentialRef)
		passwordVar := fmt.Sprintf("OWLCTL_%s_PASSWORD", inst.CredentialRef)
		resolved.Username = os.Getenv(usernameVar)
		resolved.Password = os.Getenv(passwordVar)
		if resolved.Username == "" {
			return nil, fmt.Errorf("instance %q has credentialRef=%q but %s is not set", name, inst.CredentialRef, usernameVar)
		}
		if resolved.Password == "" {
			return nil, fmt.Errorf("instance %q has credentialRef=%q but %s is not set", name, inst.CredentialRef, passwordVar)
		}
	} else {
		resolved.Username = os.Getenv("OWLCTL_USERNAME")
		resolved.Password = os.Getenv("OWLCTL_PASSWORD")
	}

	return resolved, nil
}

// ActivateInstance sets process-global state so that existing vhttp/auth code
// picks up this instance's connection parameters without any call-site changes.
//
// It does four things:
//  1. Sets OWLCTL_URL, OWLCTL_USERNAME, OWLCTL_PASSWORD env vars
//  2. Sets OWLCTL_KEYCHAIN_KEY so token_manager stores/retrieves per-instance tokens
//  3. Overrides utils.ReadSettings() to return the instance's product as SelectedProfile
//     and the instance's insecure flag as ApiNotSecure
//  4. Returns nil on success
func ActivateInstance(resolved *ResolvedInstance) error {
	// 1. Set connection env vars
	if err := os.Setenv("OWLCTL_URL", resolved.URL); err != nil {
		return fmt.Errorf("failed to set OWLCTL_URL: %w", err)
	}
	if resolved.Username != "" {
		if err := os.Setenv("OWLCTL_USERNAME", resolved.Username); err != nil {
			return fmt.Errorf("failed to set OWLCTL_USERNAME: %w", err)
		}
	}
	if resolved.Password != "" {
		if err := os.Setenv("OWLCTL_PASSWORD", resolved.Password); err != nil {
			return fmt.Errorf("failed to set OWLCTL_PASSWORD: %w", err)
		}
	}

	// 2. Set keychain key for per-instance token storage
	if err := os.Setenv("OWLCTL_KEYCHAIN_KEY", resolved.KeychainKey); err != nil {
		return fmt.Errorf("failed to set OWLCTL_KEYCHAIN_KEY: %w", err)
	}

	// 3. Override product port if the instance specifies a non-default port
	if resolved.Port != 0 {
		utils.OverrideProfilePort(resolved.Port)
	}

	// 4. Override settings so ReadSettings() returns the instance's product + insecure
	// Read current settings first to preserve fields not overridden by the instance
	currentSettings := utils.ReadSettings()
	currentSettings.SelectedProfile = resolved.Product
	if resolved.Insecure != nil {
		currentSettings.ApiNotSecure = *resolved.Insecure
	}
	utils.OverrideSettings(currentSettings)

	return nil
}
