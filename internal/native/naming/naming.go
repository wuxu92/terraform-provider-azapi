// Package naming converts ARM resource types and property names to Terraform-compatible
// identifiers. Terraform SDK enforces ^[a-z_][a-z0-9_]*$ for attribute names.
package naming

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// resourceNameOverrides pins the Terraform resource name for ARM types whose
// mechanical service+segment derivation is redundant or would collide. Keyed by
// the bare ARM type (no @version). This is the naming override table the design
// calls for (DEVELOPER_SPEC / GENERATOR.md): ResourceName consults it first, so a
// single entry propagates to the descriptor name, the generated file name, and the
// emitted Go identifiers. Keep it tiny and deliberate — the default is the
// mechanical, predictable conversion.
var resourceNameOverrides = map[string]string{
	// The service token "managedidentity" already carries "identity", so the
	// mechanical join stutters ("managedidentity_user_assigned_identity"). AzureRM
	// and users know this resource simply as the user-assigned identity.
	"Microsoft.ManagedIdentity/userAssignedIdentities": "azapi_user_assigned_identity",
}

// ResourceName converts an ARM resource type (e.g., "Microsoft.Storage/storageAccounts")
// to a Terraform resource name (e.g., "azapi_storage_account").
func ResourceName(armType string) string {
	if override, ok := resourceNameOverrides[armType]; ok {
		return override
	}
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

	// Flatten the service word and every segment word into one token list, then
	// drop any token a later token repeats (singular/plural-insensitive), keeping
	// the last occurrence. This removes the stutter a mechanical service-prefix +
	// segment join produces — "storage"+"storage_account" -> "storage_account",
	// "resources"+"resource_group" -> "resource_group", "network"+"virtual_network"
	// -> "virtual_network" — without dropping a word the name still needs.
	tokens := []string{service}
	for _, seg := range segments {
		tokens = append(tokens, strings.Split(seg, "_")...)
	}
	return "azapi_" + strings.Join(dedupRepeatedWords(tokens), "_")
}

// dedupRepeatedWords returns tokens with repeated word-classes collapsed to their
// last occurrence, comparing singular and plural forms as equal. Order of the
// survivors is preserved, so ["resources","resource","group"] becomes
// ["resource","group"] and ["network","virtual","network"] becomes
// ["virtual","network"].
func dedupRepeatedWords(tokens []string) []string {
	last := make(map[string]int, len(tokens))
	for i, t := range tokens {
		last[singularize(t)] = i
	}
	out := make([]string, 0, len(tokens))
	for i, t := range tokens {
		if last[singularize(t)] == i {
			out = append(out, t)
		}
	}
	return out
}

// ServiceName returns the service folder/package name for an ARM resource type:
// the namespace's service segment, lowercased. "Microsoft.Storage/storageAccounts"
// -> "storage", "Dynatrace.Observability/monitors" -> "dynatrace". Generated
// resources are grouped by service (services/<service>/...).
func ServiceName(armType string) string {
	namespace := armType
	if i := strings.Index(armType, "/"); i >= 0 {
		namespace = armType[:i]
	}
	return extractService(namespace)
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
	"dns":        true,
	"kubernetes": true,
	"status":     true,
	"access":     true,
	"bus":        true,
	"address":    true,
	"atlas":      true,
	"cosmos":     true,
	"analysis":   true,
	"insights":   true,
	"series":     true,
}

// irregularPlurals maps irregular plural forms to their singular
var irregularPlurals = map[string]string{
	"databases":   "database",
	"namespaces":  "namespace",
	"addresses":   "address",
	"statuses":    "status",
	"serverfarms": "server_farm",
	"indices":     "index",
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

// ---------------------------------------------------------------------------
// Parent-reference derivation
// ---------------------------------------------------------------------------

// bicep ResourceType.writableScopes bitmask.
const (
	ScopeTenant          = 1
	ScopeManagementGroup = 2
	ScopeSubscription    = 4
	ScopeResourceGroup   = 8
	ScopeExtension       = 16
)

// ParentRef describes the attribute a native resource uses to reference its
// parent: the Terraform attribute name plus the ID shape that name must satisfy.
// Pattern is empty when the parent shape can't be pinned to a single scope
// (multi-scope / extension / tenant / unknown), in which case Name falls back to
// the generic "parent_id" and the attribute is left unconstrained.
type ParentRef struct {
	Name    string // e.g. "resource_group_id", "storage_account_id", "parent_id"
	Pattern string // validator regex; empty == unconstrained
	Message string // validator failure message
}

// ParentReference derives the parent-reference attribute for an ARM resource type
// from its type path and writable scope. A child resource (multi-segment type)
// references its parent ARM type by a descriptive name
// ("Microsoft.Storage/storageAccounts/blobServices" -> "storage_account_id"); a
// top-level resource references its writable scope ("resource_group_id",
// "subscription_id", "management_group_id"). When the scope is ambiguous the name
// falls back to "parent_id" with no validator.
func ParentReference(armType string, writableScopes int) ParentRef {
	if parent := parentARMType(armType); parent != armType {
		name := TypeReferenceName(parent) + "_id"
		return ParentRef{
			Name:    name,
			Pattern: childIDPattern(parent),
			Message: fmt.Sprintf("%s must be the ID of a %s resource", name, parent),
		}
	}
	switch writableScopes {
	case ScopeResourceGroup:
		return ParentRef{
			Name:    "resource_group_id",
			Pattern: `(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+$`,
			Message: "resource_group_id must be a resource group ID (/subscriptions/{id}/resourceGroups/{name})",
		}
	case ScopeSubscription:
		return ParentRef{
			Name:    "subscription_id",
			Pattern: `(?i)^/subscriptions/[^/]+$`,
			Message: "subscription_id must be a subscription ID (/subscriptions/{id})",
		}
	case ScopeManagementGroup:
		return ParentRef{
			Name:    "management_group_id",
			Pattern: `(?i)^/providers/Microsoft\.Management/managementGroups/[^/]+$`,
			Message: "management_group_id must be a management group ID",
		}
	}
	return ParentRef{Name: "parent_id"}
}

// TypeReferenceName converts an ARM resource type to the singular snake_case noun
// used to reference it from a child resource
// ("Microsoft.Storage/storageAccounts" -> "storage_account"). Unlike ResourceName
// it keeps the full last type segment (no service-prefix stripping) so the result
// reads unambiguously as a parent-id attribute name.
func TypeReferenceName(armType string) string {
	parts := strings.Split(armType, "/")
	last := strings.ReplaceAll(parts[len(parts)-1], "-", "_")
	return singularize(CamelToSnake(last))
}

// parentARMType strips the trailing type segment from a multi-segment ARM type
// ("Microsoft.Storage/storageAccounts/blobServices" ->
// "Microsoft.Storage/storageAccounts"). A top-level type (namespace + one
// segment) is returned unchanged.
func parentARMType(armType string) string {
	parts := strings.Split(armType, "/")
	if len(parts) <= 2 {
		return armType
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// childIDPattern builds the regex an ARM resource ID of parentType must satisfy.
// ARM IDs interleave each type segment with an instance name, so the pattern
// alternates a literal type segment with a "/[^/]+" name placeholder:
//
//	"Microsoft.Storage/storageAccounts"            -> /providers/Microsoft\.Storage/storageAccounts/[^/]+$
//	"Microsoft.ApiManagement/service/apis"         -> /providers/Microsoft\.ApiManagement/service/[^/]+/apis/[^/]+$
//
// Anchored only at the end so the subscription/resource-group prefix may precede it.
func childIDPattern(parentType string) string {
	parts := strings.Split(parentType, "/")
	var b strings.Builder
	b.WriteString(`(?i)/providers/`)
	b.WriteString(regexp.QuoteMeta(parts[0])) // namespace, e.g. Microsoft.ApiManagement
	for _, seg := range parts[1:] {           // each type segment gets an instance name
		b.WriteString(`/`)
		b.WriteString(regexp.QuoteMeta(seg))
		b.WriteString(`/[^/]+`)
	}
	b.WriteString(`$`)
	return b.String()
}
