package models

// VbrRepoGet represents a single backup repository from the VBR API.
// The repository model is polymorphic (WinLocal, LinuxLocal, etc.), so
// only common fields are typed here. For drift detection, the full API
// response is compared as map[string]interface{}.
type VbrRepoGet struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
	UniqueID    string `json:"uniqueId" yaml:"uniqueId"`
	HostID      string `json:"hostId,omitempty" yaml:"hostId,omitempty"`

	// Repository settings (path, task limits, advanced settings)
	Repository map[string]interface{} `json:"repository,omitempty" yaml:"repository,omitempty"`

	// Mount server configuration
	MountServer map[string]interface{} `json:"mountServer,omitempty" yaml:"mountServer,omitempty"`
}

// VbrRepoList represents the list response from the repositories endpoint
type VbrRepoList struct {
	Data       []VbrRepoGet           `json:"data"`
	Pagination map[string]interface{} `json:"pagination,omitempty"`
}
