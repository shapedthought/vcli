package state

import "time"

// CurrentStateVersion is the latest state file format version.
// Increment when changing the Resource schema and add a migration in manager.go.
const CurrentStateVersion = 3

// DefaultMaxHistoryEvents is the maximum number of history events to keep per resource
const DefaultMaxHistoryEvents = 20

// State represents the vcli state file structure
//
// WARNING: State files are mutable and NOT suitable for compliance or audit.
// For compliance, use Git commit history + CI/CD logs + VBR audit logs.
// State is purely operational - used for drift detection and tracking what was last applied.
type State struct {
	Version   int                  `json:"version"`
	Resources map[string]*Resource `json:"resources"`
}

// ResourceEvent represents an action taken on a resource
type ResourceEvent struct {
	Action    string    `json:"action"`              // "snapshotted", "adopted", "applied"
	Timestamp time.Time `json:"timestamp"`           // When the action occurred
	User      string    `json:"user"`                // Who performed the action
	Fields    []string  `json:"fields,omitempty"`    // Fields that were changed (for apply)
	Partial   bool      `json:"partial,omitempty"`   // True if some fields failed
}

// Resource represents a managed resource in state
type Resource struct {
	Type          string                 `json:"type"`                    // e.g., "VBRJob"
	ID            string                 `json:"id"`                      // VBR resource ID
	Name          string                 `json:"name"`                    // Resource name
	LastApplied   time.Time              `json:"lastApplied"`             // When it was last applied
	LastAppliedBy string                 `json:"lastAppliedBy"`           // User who applied it
	Origin        string                 `json:"origin"`                  // "applied" (declarative) or "observed" (snapshot)
	Spec          map[string]interface{} `json:"spec"`                    // The applied configuration
	History       []ResourceEvent        `json:"history,omitempty"`       // Audit trail of actions
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

// AddEvent records a new event to the resource's history and prunes old events
func (r *Resource) AddEvent(event ResourceEvent) {
	// Prepend new event (most recent first)
	r.History = append([]ResourceEvent{event}, r.History...)

	// Prune to max length
	if len(r.History) > DefaultMaxHistoryEvents {
		r.History = r.History[:DefaultMaxHistoryEvents]
	}
}

// NewEvent creates a new ResourceEvent with the current timestamp and user
func NewEvent(action string, user string) ResourceEvent {
	return ResourceEvent{
		Action:    action,
		Timestamp: time.Now(),
		User:      user,
	}
}

// NewEventWithFields creates a new ResourceEvent with changed fields
func NewEventWithFields(action string, user string, fields []string, partial bool) ResourceEvent {
	return ResourceEvent{
		Action:    action,
		Timestamp: time.Now(),
		User:      user,
		Fields:    fields,
		Partial:   partial,
	}
}
