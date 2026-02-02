package state

import "time"

// CurrentStateVersion is the latest state file format version.
// Increment when changing the Resource schema and add a migration in manager.go.
const CurrentStateVersion = 2

// State represents the vcli state file structure
//
// WARNING: State files are mutable and NOT suitable for compliance or audit.
// For compliance, use Git commit history + CI/CD logs + VBR audit logs.
// State is purely operational - used for drift detection and tracking what was last applied.
type State struct {
	Version   int                  `json:"version"`
	Resources map[string]*Resource `json:"resources"`
}

// Resource represents a managed resource in state
type Resource struct {
	Type          string                 `json:"type"`           // e.g., "VBRJob"
	ID            string                 `json:"id"`             // VBR resource ID
	Name          string                 `json:"name"`           // Resource name
	LastApplied   time.Time              `json:"lastApplied"`    // When it was last applied
	LastAppliedBy string                 `json:"lastAppliedBy"`  // User who applied it
	Origin        string                 `json:"origin"`         // "applied" (declarative) or "observed" (snapshot)
	Spec          map[string]interface{} `json:"spec"`           // The applied configuration
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Version:   CurrentStateVersion,
		Resources: make(map[string]*Resource),
	}
}

// GetResource retrieves a resource by name
func (s *State) GetResource(name string) (*Resource, bool) {
	resource, exists := s.Resources[name]
	return resource, exists
}

// SetResource adds or updates a resource in state
func (s *State) SetResource(resource *Resource) {
	s.Resources[resource.Name] = resource
}

// DeleteResource removes a resource from state
func (s *State) DeleteResource(name string) {
	delete(s.Resources, name)
}

// ListResources returns all resources of a given type
func (s *State) ListResources(resourceType string) []*Resource {
	var resources []*Resource
	for _, resource := range s.Resources {
		if resourceType == "" || resource.Type == resourceType {
			resources = append(resources, resource)
		}
	}
	return resources
}
