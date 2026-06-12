// Package naming converts ARM resource types and property names to Terraform-compatible
// identifiers. Terraform SDK enforces ^[a-z_][a-z0-9_]*$ for attribute names.
package naming

import (
	"regexp"
	"strings"
	"unicode"
)

// ResourceName converts an ARM resource type (e.g., "Microsoft.Storage/storageAccounts")
// to a Terraform resource name (e.g., "azapi_storage_account").
func ResourceName(armType string) string {
	parts := strings.Split(armType, "/")
	if len(parts) < 2 {
		return "azapi_" + CamelToSnake(armType)
	}

	// Extract service from namespace: Microsoft.Storage → storage
	namespace := parts[0]
	service := extractService(namespace)

	// Convert each resource segment
	segments := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		// Skip segments that look like namespaces (contain dots) — these are
		// sub-resource type qualifiers, not resource names
		if strings.Contains(part, ".") {
			continue
		}
		// Replace hyphens with underscores before camelCase conversion
		part = strings.ReplaceAll(part, "-", "_")
		snake := CamelToSnake(part)
		snake = singularize(snake)
		segments = append(segments, snake)
	}

	// Strip redundant service prefix from first segment
	// e.g., storage/storageAccounts → storage + storage_account → strip to account
	if len(segments) > 0 && strings.HasPrefix(segments[0], service+"_") {
		segments[0] = segments[0][len(service)+1:]
	}

	resourceName := strings.Join(segments, "_")
	return "azapi_" + service + "_" + resourceName
}

// CamelToSnake converts a camelCase or PascalCase string to snake_case.
// Handles acronyms: "isHnsEnabled" → "is_hns_enabled", "IPRules" → "ip_rules".
func CamelToSnake(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s) + 4) // pre-allocate slightly larger

	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				// Insert underscore when:
				// - previous is lowercase: "minT" → "min_t"
				// - previous is digit: "V3E" → "v3_e"
				// - previous is uppercase but next is lowercase (end of acronym): "HNSe" → "hns_e"
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					result.WriteByte('_')
				} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					result.WriteByte('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SnakeToCamel converts a snake_case string back to camelCase.
// This is the inverse of CamelToSnake for simple cases.
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}
	var result strings.Builder
	result.WriteString(parts[0])
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		result.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return result.String()
}

// extractService gets the service name from an ARM namespace.
// "Microsoft.Storage" → "storage", "Dynatrace.Observability" → "dynatrace"
func extractService(namespace string) string {
	parts := strings.Split(strings.ToLower(namespace), ".")
	if len(parts) < 2 {
		return strings.ToLower(namespace)
	}
	// Use last segment unless it's a generic term
	service := parts[len(parts)-1]
	// For "Microsoft.X" namespaces, use X
	// For third-party like "Dynatrace.Observability", use the vendor (first part)
	if strings.ToLower(parts[0]) != "microsoft" {
		service = parts[0]
	}
	return service
}

// Common irregular plurals and words that should not be singularized
var noSingularize = map[string]bool{
	"redis":      true,
	"dns":       true,
	"kubernetes": true,
	"status":    true,
	"access":    true,
	"bus":       true,
	"address":   true,
	"atlas":     true,
	"cosmos":    true,
	"analysis":  true,
	"insights":  true,
	"series":    true,
}

// irregularPlurals maps irregular plural forms to their singular
var irregularPlurals = map[string]string{
	"databases":  "database",
	"namespaces": "namespace",
	"addresses":  "address",
	"statuses":   "status",
	"indices":    "index",
}

func singularize(s string) string {
	if noSingularize[s] {
		return s
	}
	if singular, ok := irregularPlurals[s]; ok {
		return singular
	}
	// Standard rules
	if strings.HasSuffix(s, "ies") && len(s) > 3 {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "ses") && !strings.HasSuffix(s, "sses") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") && len(s) > 1 {
		return s[:len(s)-1]
	}
	return s
}

// ValidTerraformName checks if a name matches the Terraform SDK requirement.
var validNameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// IsValidTerraformName returns true if the name is a valid Terraform attribute name.
func IsValidTerraformName(name string) bool {
	return validNameRegex.MatchString(name)
}
