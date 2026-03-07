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
// Uses OWLCTL_SETTINGS_PATH if set, otherwise ~/.owlctl/
func NewManager() *Manager {
	var statePath string

	settingsPath := os.Getenv("OWLCTL_SETTINGS_PATH")
	if settingsPath != "" {
		statePath = filepath.Join(settingsPath, "state.json")
	} else {
		usr, err := user.Current()
		if err != nil {
			// Fallback to current directory if we can't get home dir
			statePath = "state.json"
		} else {
			owlctlDir := filepath.Join(usr.HomeDir, ".owlctl")
			// Create .owlctl directory if it doesn't exist
			os.MkdirAll(owlctlDir, 0755)
			statePath = filepath.Join(owlctlDir, "state.json")
		}
	}

	return &Manager{
		statePath: statePath,
	}
}

// activeInstance returns the currently active instance name.
// Reads OWLCTL_ACTIVE_INSTANCE env var; falls back to "default".
func activeInstance() string {
	if inst := os.Getenv("OWLCTL_ACTIVE_INSTANCE"); inst != "" {
		return inst
	}
	return "default"
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

	// Initialize instances map if nil (defensive)
	if state.Instances == nil {
		state.Instances = make(map[string]*InstanceState)
	}

	// Migrate old state versions forward
	if state.Version < CurrentStateVersion {
		migrateState(&state)
	}

	return &state, nil
}

// migrateState upgrades an older state to CurrentStateVersion in-memory.
// The migrated state is persisted on the next Save() call.
func migrateState(s *State) {
	// v1 → v2: populate Origin field (operates on legacy Resources field)
	if s.Version < 2 {
		for _, r := range s.Resources {
			if r.Origin == "" {
				if r.Type == "VBRJob" {
					r.Origin = "applied"
				} else {
					r.Origin = "observed"
				}
			}
		}
	}

	// v2 → v3: History field added (no migration needed - omitempty handles it)

	// v3 → v4: move flat Resources into Instances["default"]
	if s.Version < 4 && s.Resources != nil {
		if s.Instances == nil {
			s.Instances = make(map[string]*InstanceState)
		}
		s.Instances["default"] = &InstanceState{Resources: s.Resources}
		s.Resources = nil
	}

	s.Version = CurrentStateVersion
}

// Save writes the state to disk atomically using temp file + rename
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

	// Write atomically: write to temp file, then rename
	tmpFile, err := os.CreateTemp(dir, "state.json.tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp state file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure temp file is cleaned up on error
	defer func() {
		// Best-effort cleanup of temp file if it still exists
		if _, statErr := os.Stat(tmpPath); statErr == nil {
			_ = os.Remove(tmpPath)
		}
	}()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to sync temp state file: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp state file: %w", err)
	}

	// Atomically replace the old state file with the new one
	if err := os.Rename(tmpPath, m.statePath); err != nil {
		return fmt.Errorf("failed to rename temp state file: %w", err)
	}

	return nil
}

// GetStatePath returns the path to the state file
func (m *Manager) GetStatePath() string {
	return m.statePath
}

// UpdateResource loads state, updates a resource under the active instance, and saves
func (m *Manager) UpdateResource(resource *Resource) error {
	state, err := m.Load()
	if err != nil {
		return err
	}

	state.SetResource(activeInstance(), resource)

	return m.Save(state)
}

// RemoveResource loads state, removes a resource from the active instance, and saves
func (m *Manager) RemoveResource(name string) error {
	state, err := m.Load()
	if err != nil {
		return err
	}

	state.DeleteResource(activeInstance(), name)

	return m.Save(state)
}

// GetResource loads state and retrieves a single resource from the active instance
func (m *Manager) GetResource(name string) (*Resource, error) {
	state, err := m.Load()
	if err != nil {
		return nil, err
	}

	resource, exists := state.GetResource(activeInstance(), name)
	if !exists {
		return nil, fmt.Errorf("resource '%s' not found in state", name)
	}

	return resource, nil
}

// ListResources loads state and lists resources of the given type from the active instance
func (m *Manager) ListResources(resourceType string) ([]*Resource, error) {
	state, err := m.Load()
	if err != nil {
		return nil, err
	}

	return state.ListResources(activeInstance(), resourceType), nil
}

// stateExists checks if the state file exists
func (m *Manager) StateExists() bool {
	_, err := os.Stat(m.statePath)
	return err == nil
}
