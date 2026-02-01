package state

import "time"

// State represents the entire state file structure
type State struct {
	Resources    []Resource `json:"resources"`
	Version      int        `json:"version"`
	LastModified time.Time  `json:"last_modified"`
}

// Resource represents a managed resource in the state
type Resource struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	ID          string                 `json:"id"`
	Spec        map[string]interface{} `json:"spec"`
	LastApplied time.Time              `json:"last_applied"`
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Resources:    []Resource{},
		Version:      1,
		LastModified: time.Now(),
	}
}

// GetResource retrieves a resource by name
func (s *State) GetResource(name string) (*Resource, bool) {
	for i := range s.Resources {
		if s.Resources[i].Name == name {
			return &s.Resources[i], true
		}
	}
	return nil, false
}

// UpsertResource adds or updates a resource in the state
func (s *State) UpsertResource(resource Resource) {
	resource.LastApplied = time.Now()
	for i := range s.Resources {
		if s.Resources[i].Name == resource.Name {
			s.Resources[i] = resource
			s.LastModified = time.Now()
			return
		}
	}
	s.Resources = append(s.Resources, resource)
	s.LastModified = time.Now()
}

// DeleteResource removes a resource from the state by name
func (s *State) DeleteResource(name string) bool {
	for i := range s.Resources {
		if s.Resources[i].Name == name {
			s.Resources = append(s.Resources[:i], s.Resources[i+1:]...)
			s.LastModified = time.Now()
			return true
		}
	}
	return false
}
