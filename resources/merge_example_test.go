package resources_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shapedthought/owlctl/resources"
	"gopkg.in/yaml.v3"
)

// TestMergeExampleFiles demonstrates the overlay system with real example files
func TestMergeExampleFiles(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "owlctl-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create base template
	base := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
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
		APIVersion: "owlctl.veeam.com/v1",
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

// TestApplyGroupMerge tests the 3-way group merge (Profile + Spec + Overlay)
func TestApplyGroupMerge(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "owlctl-group-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create profile (kind: Profile - base defaults)
	profile := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       "Profile",
		Metadata: resources.Metadata{
			Name:   "gold-standard",
			Labels: map[string]string{"tier": "gold", "managed-by": "owlctl"},
		},
		Spec: map[string]interface{}{
			"type":        "VSphereBackup",
			"description": "Gold standard backup policy",
			"storage": map[string]interface{}{
				"compression": "Optimal",
				"retention": map[string]interface{}{
					"type":     "Days",
					"quantity": 30,
				},
			},
			"schedule": map[string]interface{}{
				"daily": map[string]interface{}{
					"isEnabled": true,
					"localTime": "22:00",
				},
			},
		},
	}

	// Create spec (kind: VBRJob - identity + exceptions)
	spec := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       "VBRJob",
		Metadata: resources.Metadata{
			Name:   "sql-backup",
			Labels: map[string]string{"app": "database", "managed-by": "spec-override"},
		},
		Spec: map[string]interface{}{
			"description": "SQL Server backup",
			"virtualMachines": map[string]interface{}{
				"includes": []interface{}{"sql-vm-01", "sql-vm-02"},
			},
			"storage": map[string]interface{}{
				"retention": map[string]interface{}{
					"quantity": 14,
				},
			},
		},
	}

	// Create overlay (kind: Overlay - policy patch)
	overlay := resources.ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       "Overlay",
		Metadata: resources.Metadata{
			Name:   "compliance-patch",
			Labels: map[string]string{"compliance": "sox"},
		},
		Spec: map[string]interface{}{
			"storage": map[string]interface{}{
				"advancedSettings": map[string]interface{}{
					"storageData": map[string]interface{}{
						"encryption": map[string]interface{}{
							"isEnabled": true,
						},
					},
				},
			},
		},
	}

	// Save all files
	profilePath := filepath.Join(tmpDir, "profile.yaml")
	specPath := filepath.Join(tmpDir, "spec.yaml")
	overlayPath := filepath.Join(tmpDir, "overlay.yaml")

	if err := resources.SaveResourceSpec(profile, profilePath); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}
	if err := resources.SaveResourceSpec(spec, specPath); err != nil {
		t.Fatalf("Failed to save spec: %v", err)
	}
	if err := resources.SaveResourceSpec(overlay, overlayPath); err != nil {
		t.Fatalf("Failed to save overlay: %v", err)
	}

	t.Run("ProfileAndSpec", func(t *testing.T) {
		merged, err := resources.ApplyGroupMerge(specPath, profilePath, "", resources.DefaultMergeOptions())
		if err != nil {
			t.Fatalf("ApplyGroupMerge failed: %v", err)
		}

		// Spec metadata preserved
		if merged.Kind != "VBRJob" {
			t.Errorf("Expected Kind=VBRJob, got %s", merged.Kind)
		}
		if merged.Metadata.Name != "sql-backup" {
			t.Errorf("Expected Name=sql-backup, got %s", merged.Metadata.Name)
		}

		// Spec description overrides profile
		if desc, ok := merged.Spec["description"].(string); !ok || desc != "SQL Server backup" {
			t.Errorf("Expected description from spec, got %v", merged.Spec["description"])
		}

		// Profile type used (spec didn't set it)
		if typ, ok := merged.Spec["type"].(string); !ok || typ != "VSphereBackup" {
			t.Errorf("Expected type from profile, got %v", merged.Spec["type"])
		}

		// Spec retention overrides profile retention quantity
		storage, _ := merged.Spec["storage"].(map[string]interface{})
		retention, _ := storage["retention"].(map[string]interface{})
		if quantity, ok := retention["quantity"].(int); !ok || quantity != 14 {
			t.Errorf("Expected retention quantity=14 from spec, got %v", retention["quantity"])
		}
		// Profile retention type preserved
		if retType, ok := retention["type"].(string); !ok || retType != "Days" {
			t.Errorf("Expected retention type=Days from profile, got %v", retention["type"])
		}

		// Profile compression preserved
		if comp, ok := storage["compression"].(string); !ok || comp != "Optimal" {
			t.Errorf("Expected compression=Optimal from profile, got %v", storage["compression"])
		}

		// Labels merged additively, spec wins on conflict
		if merged.Metadata.Labels["tier"] != "gold" {
			t.Errorf("Expected label tier=gold from profile, got %v", merged.Metadata.Labels["tier"])
		}
		if merged.Metadata.Labels["app"] != "database" {
			t.Errorf("Expected label app=database from spec, got %v", merged.Metadata.Labels["app"])
		}
		if merged.Metadata.Labels["managed-by"] != "spec-override" {
			t.Errorf("Expected label managed-by=spec-override (spec wins), got %v", merged.Metadata.Labels["managed-by"])
		}
	})

	t.Run("SpecAndOverlay", func(t *testing.T) {
		merged, err := resources.ApplyGroupMerge(specPath, "", overlayPath, resources.DefaultMergeOptions())
		if err != nil {
			t.Fatalf("ApplyGroupMerge failed: %v", err)
		}

		// Spec metadata preserved
		if merged.Metadata.Name != "sql-backup" {
			t.Errorf("Expected Name=sql-backup, got %s", merged.Metadata.Name)
		}

		// Overlay encryption added
		storage, _ := merged.Spec["storage"].(map[string]interface{})
		advanced, _ := storage["advancedSettings"].(map[string]interface{})
		storageData, _ := advanced["storageData"].(map[string]interface{})
		encryption, _ := storageData["encryption"].(map[string]interface{})
		if isEnabled, ok := encryption["isEnabled"].(bool); !ok || !isEnabled {
			t.Errorf("Expected encryption.isEnabled=true from overlay, got %v", encryption["isEnabled"])
		}

		// Spec fields preserved
		if desc, ok := merged.Spec["description"].(string); !ok || desc != "SQL Server backup" {
			t.Errorf("Expected description from spec, got %v", merged.Spec["description"])
		}

		// Labels: spec + overlay combined
		if merged.Metadata.Labels["compliance"] != "sox" {
			t.Errorf("Expected label compliance=sox from overlay, got %v", merged.Metadata.Labels["compliance"])
		}
		if merged.Metadata.Labels["app"] != "database" {
			t.Errorf("Expected label app=database from spec, got %v", merged.Metadata.Labels["app"])
		}
	})

	t.Run("FullThreeWayMerge", func(t *testing.T) {
		merged, err := resources.ApplyGroupMerge(specPath, profilePath, overlayPath, resources.DefaultMergeOptions())
		if err != nil {
			t.Fatalf("ApplyGroupMerge failed: %v", err)
		}

		// Identity from spec
		if merged.Kind != "VBRJob" {
			t.Errorf("Expected Kind=VBRJob, got %s", merged.Kind)
		}
		if merged.Metadata.Name != "sql-backup" {
			t.Errorf("Expected Name=sql-backup, got %s", merged.Metadata.Name)
		}

		// Profile type preserved
		if typ, ok := merged.Spec["type"].(string); !ok || typ != "VSphereBackup" {
			t.Errorf("Expected type from profile, got %v", merged.Spec["type"])
		}

		// Spec description preserved
		if desc, ok := merged.Spec["description"].(string); !ok || desc != "SQL Server backup" {
			t.Errorf("Expected description from spec, got %v", merged.Spec["description"])
		}

		// Overlay encryption added
		storage, _ := merged.Spec["storage"].(map[string]interface{})
		advanced, _ := storage["advancedSettings"].(map[string]interface{})
		storageData, _ := advanced["storageData"].(map[string]interface{})
		encryption, _ := storageData["encryption"].(map[string]interface{})
		if isEnabled, ok := encryption["isEnabled"].(bool); !ok || !isEnabled {
			t.Errorf("Expected encryption.isEnabled=true from overlay, got %v", encryption["isEnabled"])
		}

		// All three label sources combined
		expectedLabels := map[string]string{
			"tier":       "gold",
			"app":        "database",
			"managed-by": "spec-override",
			"compliance": "sox",
		}
		for k, v := range expectedLabels {
			if merged.Metadata.Labels[k] != v {
				t.Errorf("Expected label %s=%s, got %s", k, v, merged.Metadata.Labels[k])
			}
		}
	})

	t.Run("InvalidProfileKind", func(t *testing.T) {
		// Use a VBRJob file as profile — should fail
		_, err := resources.ApplyGroupMerge(specPath, specPath, "", resources.DefaultMergeOptions())
		if err == nil {
			t.Fatal("Expected error for invalid profile kind")
		}
	})

	t.Run("InvalidOverlayKind", func(t *testing.T) {
		// Use a Profile file as overlay — should fail
		_, err := resources.ApplyGroupMerge(specPath, "", profilePath, resources.DefaultMergeOptions())
		if err == nil {
			t.Fatal("Expected error for invalid overlay kind")
		}
	})

	t.Run("SpecMetadataPreservedThroughAllMerges", func(t *testing.T) {
		merged, err := resources.ApplyGroupMerge(specPath, profilePath, overlayPath, resources.DefaultMergeOptions())
		if err != nil {
			t.Fatalf("ApplyGroupMerge failed: %v", err)
		}

		// metadata.name must ALWAYS come from spec, never profile or overlay
		if merged.Metadata.Name != "sql-backup" {
			t.Errorf("Expected metadata.name=sql-backup preserved from spec, got %s", merged.Metadata.Name)
		}

		// apiVersion and kind from spec
		if merged.APIVersion != "owlctl.veeam.com/v1" {
			t.Errorf("Expected apiVersion from spec, got %s", merged.APIVersion)
		}
		if merged.Kind != "VBRJob" {
			t.Errorf("Expected Kind from spec, got %s", merged.Kind)
		}
	})
}
