package resources

import (
	"os"
	"path/filepath"
	"testing"
)

// assertMap extracts a map[string]interface{} from a parent map by key, failing the test if missing or wrong type.
func assertMap(t *testing.T, parent map[string]interface{}, key string) map[string]interface{} {
	t.Helper()
	val, ok := parent[key]
	if !ok {
		t.Fatalf("%q missing from map; got keys: %v", key, mapKeys(parent))
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		t.Fatalf("%q has unexpected type %T; value: %#v", key, val, val)
	}
	return m
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- ApplyGroupMergeFromSpecs tests ---

func TestApplyGroupMergeFromSpecs_NoProfileNoOverlay(t *testing.T) {
	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRJob,
		Metadata:   Metadata{Name: "my-job"},
		Spec: map[string]interface{}{
			"description": "original",
			"isDisabled":  false,
		},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, nil, nil, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata.Name != "my-job" {
		t.Errorf("Name = %q, want %q", result.Metadata.Name, "my-job")
	}
	if result.Kind != KindVBRJob {
		t.Errorf("Kind = %q, want %q", result.Kind, KindVBRJob)
	}
	if result.Spec["description"] != "original" {
		t.Errorf("description = %v, want %q", result.Spec["description"], "original")
	}
}

func TestApplyGroupMergeFromSpecs_ProfileOnly(t *testing.T) {
	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRJob,
		Metadata:   Metadata{Name: "my-job"},
		Spec: map[string]interface{}{
			"description": "spec description",
		},
	}
	profile := &ResourceSpec{
		Kind: KindProfile,
		Spec: map[string]interface{}{
			"description": "profile description",
			"storage": map[string]interface{}{
				"retentionPolicy": map[string]interface{}{
					"type":     "Days",
					"quantity": 30,
				},
			},
		},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, profile, nil, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Spec wins over profile for description
	if result.Spec["description"] != "spec description" {
		t.Errorf("description = %v, want %q (spec should override profile)", result.Spec["description"], "spec description")
	}

	// Profile provides defaults for fields not in spec
	storage, ok := result.Spec["storage"].(map[string]interface{})
	if !ok {
		t.Fatal("storage should be present from profile defaults")
	}
	retention, ok := storage["retentionPolicy"].(map[string]interface{})
	if !ok {
		t.Fatal("retentionPolicy should be present from profile")
	}
	if retention["quantity"] != 30 {
		t.Errorf("retention quantity = %v, want 30", retention["quantity"])
	}
}

func TestApplyGroupMergeFromSpecs_OverlayOnly(t *testing.T) {
	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRJob,
		Metadata:   Metadata{Name: "my-job"},
		Spec: map[string]interface{}{
			"description": "spec description",
			"storage": map[string]interface{}{
				"retentionPolicy": map[string]interface{}{
					"type":     "Days",
					"quantity": 14,
				},
			},
		},
	}
	overlay := &ResourceSpec{
		Kind: KindOverlay,
		Spec: map[string]interface{}{
			"storage": map[string]interface{}{
				"retentionPolicy": map[string]interface{}{
					"quantity": 90,
				},
			},
		},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, nil, overlay, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Overlay wins for retention quantity
	storage := assertMap(t, result.Spec, "storage")
	retention := assertMap(t, storage, "retentionPolicy")
	if retention["quantity"] != 90 {
		t.Errorf("retention quantity = %v, want 90 (overlay should override spec)", retention["quantity"])
	}
	// Spec's type should be preserved (overlay didn't set it)
	if retention["type"] != "Days" {
		t.Errorf("retention type = %v, want %q (should be preserved from spec)", retention["type"], "Days")
	}
	// Spec's description should be preserved
	if result.Spec["description"] != "spec description" {
		t.Errorf("description = %v, want %q", result.Spec["description"], "spec description")
	}
}

func TestApplyGroupMergeFromSpecs_ProfileAndOverlay(t *testing.T) {
	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRJob,
		Metadata:   Metadata{Name: "my-job"},
		Spec: map[string]interface{}{
			"description": "spec description",
		},
	}
	profile := &ResourceSpec{
		Kind: KindProfile,
		Spec: map[string]interface{}{
			"description": "profile description",
			"storage": map[string]interface{}{
				"retentionPolicy": map[string]interface{}{
					"type":     "Days",
					"quantity": 30,
				},
			},
			"schedule": map[string]interface{}{
				"runAutomatically": true,
			},
		},
	}
	overlay := &ResourceSpec{
		Kind: KindOverlay,
		Spec: map[string]interface{}{
			"storage": map[string]interface{}{
				"retentionPolicy": map[string]interface{}{
					"quantity": 90,
				},
			},
		},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, profile, overlay, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Spec wins over profile for description
	if result.Spec["description"] != "spec description" {
		t.Errorf("description = %v, want %q", result.Spec["description"], "spec description")
	}

	// Overlay wins for retention quantity
	storage := assertMap(t, result.Spec, "storage")
	retention := assertMap(t, storage, "retentionPolicy")
	if retention["quantity"] != 90 {
		t.Errorf("retention quantity = %v, want 90 (overlay wins)", retention["quantity"])
	}

	// Profile provides defaults for type (overlay didn't change it)
	if retention["type"] != "Days" {
		t.Errorf("retention type = %v, want %q (from profile defaults)", retention["type"], "Days")
	}

	// Profile provides schedule defaults (overlay and spec didn't set it)
	schedule, ok := result.Spec["schedule"].(map[string]interface{})
	if !ok {
		t.Fatal("schedule should be present from profile defaults")
	}
	if schedule["runAutomatically"] != true {
		t.Errorf("runAutomatically = %v, want true", schedule["runAutomatically"])
	}
}

// --- Kind validation tests ---

func TestApplyGroupMergeFromSpecs_RejectsMixinKindSpec(t *testing.T) {
	spec := ResourceSpec{
		Kind:     KindProfile,
		Metadata: Metadata{Name: "not-a-resource"},
		Spec:     map[string]interface{}{},
	}

	_, err := ApplyGroupMergeFromSpecs(spec, nil, nil, DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for mixin kind spec, got nil")
	}
}

func TestApplyGroupMergeFromSpecs_RejectsOverlayKindSpec(t *testing.T) {
	spec := ResourceSpec{
		Kind:     KindOverlay,
		Metadata: Metadata{Name: "not-a-resource"},
		Spec:     map[string]interface{}{},
	}

	_, err := ApplyGroupMergeFromSpecs(spec, nil, nil, DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for Overlay kind spec, got nil")
	}
}

func TestApplyGroupMergeFromSpecs_RejectsInvalidProfileKind(t *testing.T) {
	spec := ResourceSpec{
		Kind:     KindVBRJob,
		Metadata: Metadata{Name: "my-job"},
		Spec:     map[string]interface{}{},
	}
	badProfile := &ResourceSpec{
		Kind: KindVBRJob, // Should be KindProfile
		Spec: map[string]interface{}{},
	}

	_, err := ApplyGroupMergeFromSpecs(spec, badProfile, nil, DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for non-Profile kind in profile arg, got nil")
	}
}

func TestApplyGroupMergeFromSpecs_RejectsInvalidOverlayKind(t *testing.T) {
	spec := ResourceSpec{
		Kind:     KindVBRJob,
		Metadata: Metadata{Name: "my-job"},
		Spec:     map[string]interface{}{},
	}
	badOverlay := &ResourceSpec{
		Kind: KindVBRJob, // Should be KindOverlay
		Spec: map[string]interface{}{},
	}

	_, err := ApplyGroupMergeFromSpecs(spec, nil, badOverlay, DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for non-Overlay kind in overlay arg, got nil")
	}
}

func TestApplyGroupMergeFromSpecs_AcceptsAllResourceKinds(t *testing.T) {
	kinds := []string{
		KindVBRJob,
		KindVBRRepository,
		KindVBRSOBR,
		KindVBRScaleOutRepository,
		KindVBREncryptionPassword,
		KindVBRKmsServer,
	}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			spec := ResourceSpec{
				Kind:     kind,
				Metadata: Metadata{Name: "test"},
				Spec:     map[string]interface{}{},
			}
			_, err := ApplyGroupMergeFromSpecs(spec, nil, nil, DefaultMergeOptions())
			if err != nil {
				t.Errorf("kind %q should be accepted, got error: %v", kind, err)
			}
		})
	}
}

// --- Metadata preservation tests ---

func TestApplyGroupMergeFromSpecs_PreservesMetadataName(t *testing.T) {
	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRJob,
		Metadata:   Metadata{Name: "original-name"},
		Spec:       map[string]interface{}{},
	}
	profile := &ResourceSpec{
		Kind:     KindProfile,
		Metadata: Metadata{Name: "profile-name"},
		Spec:     map[string]interface{}{},
	}
	overlay := &ResourceSpec{
		Kind:     KindOverlay,
		Metadata: Metadata{Name: "overlay-name"},
		Spec:     map[string]interface{}{},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, profile, overlay, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Name must always come from spec, not profile or overlay
	if result.Metadata.Name != "original-name" {
		t.Errorf("Name = %q, want %q (spec name must be preserved)", result.Metadata.Name, "original-name")
	}
}

func TestApplyGroupMergeFromSpecs_PreservesAPIVersionAndKind(t *testing.T) {
	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRRepository,
		Metadata:   Metadata{Name: "my-repo"},
		Spec:       map[string]interface{}{},
	}
	profile := &ResourceSpec{
		Kind: KindProfile,
		Spec: map[string]interface{}{"foo": "bar"},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, profile, nil, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.APIVersion != "owlctl.veeam.com/v1" {
		t.Errorf("APIVersion = %q, want %q", result.APIVersion, "owlctl.veeam.com/v1")
	}
	if result.Kind != KindVBRRepository {
		t.Errorf("Kind = %q, want %q", result.Kind, KindVBRRepository)
	}
}

func TestApplyGroupMergeFromSpecs_MergesLabelsAdditively(t *testing.T) {
	spec := ResourceSpec{
		Kind: KindVBRJob,
		Metadata: Metadata{
			Name:   "my-job",
			Labels: map[string]string{"app": "backup", "tier": "gold"},
		},
		Spec: map[string]interface{}{},
	}
	profile := &ResourceSpec{
		Kind: KindProfile,
		Metadata: Metadata{
			Labels: map[string]string{"team": "platform", "tier": "silver"},
		},
		Spec: map[string]interface{}{},
	}
	overlay := &ResourceSpec{
		Kind: KindOverlay,
		Metadata: Metadata{
			Labels: map[string]string{"env": "prod", "tier": "platinum"},
		},
		Spec: map[string]interface{}{},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, profile, overlay, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: profile labels first, spec overrides, then overlay wins on top
	// profile: team=platform, tier=silver
	// + spec:  app=backup, tier=gold  → team=platform, app=backup, tier=gold
	// + overlay: env=prod, tier=platinum → team=platform, app=backup, env=prod, tier=platinum
	want := map[string]string{
		"team": "platform",
		"app":  "backup",
		"env":  "prod",
		"tier": "platinum",
	}

	for k, v := range want {
		if result.Metadata.Labels[k] != v {
			t.Errorf("Labels[%q] = %q, want %q", k, result.Metadata.Labels[k], v)
		}
	}
	if len(result.Metadata.Labels) != len(want) {
		t.Errorf("Labels length = %d, want %d", len(result.Metadata.Labels), len(want))
	}
}

func TestApplyGroupMergeFromSpecs_MergesAnnotationsAdditively(t *testing.T) {
	spec := ResourceSpec{
		Kind: KindVBRJob,
		Metadata: Metadata{
			Name:        "my-job",
			Annotations: map[string]string{"note": "from-spec"},
		},
		Spec: map[string]interface{}{},
	}
	overlay := &ResourceSpec{
		Kind: KindOverlay,
		Metadata: Metadata{
			Annotations: map[string]string{"deploy": "ci", "note": "from-overlay"},
		},
		Spec: map[string]interface{}{},
	}

	result, err := ApplyGroupMergeFromSpecs(spec, nil, overlay, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata.Annotations["note"] != "from-overlay" {
		t.Errorf("Annotations[note] = %q, want %q (overlay wins)", result.Metadata.Annotations["note"], "from-overlay")
	}
	if result.Metadata.Annotations["deploy"] != "ci" {
		t.Errorf("Annotations[deploy] = %q, want %q", result.Metadata.Annotations["deploy"], "ci")
	}
}

func TestApplyGroupMergeFromSpecs_DoesNotMutateInput(t *testing.T) {
	spec := ResourceSpec{
		Kind:     KindVBRJob,
		Metadata: Metadata{Name: "my-job", Labels: map[string]string{"app": "backup"}},
		Spec:     map[string]interface{}{"description": "original"},
	}
	profile := &ResourceSpec{
		Kind:     KindProfile,
		Metadata: Metadata{Labels: map[string]string{"team": "platform"}},
		Spec:     map[string]interface{}{"storage": map[string]interface{}{"compression": "Optimal"}},
	}

	_, err := ApplyGroupMergeFromSpecs(spec, profile, nil, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Spec's labels should not have been mutated
	if _, ok := spec.Metadata.Labels["team"]; ok {
		t.Error("spec labels were mutated — 'team' key should not exist")
	}

	// Spec's Spec map should not have been mutated
	if _, ok := spec.Spec["storage"]; ok {
		t.Error("spec.Spec was mutated — 'storage' key should not exist")
	}
	if len(spec.Spec) != 1 {
		t.Errorf("spec.Spec length = %d, want 1 (only 'description')", len(spec.Spec))
	}

	// Profile's Spec and Labels should not have been mutated
	if _, ok := profile.Spec["description"]; ok {
		t.Error("profile.Spec was mutated — 'description' key should not exist")
	}
	if _, ok := profile.Metadata.Labels["app"]; ok {
		t.Error("profile.Metadata.Labels was mutated — 'app' key should not exist")
	}
}

// --- ApplyGroupMergeFromSpec (file-path based) tests ---

func writeTestYAML(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", filename, err)
	}
	return path
}

func TestApplyGroupMergeFromSpec_FileBasedMerge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	profilePath := writeTestYAML(t, tmpDir, "profile.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: Profile
metadata:
  name: defaults
  labels:
    team: platform
spec:
  storage:
    retentionPolicy:
      type: Days
      quantity: 30
  schedule:
    runAutomatically: true
`)

	overlayPath := writeTestYAML(t, tmpDir, "overlay.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: Overlay
metadata:
  name: prod-overlay
  labels:
    env: prod
spec:
  storage:
    retentionPolicy:
      quantity: 90
`)

	spec := ResourceSpec{
		APIVersion: "owlctl.veeam.com/v1",
		Kind:       KindVBRJob,
		Metadata:   Metadata{Name: "my-job", Labels: map[string]string{"app": "backup"}},
		Spec: map[string]interface{}{
			"description": "my job",
		},
	}

	result, err := ApplyGroupMergeFromSpec(spec, profilePath, overlayPath, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Name preserved from spec
	if result.Metadata.Name != "my-job" {
		t.Errorf("Name = %q, want %q", result.Metadata.Name, "my-job")
	}

	// Description from spec
	if result.Spec["description"] != "my job" {
		t.Errorf("description = %v, want %q", result.Spec["description"], "my job")
	}

	// Retention quantity from overlay (90), type from profile (Days)
	storage := assertMap(t, result.Spec, "storage")
	retention := assertMap(t, storage, "retentionPolicy")
	if retention["quantity"] != 90 {
		t.Errorf("retention quantity = %v, want 90", retention["quantity"])
	}
	if retention["type"] != "Days" {
		t.Errorf("retention type = %v, want %q", retention["type"], "Days")
	}

	// Schedule from profile
	schedule := assertMap(t, result.Spec, "schedule")
	if schedule["runAutomatically"] != true {
		t.Errorf("runAutomatically = %v, want true", schedule["runAutomatically"])
	}

	// Labels: profile(team=platform) + spec(app=backup) + overlay(env=prod)
	if result.Metadata.Labels["team"] != "platform" {
		t.Errorf("Labels[team] = %q, want %q", result.Metadata.Labels["team"], "platform")
	}
	if result.Metadata.Labels["app"] != "backup" {
		t.Errorf("Labels[app] = %q, want %q", result.Metadata.Labels["app"], "backup")
	}
	if result.Metadata.Labels["env"] != "prod" {
		t.Errorf("Labels[env] = %q, want %q", result.Metadata.Labels["env"], "prod")
	}
}

func TestApplyGroupMergeFromSpec_EmptyPaths(t *testing.T) {
	spec := ResourceSpec{
		Kind:     KindVBRJob,
		Metadata: Metadata{Name: "test"},
		Spec:     map[string]interface{}{"foo": "bar"},
	}

	// Empty string paths should skip profile/overlay
	result, err := ApplyGroupMergeFromSpec(spec, "", "", DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Spec["foo"] != "bar" {
		t.Errorf("Spec[foo] = %v, want %q", result.Spec["foo"], "bar")
	}
}

func TestApplyGroupMergeFromSpec_InvalidProfileKind(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	profilePath := writeTestYAML(t, tmpDir, "bad-profile.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: VBRJob
metadata:
  name: not-a-profile
spec:
  description: wrong kind
`)

	spec := ResourceSpec{
		Kind:     KindVBRJob,
		Metadata: Metadata{Name: "test"},
		Spec:     map[string]interface{}{},
	}

	_, err = ApplyGroupMergeFromSpec(spec, profilePath, "", DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for non-Profile kind, got nil")
	}
}

func TestApplyGroupMergeFromSpec_InvalidOverlayKind(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	overlayPath := writeTestYAML(t, tmpDir, "bad-overlay.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: VBRJob
metadata:
  name: not-an-overlay
spec:
  description: wrong kind
`)

	spec := ResourceSpec{
		Kind:     KindVBRJob,
		Metadata: Metadata{Name: "test"},
		Spec:     map[string]interface{}{},
	}

	_, err = ApplyGroupMergeFromSpec(spec, "", overlayPath, DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for non-Overlay kind, got nil")
	}
}

// --- ApplyGroupMerge (full file-path based) tests ---

func TestApplyGroupMerge_LoadsSpecFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	specPath := writeTestYAML(t, tmpDir, "spec.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: VBRJob
metadata:
  name: file-loaded-job
spec:
  description: loaded from file
  isDisabled: false
`)

	result, err := ApplyGroupMerge(specPath, "", "", DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata.Name != "file-loaded-job" {
		t.Errorf("Name = %q, want %q", result.Metadata.Name, "file-loaded-job")
	}
	if result.Spec["description"] != "loaded from file" {
		t.Errorf("description = %v, want %q", result.Spec["description"], "loaded from file")
	}
}

func TestApplyGroupMerge_FullThreeWay(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "owlctl-merge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	specPath := writeTestYAML(t, tmpDir, "spec.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: VBRJob
metadata:
  name: three-way-job
spec:
  description: spec value
  isDisabled: false
`)

	profilePath := writeTestYAML(t, tmpDir, "profile.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: Profile
metadata:
  name: defaults
spec:
  description: profile value
  storage:
    retentionPolicy:
      type: Days
      quantity: 14
`)

	overlayPath := writeTestYAML(t, tmpDir, "overlay.yaml", `
apiVersion: owlctl.veeam.com/v1
kind: Overlay
metadata:
  name: prod
spec:
  storage:
    retentionPolicy:
      quantity: 90
`)

	result, err := ApplyGroupMerge(specPath, profilePath, overlayPath, DefaultMergeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Name from spec
	if result.Metadata.Name != "three-way-job" {
		t.Errorf("Name = %q, want %q", result.Metadata.Name, "three-way-job")
	}

	// Description: spec wins over profile
	if result.Spec["description"] != "spec value" {
		t.Errorf("description = %v, want %q", result.Spec["description"], "spec value")
	}

	// Retention: overlay wins quantity, profile provides type
	storage := assertMap(t, result.Spec, "storage")
	retention := assertMap(t, storage, "retentionPolicy")
	if retention["quantity"] != 90 {
		t.Errorf("retention quantity = %v, want 90", retention["quantity"])
	}
	if retention["type"] != "Days" {
		t.Errorf("retention type = %v, want %q", retention["type"], "Days")
	}

	// isDisabled from spec (not in profile or overlay)
	if result.Spec["isDisabled"] != false {
		t.Errorf("isDisabled = %v, want false", result.Spec["isDisabled"])
	}
}

func TestApplyGroupMerge_NonExistentSpec(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "spec.yaml")
	_, err := ApplyGroupMerge(missingPath, "", "", DefaultMergeOptions())
	if err == nil {
		t.Fatal("expected error for non-existent spec file, got nil")
	}
}
