package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/shapedthought/vcli/resources"
	"gopkg.in/yaml.v3"
)

// ResourceExportConfig holds per-resource-type export configuration
type ResourceExportConfig struct {
	Kind         string
	IgnoreFields map[string]bool
	HeaderLines  []string
}

// convertResourceToYAML converts a named resource (as json.RawMessage) to declarative YAML
func convertResourceToYAML(name string, rawData json.RawMessage, cfg ResourceExportConfig) ([]byte, error) {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(rawData, &dataMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource data: %w", err)
	}

	// Strip ignored fields
	dataMap = stripIgnoredFields(dataMap, cfg.IgnoreFields)

	resourceSpec := resources.ResourceSpec{
		APIVersion: "vcli.veeam.com/v1",
		Kind:       cfg.Kind,
		Metadata: resources.Metadata{
			Name: name,
		},
		Spec: dataMap,
	}

	// Build header comment
	header := ""
	for _, line := range cfg.HeaderLines {
		header += line + "\n"
	}
	header += "\n"

	yamlBytes, err := yaml.Marshal(resourceSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	result := []byte(header)
	result = append(result, yamlBytes...)
	return result, nil
}

// stripIgnoredFields returns a new map with the specified keys removed
func stripIgnoredFields(data map[string]interface{}, ignoreFields map[string]bool) map[string]interface{} {
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		if ignoreFields[k] {
			continue
		}
		out[k] = v
	}
	return out
}

// writeExportOutput writes YAML content to a file or stdout for single-resource export
func writeExportOutput(yamlContent []byte, outputPath, resourceName string) {
	if outputPath != "" {
		if err := os.WriteFile(outputPath, yamlContent, 0644); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
		fmt.Printf("Exported %s to %s\n", resourceName, outputPath)
	} else {
		fmt.Println(string(yamlContent))
	}
}

// writeExportAllOutput writes YAML content to a directory for bulk export
func writeExportAllOutput(yamlContent []byte, outputDir, resourceName string, index, total int) error {
	filename := sanitizeFilename(resourceName) + ".yaml"
	path := filepath.Join(outputDir, filename)

	if err := os.WriteFile(path, yamlContent, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	fmt.Printf("  [%d/%d] Exported %s\n", index+1, total, filename)
	return nil
}
