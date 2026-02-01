package state

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Manager handles state file operations
type Manager struct {
	statePath string
}

// NewManager creates a new state manager
// Uses VCLI_SETTINGS_PATH if set, otherwise ~/.vcli/
func NewManager() *Manager {
	var statePath string

	settingsPath := os.Getenv("VCLI_SETTINGS_PATH")
	if settingsPath != "" {
		statePath = filepath.Join(settingsPath, "state.json")
	} else {
		usr, err := user.Current()
		if err != nil {
			// Fallback to current directory if we can't get home dir
			statePath = "state.json"
		} else {
			vcliDir := filepath.Join(usr.HomeDir, ".vcli")
			// Create .vcli directory if it doesn't exist
			os.MkdirAll(vcliDir, 0755)
			statePath = filepath.Join(vcliDir, "state.json")
		}
	}

	return &Manager{
		statePath: statePath,
	}
}

// Load reads the state file from disk
// Returns a new empty state if file doesn't exist
func (m *Manager) Load() (*State, error) {
	// Check if file exists
	if _, err := os.Stat(m.statePath); os.IsNotExist(err) {
		// Return new empty state if file doesn't exist
		return NewState(), nil
	}

	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Initialize resources map if nil (shouldn't happen, but defensive)
	if state.Resources == nil {
		state.Resources = make(map[string]*Resource)
	}

	return &state, nil
}

// Save writes the state to disk
func (m *Manager) Save(state *State) error {
	// Ensure directory exists
	dir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// GetStatePath returns the path to the state file
func (m *Manager) GetStatePath() string {
	return m.statePath
}

// UpdateResource is a convenience method that loads state, updates a resource, and saves
func (m *Manager) UpdateResource(resource *Resource) error {
	state, err := m.Load()
	if err != nil {
		return err
	}

	state.SetResource(resource)

	return m.Save(state)
}

// RemoveResource is a convenience method that loads state, removes a resource, and saves
func (m *Manager) RemoveResource(name string) error {
	state, err := m.Load()
	if err != nil {
		return err
	}

	state.DeleteResource(name)

	return m.Save(state)
}

// GetResource is a convenience method to load and retrieve a single resource
func (m *Manager) GetResource(name string) (*Resource, error) {
	state, err := m.Load()
	if err != nil {
		return nil, err
	}

	resource, exists := state.GetResource(name)
	if !exists {
		return nil, fmt.Errorf("resource '%s' not found in state", name)
	}

	return resource, nil
}

// ListResources is a convenience method to load and list resources
func (m *Manager) ListResources(resourceType string) ([]*Resource, error) {
	state, err := m.Load()
	if err != nil {
		return nil, err
	}

	return state.ListResources(resourceType), nil
}

// stateExists checks if the state file exists
func (m *Manager) StateExists() bool {
	_, err := os.Stat(m.statePath)
	return err == nil
}
