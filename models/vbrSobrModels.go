package models

// VbrSobrGet represents a scale-out backup repository from the VBR API.
// Only common fields are typed; the full response is compared as map[string]interface{}
// for drift detection since tier configurations vary.
type VbrSobrGet struct {
	ID              string                 `json:"id" yaml:"id"`
	Name            string                 `json:"name" yaml:"name"`
	Description     string                 `json:"description" yaml:"description"`
	UniqueID        string                 `json:"uniqueId" yaml:"uniqueId"`
	PerformanceTier map[string]interface{} `json:"performanceTier,omitempty" yaml:"performanceTier,omitempty"`
	CapacityTier    map[string]interface{} `json:"capacityTier,omitempty" yaml:"capacityTier,omitempty"`
	ArchiveTier     map[string]interface{} `json:"archiveTier,omitempty" yaml:"archiveTier,omitempty"`
	PlacementPolicy map[string]interface{} `json:"placementPolicy,omitempty" yaml:"placementPolicy,omitempty"`
}

// VbrSobrList represents the list response from the scale-out repositories endpoint
type VbrSobrList struct {
	Data       []VbrSobrGet           `json:"data"`
	Pagination map[string]interface{} `json:"pagination,omitempty"`
}
