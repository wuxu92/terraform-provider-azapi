package azwise

import "strings"

// extractNestedValue traverses a map using a dot-separated path and returns the value.
// Returns nil if any segment is missing or the intermediate value is not a map.
func extractNestedValue(m map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = m
	for _, part := range parts {
		cm, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = cm[part]
		if !ok {
			return nil
		}
	}
	return current
}

// extractAllValues traverses a map using a dot-separated path that may contain
// "[*]" segments for array iteration, returning all leaf values found.
//
// Supported path syntax:
//
//	"properties.sku.name"                               → single scalar
//	"properties.networkAcls.ipRules[*].value"           → each ipRule's value
//	"properties.accessPolicies[*].permissions.keys[*]"  → each key permission string
//
// A segment ending in "[*]" (e.g. "ipRules[*]") means: look up the field name
// before "[*]" in the current map, treat the result as []interface{}, and recurse
// into each element with the remaining path segments.
func extractAllValues(m map[string]interface{}, path string) []interface{} {
	return extractValuesRecursive(m, strings.Split(path, "."))
}

func extractValuesRecursive(current interface{}, parts []string) []interface{} {
	if len(parts) == 0 {
		return []interface{}{current}
	}

	part := parts[0]
	rest := parts[1:]

	if strings.HasSuffix(part, "[*]") {
		fieldName := part[:len(part)-3] // strip "[*]"
		cm, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		arr, ok := cm[fieldName].([]interface{})
		if !ok {
			return nil
		}
		var results []interface{}
		for _, elem := range arr {
			results = append(results, extractValuesRecursive(elem, rest)...)
		}
		return results
	}

	// Regular field traversal.
	cm, ok := current.(map[string]interface{})
	if !ok {
		return nil
	}
	val, ok := cm[part]
	if !ok {
		return nil
	}
	return extractValuesRecursive(val, rest)
}

// extractArray returns the []interface{} at the given dot-separated path, or nil.
func extractArray(m map[string]interface{}, path string) []interface{} {
	v := extractNestedValue(m, path)
	if arr, ok := v.([]interface{}); ok {
		return arr
	}
	return nil
}

// extractStringValue is a convenience wrapper that returns the value at path as a string.
// Returns "" if the path doesn't exist or the value is not a string.
func extractStringValue(m map[string]interface{}, path string) string {
	v := extractNestedValue(m, path)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// toFloat64 attempts to convert a JSON number (which Go's json.Unmarshal represents as float64)
// to float64. Returns 0, false if the value is not numeric.
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// ptr returns a pointer to v.
func ptr[T any](v T) *T {
	return &v
}

// removeNestedField deletes the leaf key at a dot-separated path from a nested map.
// Intermediate segments that are not maps are silently skipped.
func removeNestedField(m map[string]interface{}, dotPath string) {
	parts := strings.Split(dotPath, ".")
	target := m
	for i := 0; i < len(parts)-1; i++ {
		next, ok := target[parts[i]].(map[string]interface{})
		if !ok {
			return
		}
		target = next
	}
	delete(target, parts[len(parts)-1])
}
