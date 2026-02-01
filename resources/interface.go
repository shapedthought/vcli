package resources

// Change represents a difference between two resource states
type Change struct {
	Path   string      `json:"path"`
	Old    interface{} `json:"old"`
	New    interface{} `json:"new"`
	Action string      `json:"action"` // "add", "modify", "delete"
}

// Resource defines the interface for all manageable resources
type Resource interface {
	// Type returns the resource type (e.g., "vbr_job")
	Type() string

	// Name returns the resource name
	Name() string

	// ID returns the resource ID (empty if not yet created)
	ID() string

	// SetID sets the resource ID
	SetID(id string)

	// Spec returns the resource specification as a map
	Spec() map[string]interface{}

	// Create creates the resource in VBR
	Create() error

	// Update updates the resource in VBR
	Update(current Resource) error

	// Delete deletes the resource from VBR
	Delete() error

	// Fetch retrieves the resource from VBR by ID
	Fetch(id string) (Resource, error)

	// Diff compares this resource with another and returns the differences
	Diff(current Resource) ([]Change, error)

	// Validate validates the resource configuration
	Validate() error
}
