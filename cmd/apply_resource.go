package cmd

import (
	"encoding/json"
	"fmt"
	"os/user"
	"time"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/resources"
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/vhttp"
)

// ApplyMode determines how apply handles missing resources
type ApplyMode int

const (
	// ApplyUpdateOnly only allows updating existing resources (PUT only).
	// Returns an error if the resource doesn't exist.
	// Used for repos, SOBRs, KMS servers which are created via VBR console.
	ApplyUpdateOnly ApplyMode = iota

	// ApplyCreateOrUpdate creates new resources or updates existing ones (POST or PUT).
	// Used for jobs which can be created via API.
	ApplyCreateOrUpdate
)

// ResourceApplyConfig defines how to apply a specific resource type
type ResourceApplyConfig struct {
	// Kind is the expected resource kind (e.g., "VBRJob", "VBRRepository")
	Kind string

	// Endpoint is the API endpoint (e.g., "jobs", "backupInfrastructure/repositories")
	Endpoint string

	// IgnoreFields are fields to exclude from the PUT payload (read-only fields)
	IgnoreFields map[string]bool

	// Mode determines whether POST is allowed for new resources
	Mode ApplyMode

	// FetchCurrent retrieves the current resource from VBR by name.
	// Returns (rawJSON, resourceID, error).
	// If not found, returns (nil, "", nil) - not an error.
	FetchCurrent func(name string, profile models.Profile) (json.RawMessage, string, error)

	// PreparePayload optionally transforms the merged spec before sending.
	// If nil, the merged spec is sent as-is.
	PreparePayload func(spec, existing map[string]interface{}) (map[string]interface{}, error)

	// PostCreate is called after creating a new resource (ApplyCreateOrUpdate mode only).
	// Returns the resource ID from the API response.
	// If nil, a default implementation extracts "id" from the response.
	PostCreate func(spec map[string]interface{}, profile models.Profile, endpoint string) (string, error)
}

// ApplyResult contains the result of an apply operation
type ApplyResult struct {
	ResourceName string
	ResourceID   string
	Action       string // "created", "updated"
	Error        error
}

// applyResource applies a resource spec to VBR using the provided config.
// It handles loading specs, fetching existing resources, merging, and state updates.
func applyResource(specFile string, cfg ResourceApplyConfig, profile models.Profile) ApplyResult {
	result := ApplyResult{}

	// Load the YAML spec
	spec, err := resources.LoadResourceSpec(specFile)
	if err != nil {
		result.Error = fmt.Errorf("failed to load spec file: %w", err)
		return result
	}

	result.ResourceName = spec.Metadata.Name

	// Validate resource kind
	if spec.Kind != cfg.Kind {
		result.Error = fmt.Errorf("invalid resource kind: expected %s, got %s", cfg.Kind, spec.Kind)
		return result
	}

	// Fetch existing resource by name
	existingRaw, existingID, err := cfg.FetchCurrent(spec.Metadata.Name, profile)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch current resource: %w", err)
		return result
	}

	resourceExists := existingRaw != nil && existingID != ""

	if !resourceExists {
		// Resource doesn't exist
		if cfg.Mode == ApplyUpdateOnly {
			// Update-only mode: error on missing resource
			result.Error = fmt.Errorf("resource '%s' not found in VBR (update-only mode)", spec.Metadata.Name)
			return result
		}

		// ApplyCreateOrUpdate mode: create new resource
		fmt.Printf("Creating new %s: %s\n", cfg.Kind, spec.Metadata.Name)

		// Remove ignored fields from spec
		cleanedSpec := cleanSpec(spec.Spec, cfg.IgnoreFields)

		// Apply payload transformation if defined
		if cfg.PreparePayload != nil {
			cleanedSpec, err = cfg.PreparePayload(cleanedSpec, nil)
			if err != nil {
				result.Error = fmt.Errorf("failed to prepare payload: %w", err)
				return result
			}
		}

		// Create the resource
		var newID string
		if cfg.PostCreate != nil {
			newID, err = cfg.PostCreate(cleanedSpec, profile, cfg.Endpoint)
		} else {
			newID, err = defaultPostCreate(cleanedSpec, profile, cfg.Endpoint)
		}
		if err != nil {
			result.Error = fmt.Errorf("failed to create resource: %w", err)
			return result
		}

		result.ResourceID = newID
		result.Action = "created"
		fmt.Printf("Created %s with ID: %s\n", cfg.Kind, newID)

	} else {
		// Resource exists: update it
		fmt.Printf("Updating %s: %s (ID: %s)\n", cfg.Kind, spec.Metadata.Name, existingID)
		result.ResourceID = existingID

		// Parse existing resource into map
		var existingMap map[string]interface{}
		if err := json.Unmarshal(existingRaw, &existingMap); err != nil {
			result.Error = fmt.Errorf("failed to parse existing resource: %w", err)
			return result
		}

		// Merge spec into existing (spec values override existing)
		mergedSpec := mergeSpecs(existingMap, spec.Spec)

		// Remove ignored fields
		mergedSpec = cleanSpec(mergedSpec, cfg.IgnoreFields)

		// Apply payload transformation if defined
		if cfg.PreparePayload != nil {
			mergedSpec, err = cfg.PreparePayload(mergedSpec, existingMap)
			if err != nil {
				result.Error = fmt.Errorf("failed to prepare payload: %w", err)
				return result
			}
		}

		// PUT the updated resource
		endpoint := fmt.Sprintf("%s/%s", cfg.Endpoint, existingID)
		_, err = vhttp.PutDataWithError(endpoint, mergedSpec, profile)
		if err != nil {
			result.Error = fmt.Errorf("failed to update resource: %w", err)
			return result
		}

		result.Action = "updated"
		fmt.Printf("Updated %s: %s\n", cfg.Kind, spec.Metadata.Name)
	}

	// Update state with origin: "applied"
	if err := updateResourceState(spec, result.ResourceID, cfg.Kind); err != nil {
		// Log warning but don't fail the apply
		fmt.Printf("Warning: Failed to update state: %v\n", err)
	}

	return result
}

// cleanSpec removes ignored fields from a spec
func cleanSpec(spec map[string]interface{}, ignoreFields map[string]bool) map[string]interface{} {
	cleaned := make(map[string]interface{})
	for k, v := range spec {
		if ignoreFields[k] {
			continue
		}
		cleaned[k] = v
	}
	return cleaned
}

// mergeSpecs merges the new spec into the existing spec.
// Values from newSpec override values in existing.
// This is a shallow merge at the top level.
func mergeSpecs(existing, newSpec map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy existing values
	for k, v := range existing {
		merged[k] = v
	}

	// Override with new spec values
	for k, v := range newSpec {
		merged[k] = v
	}

	return merged
}

// defaultPostCreate creates a new resource via POST and extracts the ID from the response
func defaultPostCreate(spec map[string]interface{}, profile models.Profile, endpoint string) (string, error) {
	// Use the typed PostData which returns the response
	type IDResponse struct {
		ID string `json:"id"`
	}

	response := vhttp.PostData[IDResponse](endpoint, spec, profile)
	if response.ID == "" {
		return "", fmt.Errorf("no ID in response")
	}
	return response.ID, nil
}

// updateResourceState saves the applied configuration to state
func updateResourceState(spec resources.ResourceSpec, resourceID, resourceType string) error {
	stateMgr := state.NewManager()

	// Get current user
	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	// Create state resource
	resource := &state.Resource{
		Type:          resourceType,
		ID:            resourceID,
		Name:          spec.Metadata.Name,
		LastApplied:   time.Now(),
		LastAppliedBy: currentUser,
		Origin:        "applied",
		Spec:          spec.Spec,
	}

	// Update state
	if err := stateMgr.UpdateResource(resource); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	fmt.Printf("State updated: %s\n", stateMgr.GetStatePath())
	return nil
}
