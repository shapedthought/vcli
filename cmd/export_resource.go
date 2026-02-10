package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/shapedthought/owlctl/models"
	"github.com/shapedthought/owlctl/resources"
	"gopkg.in/yaml.v3"
)

// ResourceExportConfig defines how to export a specific resource type
type ResourceExportConfig struct {
	// Kind is the resource kind (e.g., "VBRRepository")
	Kind string

	// DisplayName is the human-readable singular name (e.g., "repository", "SOBR")
	DisplayName string

	// PluralName is the human-readable plural name (e.g., "repositories", "SOBRs")
	PluralName string

	// IgnoreFields are fields to strip from the exported spec (id, uniqueId, etc.)
	IgnoreFields map[string]bool

	// FetchSingle retrieves a single resource by name (used for single-resource export).
	// Returns (rawJSON, resourceID, error). If not found, returns (nil, "", nil).
	FetchSingle func(name string, profile models.Profile) (json.RawMessage, string, error)

	// FetchByID retrieves a single resource by ID (used for bulk export).
	// If nil, bulk export falls back to FetchSingle(item.Name).
	// Returns (rawJSON, error). If not found, returns (nil, nil).
	FetchByID func(id string, profile models.Profile) (json.RawMessage, error)

	// ListAll lists all resources of this type.
	ListAll func(profile models.Profile) ([]ResourceListItem, error)

	// SanitizeSpec optionally strips sensitive fields from the spec before export.
	// If nil, no sanitization is performed.
	SanitizeSpec func(spec map[string]interface{})

	// SupportsOverlay indicates whether --as-overlay is allowed for this resource type.
	SupportsOverlay bool
}

// ResourceListItem represents a single item in a resource listing
type ResourceListItem struct {
	ID   string
	Name string
}

// exportSingleResource fetches one resource and writes it as YAML
func exportSingleResource(name string, cfg ResourceExportConfig, profile models.Profile, output string, asOverlay bool, basePath string) {
	if asOverlay && !cfg.SupportsOverlay {
		log.Fatalf("--as-overlay is not supported for %s exports", cfg.DisplayName)
	}

	rawData, id, err := cfg.FetchSingle(name, profile)
	if err != nil {
		log.Fatalf("Failed to fetch %s: %v", cfg.DisplayName, err)
	}
	if rawData == nil {
		log.Fatalf("%s '%s' not found in VBR.", cfg.DisplayName, name)
	}

	yamlContent, err := convertResourceToYAML(name, id, cfg, rawData, asOverlay, basePath)
	if err != nil {
		log.Fatalf("Failed to convert %s to YAML: %v", cfg.DisplayName, err)
	}

	if output != "" {
		if err := os.WriteFile(output, yamlContent, 0644); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
		fmt.Printf("Exported %s to %s\n", cfg.DisplayName, output)
	} else {
		fmt.Println(string(yamlContent))
	}
}

// exportAllResources lists all resources of a type and writes each as a YAML file
func exportAllResources(cfg ResourceExportConfig, profile models.Profile, directory string, asOverlay bool, basePath string) {
	if asOverlay && !cfg.SupportsOverlay {
		log.Fatalf("--as-overlay is not supported for %s exports", cfg.DisplayName)
	}

	items, err := cfg.ListAll(profile)
	if err != nil {
		log.Fatalf("Failed to list %s: %v", cfg.PluralName, err)
	}

	if len(items) == 0 {
		fmt.Printf("No %s found\n", cfg.PluralName)
		return
	}

	outputDir := directory
	if outputDir == "" {
		outputDir = "."
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	successCount := 0
	failedCount := 0

	fmt.Printf("Exporting %d %s...\n", len(items), cfg.PluralName)

	for i, item := range items {
		var rawData json.RawMessage
		var id string
		var err error

		if cfg.FetchByID != nil {
			// Use ID-based fetch for bulk export (handles disambiguated names)
			rawData, err = cfg.FetchByID(item.ID, profile)
			id = item.ID
		} else {
			rawData, id, err = cfg.FetchSingle(item.Name, profile)
		}
		if err != nil {
			fmt.Printf("Warning: Failed to fetch %s '%s': %v\n", cfg.DisplayName, item.Name, err)
			failedCount++
			continue
		}
		if rawData == nil {
			fmt.Printf("Warning: %s '%s' not found\n", cfg.DisplayName, item.Name)
			failedCount++
			continue
		}

		yamlContent, err := convertResourceToYAML(item.Name, id, cfg, rawData, asOverlay, basePath)
		if err != nil {
			fmt.Printf("Warning: Failed to convert %s '%s': %v\n", cfg.DisplayName, item.Name, err)
			failedCount++
			continue
		}

		filename := sanitizeFilename(item.Name) + ".yaml"
		fp := filepath.Join(outputDir, filename)

		if err := os.WriteFile(fp, yamlContent, 0644); err != nil {
			fmt.Printf("Warning: Failed to write %s: %v\n", filename, err)
			failedCount++
			continue
		}

		successCount++
		fmt.Printf("  [%d/%d] Exported %s\n", i+1, len(items), filename)
	}

	fmt.Printf("\nExport complete: %d successful, %d failed\n", successCount, failedCount)
}

// convertResourceToYAML converts raw JSON resource data to a declarative YAML spec
func convertResourceToYAML(name, id string, cfg ResourceExportConfig, rawData json.RawMessage, asOverlay bool, basePath string) ([]byte, error) {
	var specMap map[string]interface{}
	if err := json.Unmarshal(rawData, &specMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Remove ignore fields
	removeIgnoreFields(specMap, cfg.IgnoreFields)

	// Apply sanitization if defined
	if cfg.SanitizeSpec != nil {
		cfg.SanitizeSpec(specMap)
	}

	if asOverlay {
		if basePath == "" {
			return nil, fmt.Errorf("--as-overlay requires --base for %s resources", cfg.DisplayName)
		}
		return convertResourceToYAMLOverlay(name, id, cfg, specMap, basePath)
	}

	return convertResourceToYAMLFull(name, id, cfg, specMap)
}

// convertResourceToYAMLFull creates a full YAML export
func convertResourceToYAMLFull(name, id string, cfg ResourceExportConfig, specMap map[string]interface{}) ([]byte, error) {
	resourceSpec := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       cfg.Kind,
		Metadata: resources.Metadata{
			Name: name,
		},
		Spec: specMap,
	}

	header := fmt.Sprintf("# %s Configuration (Full Export)\n# Exported from VBR\n# Resource ID: %s\n#\n# This export contains the complete %s configuration.\n# All fields from the VBR API are preserved.\n\n", cfg.Kind, id, cfg.DisplayName)

	yamlBytes, err := yaml.Marshal(resourceSpec)
	if err != nil {
		return nil, err
	}

	result := []byte(header)
	result = append(result, yamlBytes...)
	return result, nil
}

// convertResourceToYAMLOverlay creates an overlay YAML export by diffing against a base
func convertResourceToYAMLOverlay(name, id string, cfg ResourceExportConfig, specMap map[string]interface{}, basePath string) ([]byte, error) {
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}

	var baseResource resources.ResourceSpec
	if err := yaml.Unmarshal(baseData, &baseResource); err != nil {
		return nil, fmt.Errorf("failed to parse base file: %w", err)
	}

	overlaySpec := calculateDiff(baseResource.Spec, specMap)

	resourceSpec := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       cfg.Kind,
		Metadata: resources.Metadata{
			Name: name,
		},
		Spec: overlaySpec,
	}

	header := fmt.Sprintf("# %s Overlay\n# Base: %s\n# Resource ID: %s\n#\n# This overlay contains only the fields that differ from the base.\n\n", cfg.Kind, basePath, id)

	yamlBytes, err := yaml.Marshal(resourceSpec)
	if err != nil {
		return nil, err
	}

	result := []byte(header)
	result = append(result, yamlBytes...)
	return result, nil
}

// removeIgnoreFields removes specified fields from a spec map
func removeIgnoreFields(spec map[string]interface{}, fields map[string]bool) {
	for key := range fields {
		delete(spec, key)
	}
}
