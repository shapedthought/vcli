package resources

// ResourceSpec represents the declarative YAML specification structure
type ResourceSpec struct {
	APIVersion string                 `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                 `yaml:"kind" json:"kind"`
	Metadata   Metadata               `yaml:"metadata" json:"metadata"`
	Spec       map[string]interface{} `yaml:"spec" json:"spec"`
}

// Metadata contains resource metadata
type Metadata struct {
	Name        string            `yaml:"name" json:"name"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// VBRJobSpec represents the simplified user-friendly job specification
type VBRJobSpec struct {
	Type        string              `yaml:"type" json:"type"`
	Description string              `yaml:"description,omitempty" json:"description,omitempty"`
	IsDisabled  bool                `yaml:"isDisabled,omitempty" json:"isDisabled,omitempty"`
	Objects     []JobObject         `yaml:"objects" json:"objects"`
	Repository  string              `yaml:"repository" json:"repository"`
	Schedule    *JobSchedule        `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Storage     *JobStorageSettings `yaml:"storage,omitempty" json:"storage,omitempty"`
}

// JobObject represents a VM or other object to backup
type JobObject struct {
	Type     string `yaml:"type" json:"type"`         // "VM", "VMFolder", etc.
	Name     string `yaml:"name" json:"name"`         // VM name
	HostName string `yaml:"hostName,omitempty" json:"hostName,omitempty"` // vCenter hostname
}

// JobSchedule represents simplified schedule configuration
type JobSchedule struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Daily   string `yaml:"daily,omitempty" json:"daily,omitempty"` // e.g., "22:00"
	Retry   *struct {
		Enabled bool `yaml:"enabled" json:"enabled"`
		Times   int  `yaml:"times,omitempty" json:"times,omitempty"`
		Wait    int  `yaml:"wait,omitempty" json:"wait,omitempty"` // minutes
	} `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// JobStorageSettings represents simplified storage configuration
type JobStorageSettings struct {
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty"` // "Auto", "None", "Dedupe", etc.
	Encryption  bool   `yaml:"encryption,omitempty" json:"encryption,omitempty"`
	Retention   *struct {
		Type     string `yaml:"type,omitempty" json:"type,omitempty"`         // "Days", "RestorePoints"
		Quantity int    `yaml:"quantity,omitempty" json:"quantity,omitempty"` // Number of days or restore points
	} `yaml:"retention,omitempty" json:"retention,omitempty"`
}
