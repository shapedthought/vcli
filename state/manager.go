package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shapedthought/vcli/utils"
)

const stateFile = ".vcli/state.json"

// Manager handles state operations
type Manager struct {
	statePath string
	lock      *Lock
}

// NewManager creates a new state manager
func NewManager() *Manager {
	settingsPath := utils.SettingPath()
	return &Manager{
		statePath: filepath.Join(settingsPath, stateFile),
		lock:      NewLock(),
	}
}

// Lock acquires the state lock
func (m *Manager) Lock() error {
	return m.lock.Acquire()
}

// Unlock releases the state lock
func (m *Manager) Unlock() error {
	return m.lock.Release()
}

// LoadState reads the state file or initializes a new one if it doesn't exist
func (m *Manager) LoadState() (*State, error) {
	// Check if state file exists
	if _, err := os.Stat(m.statePath); os.IsNotExist(err) {
		// State file doesn't exist, return new empty state
		return NewState(), nil
	}

	// Read state file
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse state
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// SaveState writes the state to disk atomically
func (m *Manager) SaveState(state *State) error {
	// Ensure .vcli directory exists
	dir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first (atomic write)
	tmpFile := m.statePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Rename temporary file to actual state file (atomic operation)
	if err := os.Rename(tmpFile, m.statePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to save state file: %w", err)
	}

	return nil
}

// GetResource retrieves a resource by name from the current state
func (m *Manager) GetResource(name string) (*Resource, error) {
	state, err := m.LoadState()
	if err != nil {
		return nil, err
	}

	resource, found := state.GetResource(name)
	if !found {
		return nil, fmt.Errorf("resource %s not found in state", name)
	}

	return resource, nil
}

// UpsertResource adds or updates a resource and saves the state
func (m *Manager) UpsertResource(resource Resource) error {
	state, err := m.LoadState()
	if err != nil {
		return err
	}

	state.UpsertResource(resource)

	return m.SaveState(state)
}

// DeleteResource removes a resource and saves the state
func (m *Manager) DeleteResource(name string) error {
	state, err := m.LoadState()
	if err != nil {
		return err
	}

	if !state.DeleteResource(name) {
		return fmt.Errorf("resource %s not found in state", name)
	}

	return m.SaveState(state)
}
