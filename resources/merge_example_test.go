package resources_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shapedthought/vcli/resources"
	"gopkg.in/yaml.v3"
)

// TestMergeExampleFiles demonstrates the overlay system with real example files
func TestMergeExampleFiles(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "vcli-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create base template
	base := resources.ResourceSpec{
		APIVersion: "vcli.veeam.com/v1",
		Kind:       "VBRJob",
		Metadata: resources.Metadata{
			Name:   "test-job",
			Labels: map[string]string{"app": "database"},
		},
		Spec: map[string]interface{}{
			"type":        "VSphereBackup",
			"description": "Base backup job",
			"repository":  "default-repo",
			"storage": map[string]interface{}{
				"compression": "Optimal",
				"retention": map[string]interface{}{
					"type":     "Days",
					"quantity": 7,
				},
			},
		},
	}

	// Create prod overlay
	prodOverlay := resources.ResourceSpec{
		APIVersion: "vcli.veeam.com/v1",
		Kind:       "VBRJob",
		Metadata: resources.Metadata{
			Name:   "test-job",
			Labels: map[string]string{"env": "production"},
		},
		Spec: map[string]interface{}{
			"description": "Production backup job (30-day retention)",
			"repository":  "prod-repo",
			"storage": map[string]interface{}{
				"retention": map[string]interface{}{
					"quantity": 30,
				},
			},
		},
	}

	// Save files
	basePath := filepath.Join(tmpDir, "base.yaml")
	prodPath := filepath.Join(tmpDir, "prod-overlay.yaml")

	if err := resources.SaveResourceSpec(base, basePath); err != nil {
		t.Fatalf("Failed to save base: %v", err)
	}
	if err := resources.SaveResourceSpec(prodOverlay, prodPath); err != nil {
		t.Fatalf("Failed to save overlay: %v", err)
	}

	// Test the merge
	merged, err := resources.MergeYAMLFiles(basePath, prodPath, resources.DefaultMergeOptions())
	if err != nil {
		t.Fatalf("Failed to merge: %v", err)
	}

	// Verify merged result
	if merged.Metadata.Name != "test-job" {
		t.Errorf("Expected name 'test-job', got '%s'", merged.Metadata.Name)
	}

	// Check that both labels are present (base + overlay)
	if len(merged.Metadata.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d: %v", len(merged.Metadata.Labels), merged.Metadata.Labels)
	}
	if merged.Metadata.Labels["app"] != "database" {
		t.Errorf("Expected label app=database, got %v", merged.Metadata.Labels["app"])
	}
	if merged.Metadata.Labels["env"] != "production" {
		t.Errorf("Expected label env=production, got %v", merged.Metadata.Labels["env"])
	}

	// Check spec fields
	if description, ok := merged.Spec["description"].(string); ok {
		expected := "Production backup job (30-day retention)"
		if description != expected {
			t.Errorf("Expected description '%s', got '%s'", expected, description)
		}
	} else {
		t.Error("Description field missing or wrong type")
	}

	if repo, ok := merged.Spec["repository"].(string); ok {
		if repo != "prod-repo" {
			t.Errorf("Expected repository 'prod-repo', got '%s'", repo)
		}
	} else {
		t.Error("Repository field missing or wrong type")
	}

	// Check nested storage.retention.quantity
	if storage, ok := merged.Spec["storage"].(map[string]interface{}); ok {
		if retention, ok := storage["retention"].(map[string]interface{}); ok {
			if quantity, ok := retention["quantity"].(int); ok {
				if quantity != 30 {
					t.Errorf("Expected retention quantity 30, got %d", quantity)
				}
			} else {
				t.Error("Retention quantity missing or wrong type")
			}
			// Verify that base type is still present
			if retType, ok := retention["type"].(string); ok {
				if retType != "Days" {
					t.Errorf("Expected retention type 'Days', got '%s'", retType)
				}
			} else {
				t.Error("Retention type missing (should be preserved from base)")
			}
		} else {
			t.Error("Storage retention missing")
		}
		// Verify that base compression is still present
		if compression, ok := storage["compression"].(string); ok {
			if compression != "Optimal" {
				t.Errorf("Expected compression 'Optimal', got '%s'", compression)
			}
		} else {
			t.Error("Compression missing (should be preserved from base)")
		}
	} else {
		t.Error("Storage field missing or wrong type")
	}

	// Save and display the merged result for visual inspection
	mergedPath := filepath.Join(tmpDir, "merged-result.yaml")
	if err := resources.SaveResourceSpec(merged, mergedPath); err != nil {
		t.Fatalf("Failed to save merged result: %v", err)
	}

	// Print the merged YAML for manual verification
	mergedYAML, _ := yaml.Marshal(merged)
	t.Logf("Merged result:\n%s", string(mergedYAML))
}
