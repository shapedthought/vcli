package resources

import (
	"fmt"
	"reflect"
)

// MergeStrategy defines how to merge different types of fields
type MergeStrategy string

const (
	// StrategyMerge performs deep merge for maps and arrays
	StrategyMerge MergeStrategy = "merge"
	// StrategyReplace replaces base value with overlay value
	StrategyReplace MergeStrategy = "replace"
)

// MergeOptions configures the merge behavior
type MergeOptions struct {
	// ArrayStrategy determines how arrays are merged ("merge" or "replace")
	ArrayStrategy MergeStrategy
	// NullMeansDelete treats explicit null values as deletion markers
	NullMeansDelete bool
}

// DefaultMergeOptions returns sensible defaults for strategic merge
func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		ArrayStrategy:   StrategyReplace, // Arrays are replaced by default (simpler behavior)
		NullMeansDelete: false,           // Null values don't delete fields
	}
}

// MergeResourceSpecs performs a strategic merge of base and overlay ResourceSpecs
// Returns a new ResourceSpec with merged content
func MergeResourceSpecs(base, overlay ResourceSpec, opts MergeOptions) (ResourceSpec, error) {
	result := ResourceSpec{
		APIVersion: base.APIVersion,
		Kind:       base.Kind,
		Metadata:   base.Metadata,
		Spec:       make(map[string]interface{}),
	}

	// Overlay can override metadata
	if overlay.Metadata.Name != "" {
		result.Metadata.Name = overlay.Metadata.Name
	}
	if len(overlay.Metadata.Labels) > 0 {
		result.Metadata.Labels = mergeMaps(base.Metadata.Labels, overlay.Metadata.Labels)
	}
	if len(overlay.Metadata.Annotations) > 0 {
		result.Metadata.Annotations = mergeMaps(base.Metadata.Annotations, overlay.Metadata.Annotations)
	}

	// Deep merge the spec
	merged, err := mergeValues(base.Spec, overlay.Spec, opts)
	if err != nil {
		return result, fmt.Errorf("failed to merge specs: %w", err)
	}

	if mergedMap, ok := merged.(map[string]interface{}); ok {
		result.Spec = mergedMap
	} else {
		return result, fmt.Errorf("merged spec is not a map: %T", merged)
	}

	return result, nil
}

// mergeValues performs deep merge of two values based on their types
func mergeValues(base, overlay interface{}, opts MergeOptions) (interface{}, error) {
	// If overlay is nil or zero value, keep base
	if overlay == nil {
		if opts.NullMeansDelete {
			return nil, nil
		}
		return base, nil
	}

	// If base is nil, use overlay
	if base == nil {
		return overlay, nil
	}

	baseVal := reflect.ValueOf(base)
	overlayVal := reflect.ValueOf(overlay)

	// If types don't match, overlay replaces base
	if baseVal.Type() != overlayVal.Type() {
		return overlay, nil
	}

	// Handle based on type
	switch baseVal.Kind() {
	case reflect.Map:
		return mergeMapsInterface(base, overlay, opts)
	case reflect.Slice:
		return mergeSlices(base, overlay, opts)
	default:
		// For primitives, overlay wins
		return overlay, nil
	}
}

// mergeMapsInterface merges two map[string]interface{} values
func mergeMapsInterface(base, overlay interface{}, opts MergeOptions) (interface{}, error) {
	baseMap, ok := base.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("base is not map[string]interface{}: %T", base)
	}

	overlayMap, ok := overlay.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("overlay is not map[string]interface{}: %T", overlay)
	}

	result := make(map[string]interface{})

	// Copy all base keys
	for k, v := range baseMap {
		result[k] = v
	}

	// Merge overlay keys
	for k, overlayValue := range overlayMap {
		if baseValue, exists := baseMap[k]; exists {
			// Key exists in both - merge recursively
			merged, err := mergeValues(baseValue, overlayValue, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to merge key %s: %w", k, err)
			}
			if merged != nil || !opts.NullMeansDelete {
				result[k] = merged
			} else {
				// Delete the key if merged to nil and NullMeansDelete is true
				delete(result, k)
			}
		} else {
			// Key only in overlay - add it
			result[k] = overlayValue
		}
	}

	return result, nil
}

// mergeSlices merges two slice values based on the array strategy
func mergeSlices(base, overlay interface{}, opts MergeOptions) (interface{}, error) {
	if opts.ArrayStrategy == StrategyReplace {
		// Simple replacement
		return overlay, nil
	}

	// For merge strategy, append overlay to base
	baseSlice := reflect.ValueOf(base)
	overlaySlice := reflect.ValueOf(overlay)

	if baseSlice.Type() != overlaySlice.Type() {
		return overlay, nil
	}

	result := reflect.MakeSlice(baseSlice.Type(), 0, baseSlice.Len()+overlaySlice.Len())
	result = reflect.AppendSlice(result, baseSlice)
	result = reflect.AppendSlice(result, overlaySlice)

	return result.Interface(), nil
}

// DeepMergeMaps performs a deep merge of two map[string]interface{} values.
// Values from overlay override values in base. Nested maps are merged recursively.
// This is useful for merging spec configurations.
func DeepMergeMaps(base, overlay map[string]interface{}) (map[string]interface{}, error) {
	opts := DefaultMergeOptions()
	result, err := mergeMapsInterface(base, overlay, opts)
	if err != nil {
		return nil, err
	}
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	return nil, fmt.Errorf("merge result is not a map: %T", result)
}

// mergeMaps is a simple helper for string maps (labels, annotations)
func mergeMaps(base, overlay map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		result[k] = v
	}
	return result
}

// MergeYAMLFiles loads two YAML files and merges them
// This is a convenience function for the CLI
func MergeYAMLFiles(basePath, overlayPath string, opts MergeOptions) (ResourceSpec, error) {
	baseSpec, err := LoadResourceSpec(basePath)
	if err != nil {
		return ResourceSpec{}, fmt.Errorf("failed to load base spec: %w", err)
	}

	overlaySpec, err := LoadResourceSpec(overlayPath)
	if err != nil {
		return ResourceSpec{}, fmt.Errorf("failed to load overlay spec: %w", err)
	}

	// Validate that they're the same kind
	if baseSpec.Kind != overlaySpec.Kind {
		return ResourceSpec{}, fmt.Errorf("kind mismatch: base=%s, overlay=%s", baseSpec.Kind, overlaySpec.Kind)
	}

	return MergeResourceSpecs(baseSpec, overlaySpec, opts)
}
