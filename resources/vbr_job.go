package resources

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
)

// VBRJobResource implements the Resource interface for VBR backup jobs
type VBRJobResource struct {
	name     string
	id       string
	spec     VBRJobSpec
	resolver *Resolver
}

// NewVBRJobResource creates a new VBR job resource from a spec
func NewVBRJobResource(name string, spec map[string]interface{}) (*VBRJobResource, error) {
	// Convert spec map to VBRJobSpec
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec: %w", err)
	}

	var jobSpec VBRJobSpec
	if err := json.Unmarshal(specBytes, &jobSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job spec: %w", err)
	}

	return &VBRJobResource{
		name:     name,
		spec:     jobSpec,
		resolver: NewResolver(),
	}, nil
}

// Type returns the resource type
func (r *VBRJobResource) Type() string {
	return "vbr_job"
}

// Name returns the resource name
func (r *VBRJobResource) Name() string {
	return r.name
}

// ID returns the resource ID
func (r *VBRJobResource) ID() string {
	return r.id
}

// SetID sets the resource ID
func (r *VBRJobResource) SetID(id string) {
	r.id = id
}

// Spec returns the resource specification as a map
func (r *VBRJobResource) Spec() map[string]interface{} {
	specBytes, _ := json.Marshal(r.spec)
	var result map[string]interface{}
	json.Unmarshal(specBytes, &result)
	return result
}

// Validate validates the resource configuration
func (r *VBRJobResource) Validate() error {
	if r.name == "" {
		return fmt.Errorf("job name is required")
	}
	if r.spec.Type == "" {
		return fmt.Errorf("job type is required")
	}
	if len(r.spec.Objects) == 0 {
		return fmt.Errorf("at least one object is required")
	}
	if r.spec.Repository == "" {
		return fmt.Errorf("repository is required")
	}
	return nil
}

// Create creates the job in VBR
func (r *VBRJobResource) Create() error {
	if err := r.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert to VBR API format
	vbrJob, err := r.toVBRFormat()
	if err != nil {
		return fmt.Errorf("failed to convert to VBR format: %w", err)
	}

	// Create the job
	profile := utils.GetProfile()
	endpoint := "jobs"
	response := vhttp.PostData[models.VbrJobGet](endpoint, vbrJob, profile)

	// Set the ID from the response
	r.id = response.ID

	return nil
}

// Update updates the job in VBR
func (r *VBRJobResource) Update(current Resource) error {
	if r.id == "" {
		return fmt.Errorf("cannot update job without ID")
	}

	if err := r.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert to VBR API format
	vbrJob, err := r.toVBRFormat()
	if err != nil {
		return fmt.Errorf("failed to convert to VBR format: %w", err)
	}

	// Update the job
	profile := utils.GetProfile()
	endpoint := fmt.Sprintf("jobs/%s", r.id)
	vhttp.PutData(endpoint, vbrJob, profile)

	return nil
}

// Delete deletes the job from VBR
func (r *VBRJobResource) Delete() error {
	if r.id == "" {
		return fmt.Errorf("cannot delete job without ID")
	}

	profile := utils.GetProfile()
	endpoint := fmt.Sprintf("jobs/%s", r.id)
	vhttp.DeleteData(endpoint, profile)

	return nil
}

// Fetch retrieves the job from VBR by ID
func (r *VBRJobResource) Fetch(id string) (Resource, error) {
	profile := utils.GetProfile()
	endpoint := fmt.Sprintf("jobs/%s", id)
	vbrJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)

	// Convert from VBR format to our spec
	spec, err := r.fromVBRFormat(&vbrJob)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from VBR format: %w", err)
	}

	resource := &VBRJobResource{
		name:     vbrJob.Name,
		id:       vbrJob.ID,
		spec:     spec,
		resolver: r.resolver,
	}

	return resource, nil
}

// Diff compares this resource with another and returns the differences
func (r *VBRJobResource) Diff(current Resource) ([]Change, error) {
	if current == nil {
		// Resource doesn't exist, everything is new
		return []Change{{Path: "resource", Old: nil, New: r.name, Action: "add"}}, nil
	}

	changes := []Change{}

	currentSpec := current.Spec()
	desiredSpec := r.Spec()

	// Compare specs recursively
	changes = append(changes, compareSpecs("", currentSpec, desiredSpec)...)

	return changes, nil
}

// compareSpecs recursively compares two specs and returns changes
func compareSpecs(path string, current, desired interface{}) []Change {
	changes := []Change{}

	// Handle nil cases
	if current == nil && desired == nil {
		return changes
	}
	if current == nil && desired != nil {
		return []Change{{Path: path, Old: nil, New: desired, Action: "add"}}
	}
	if current != nil && desired == nil {
		return []Change{{Path: path, Old: current, New: nil, Action: "delete"}}
	}

	// Compare based on type
	currentVal := reflect.ValueOf(current)
	desiredVal := reflect.ValueOf(desired)

	if currentVal.Kind() != desiredVal.Kind() {
		return []Change{{Path: path, Old: current, New: desired, Action: "modify"}}
	}

	switch currentVal.Kind() {
	case reflect.Map:
		currentMap := current.(map[string]interface{})
		desiredMap := desired.(map[string]interface{})

		// Check for added/modified keys
		for key, desiredValue := range desiredMap {
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}
			if currentValue, exists := currentMap[key]; exists {
				changes = append(changes, compareSpecs(newPath, currentValue, desiredValue)...)
			} else {
				changes = append(changes, Change{Path: newPath, Old: nil, New: desiredValue, Action: "add"})
			}
		}

		// Check for deleted keys
		for key, currentValue := range currentMap {
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}
			if _, exists := desiredMap[key]; !exists {
				changes = append(changes, Change{Path: newPath, Old: currentValue, New: nil, Action: "delete"})
			}
		}

	case reflect.Slice:
		// For slices, just compare if they're different
		if !reflect.DeepEqual(current, desired) {
			changes = append(changes, Change{Path: path, Old: current, New: desired, Action: "modify"})
		}

	default:
		// For primitive types, compare directly
		if !reflect.DeepEqual(current, desired) {
			changes = append(changes, Change{Path: path, Old: current, New: desired, Action: "modify"})
		}
	}

	return changes
}

// toVBRFormat converts our simplified spec to VBR API format
func (r *VBRJobResource) toVBRFormat() (*models.VbrJobPost, error) {
	vbrJob := &models.VbrJobPost{
		Name:        r.name,
		Description: r.spec.Description,
		Type:        r.spec.Type,
		IsDisabled:  r.spec.IsDisabled,
	}

	// Resolve and add VMs
	var includes []models.Includes
	for _, obj := range r.spec.Objects {
		objectID, err := r.resolver.ResolveVMID(obj.Name, obj.HostName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve VM %q: %w", obj.Name, err)
		}

		includes = append(includes, models.Includes{
			Type:     obj.Type,
			Name:     obj.Name,
			HostName: obj.HostName,
			ObjectID: objectID,
		})
	}

	vbrJob.VirtualMachines.Includes = includes

	// Resolve repository
	repoID, err := r.resolver.ResolveRepositoryID(r.spec.Repository)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository: %w", err)
	}
	vbrJob.Storage.BackupRepositoryID = repoID

	// Set storage defaults
	vbrJob.Storage.BackupProxies.SetAutoSelect(true)
	vbrJob.Storage.RetentionPolicy.Type = "Days"
	vbrJob.Storage.RetentionPolicy.Quantity = 7

	// Apply storage settings if provided
	if r.spec.Storage != nil {
		if r.spec.Storage.Retention != nil {
			vbrJob.Storage.RetentionPolicy.Type = r.spec.Storage.Retention.Type
			vbrJob.Storage.RetentionPolicy.Quantity = r.spec.Storage.Retention.Quantity
		}
		if r.spec.Storage.Compression != "" {
			vbrJob.Storage.AdvancedSettings.StorageData.CompressionLevel = r.spec.Storage.Compression
		}
		vbrJob.Storage.AdvancedSettings.StorageData.Encryption.IsEnabled = r.spec.Storage.Encryption
	}

	// Set schedule if provided
	if r.spec.Schedule != nil {
		vbrJob.Schedule.RunAutomatically = r.spec.Schedule.Enabled
		if r.spec.Schedule.Daily != "" {
			vbrJob.Schedule.Daily.IsEnabled = true
			vbrJob.Schedule.Daily.LocalTime = r.spec.Schedule.Daily
			vbrJob.Schedule.Daily.DailyKind = "Everyday"
		}
		if r.spec.Schedule.Retry != nil {
			vbrJob.Schedule.Retry.IsEnabled = r.spec.Schedule.Retry.Enabled
			vbrJob.Schedule.Retry.RetryCount = r.spec.Schedule.Retry.Times
			vbrJob.Schedule.Retry.AwaitMinutes = r.spec.Schedule.Retry.Wait
		}
	} else {
		// Default: no automatic schedule
		vbrJob.Schedule.RunAutomatically = false
	}

	return vbrJob, nil
}

// fromVBRFormat converts VBR API format to our simplified spec
func (r *VBRJobResource) fromVBRFormat(vbrJob *models.VbrJobGet) (VBRJobSpec, error) {
	spec := VBRJobSpec{
		Type:        vbrJob.Type,
		Description: vbrJob.Description,
		IsDisabled:  vbrJob.IsDisabled,
	}

	// Convert VMs
	for _, include := range vbrJob.VirtualMachines.Includes {
		spec.Objects = append(spec.Objects, JobObject{
			Type:     include.InventoryObject.Type,
			Name:     include.InventoryObject.Name,
			HostName: include.InventoryObject.HostName,
		})
	}

	// Resolve repository name
	repoName, err := r.resolver.ResolveRepositoryName(vbrJob.Storage.BackupRepositoryID)
	if err != nil {
		// If we can't resolve, use the ID
		repoName = vbrJob.Storage.BackupRepositoryID
	}
	spec.Repository = repoName

	// Convert schedule
	if vbrJob.Schedule.RunAutomatically {
		spec.Schedule = &JobSchedule{
			Enabled: true,
		}
		if vbrJob.Schedule.Daily.IsEnabled {
			spec.Schedule.Daily = vbrJob.Schedule.Daily.LocalTime
		}
		if vbrJob.Schedule.Retry.IsEnabled {
			spec.Schedule.Retry = &struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
				Times   int  `yaml:"times,omitempty" json:"times,omitempty"`
				Wait    int  `yaml:"wait,omitempty" json:"wait,omitempty"`
			}{
				Enabled: true,
				Times:   vbrJob.Schedule.Retry.RetryCount,
				Wait:    vbrJob.Schedule.Retry.AwaitMinutes,
			}
		}
	}

	// Convert storage settings
	spec.Storage = &JobStorageSettings{
		Compression: vbrJob.Storage.AdvancedSettings.StorageData.CompressionLevel,
		Encryption:  vbrJob.Storage.AdvancedSettings.StorageData.Encryption.IsEnabled,
		Retention: &struct {
			Type     string `yaml:"type,omitempty" json:"type,omitempty"`
			Quantity int    `yaml:"quantity,omitempty" json:"quantity,omitempty"`
		}{
			Type:     vbrJob.Storage.RetentionPolicy.Type,
			Quantity: vbrJob.Storage.RetentionPolicy.Quantity,
		},
	}

	return spec, nil
}
