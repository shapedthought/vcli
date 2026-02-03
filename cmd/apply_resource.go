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
	Action       string        // "created", "updated", "would-create", "would-update"
	NotFound     bool          // True if resource not found in update-only mode
	Changes      []FieldChange // Fields that were changed
	DryRun       bool          // True if this was a dry-run (no changes made)
	Error        error
}

// applyResource applies a resource spec to VBR using the provided config.
// It handles loading specs, fetching existing resources, merging, and state updates.
// If dryRun is true, it fetches current state (read-only) and displays what would change,
// but makes no modifications to VBR and does not update state.
func applyResource(specFile string, cfg ResourceApplyConfig, profile models.Profile, dryRun bool) ApplyResult {
	result := ApplyResult{DryRun: dryRun}

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
			result.NotFound = true
			result.Error = fmt.Errorf("resource '%s' not found in VBR (update-only mode)", spec.Metadata.Name)
			return result
		}

		// ApplyCreateOrUpdate mode: create new resource
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

		if dryRun {
			// Dry-run mode: show what would be created
			printDryRunCreate(spec.Metadata.Name, cfg.Kind, cleanedSpec)
			result.Action = "would-create"
			return result
		}

		// Create the resource
		fmt.Printf("Creating new %s: %s\n", cfg.Kind, spec.Metadata.Name)
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
		result.ResourceID = existingID

		// Parse existing resource into map
		var existingMap map[string]interface{}
		if err := json.Unmarshal(existingRaw, &existingMap); err != nil {
			result.Error = fmt.Errorf("failed to parse existing resource: %w", err)
			return result
		}

		// Deep merge spec into existing (spec values override existing)
		mergedSpec, err := resources.DeepMergeMaps(existingMap, spec.Spec)
		if err != nil {
			result.Error = fmt.Errorf("failed to merge specs: %w", err)
			return result
		}

		// Remove ignored fields
		mergedSpec = cleanSpec(mergedSpec, cfg.IgnoreFields)

		// Compute field changes for reporting
		result.Changes = computeFieldChanges(existingMap, mergedSpec, cfg.IgnoreFields)

		// Apply payload transformation if defined
		if cfg.PreparePayload != nil {
			mergedSpec, err = cfg.PreparePayload(mergedSpec, existingMap)
			if err != nil {
				result.Error = fmt.Errorf("failed to prepare payload: %w", err)
				return result
			}
		}

		if dryRun {
			// Dry-run mode: show what would change
			printDryRunUpdate(spec.Metadata.Name, cfg.Kind, result.Changes)
			result.Action = "would-update"
			return result
		}

		// Print changes being applied
		printApplyChanges(result.Changes, spec.Metadata.Name, true)

		// PUT the updated resource
		endpoint := fmt.Sprintf("%s/%s", cfg.Endpoint, existingID)
		_, err = vhttp.PutDataWithError(endpoint, mergedSpec, profile)
		if err != nil {
			result.Error = fmt.Errorf("failed to update resource: %w", err)
			return result
		}

		result.Action = "updated"
	}

	// Update state with origin: "applied" (skip in dry-run mode)
	if !dryRun {
		if err := updateResourceState(spec, result.ResourceID, cfg.Kind); err != nil {
			// Log warning but don't fail the apply
			fmt.Printf("Warning: Failed to update state: %v\n", err)
		}
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

// defaultPostCreate creates a new resource via POST and extracts the ID from the response
func defaultPostCreate(spec map[string]interface{}, profile models.Profile, endpoint string) (string, error) {
	// Use PostDataWithError to get proper error handling instead of log.Fatal
	responseBytes, err := vhttp.PostDataWithError(endpoint, spec, profile)
	if err != nil {
		return "", err
	}

	// Extract ID from response
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

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
