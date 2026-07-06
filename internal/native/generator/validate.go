package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// ValidateEmittedSchema verifies that the emitted Go source faithfully covers
// the bicep type graph. It works directly with the type graph and the emitted
// source string — no compiled schema.Schema needed.
//
// This is designed to run inside the generator pipeline itself, immediately
// after EmitSchema, so every generation run validates automatically.
//
// It is presence-only: because the source has not been compiled, it cannot
// inspect an attribute's Terraform type. Type-mapping correctness is checked by
// the compiled-schema adapter (validate.SchemaAgainstBicep, ADR-0008).
//
// Checks:
//   - Every non-SystemManaged bicep property has a corresponding attribute in the emitted source
//   - Every quoted attribute name in the emitted source exists in the bicep type graph
func ValidateEmittedSchema(emittedSource string, body *typegraph.Type, ignoreTopLevel ...string) []typegraph.Mismatch {
	if body == nil || body.Kind != typegraph.KindObject {
		return []typegraph.Mismatch{{
			Kind:   typegraph.MismatchTypeMismatch,
			Detail: "bicep body is not an ObjectType",
		}}
	}

	// Synthesized envelope attributes (name / parent reference / id) are not part
	// of the bicep body graph; exclude them from the "extra in schema" check.
	ignore := make(map[string]bool, len(ignoreTopLevel))
	for _, n := range ignoreTopLevel {
		ignore[n] = true
	}

	// Collect all expected attribute paths from the type graph
	// (same traversal the emitter uses)
	expectedPaths := make(map[string]*typegraph.Property) // snake_case dotted path → bicep property
	CollectExpectedPaths(body, "", expectedPaths)

	// Extract all attribute paths from the emitted Go source
	emittedPaths := extractEmittedPaths(emittedSource)

	var mismatches []typegraph.Mismatch

	// Check: every emitted path should exist in expected
	for _, path := range sortedStringSet(emittedPaths) {
		if ignore[path] {
			continue
		}
		if _, ok := expectedPaths[path]; !ok {
			mismatches = append(mismatches, typegraph.Mismatch{
				Path:   path,
				Kind:   typegraph.MismatchExtraInSchema,
				Detail: fmt.Sprintf("emitted attribute %q not found in bicep type graph", path),
			})
		}
	}

	// Check: every expected path should exist in emitted
	for _, path := range sortedStringSet(expectedPaths) {
		if !emittedPaths[path] {
			prop := expectedPaths[path]
			mismatches = append(mismatches, typegraph.Mismatch{
				Path:   path,
				ARM:    prop.Name,
				Kind:   typegraph.MismatchMissingInSchema,
				Detail: fmt.Sprintf("bicep property %q (path: %s) not emitted in schema", prop.Name, path),
			})
		}
	}

	return mismatches
}

// CollectExpectedPaths walks the type graph and collects all attribute paths
// that the emitter should produce, using the shared verification leaf rule
// (typegraph.SchemaAttrName): skip SystemManaged, convert names with
// CamelToSnake, recurse into objects / arrays-of-objects / maps-of-objects.
func CollectExpectedPaths(typ *typegraph.Type, prefix string, out map[string]*typegraph.Property) {
	if typ == nil || typ.Kind != typegraph.KindObject {
		return
	}
	for armName, prop := range typ.Properties {
		tfName, skip := typegraph.SchemaAttrName(armName, prop)
		if skip {
			continue
		}
		path := typegraph.JoinPath(prefix, tfName)
		out[path] = prop

		if child := typegraph.RecurseChild(prop.Type); child != nil {
			CollectExpectedPaths(child, path, out)
		}
	}
}

// attrNameInSource matches quoted attribute names in schema declarations:
//
//	"attribute_name": schema.XxxAttribute{
var attrNameInSource = regexp.MustCompile(`"([a-z][a-z0-9_]*)"\s*:\s*schema\.\w+Attribute`)

// managedIdentityInSource matches the reusable native schema helper emitted for
// the common ARM identity envelope. The helper expands to the same fixed child
// schema as the bicep ManagedServiceIdentity shape, so validation records its
// known paths explicitly.
var managedIdentityInSource = regexp.MustCompile(`"([a-z][a-z0-9_]*)"\s*:\s*nativeschema\.ManagedServiceIdentity\(`)

// extractEmittedPaths parses the emitted Go source and extracts all attribute
// paths by tracking the nesting of quoted attribute names in schema declarations.
func extractEmittedPaths(source string) map[string]bool {
	paths := make(map[string]bool)

	// Track nesting from gofmt indentation instead of brace depth. Attribute
	// blocks contain validators, plan modifiers, and descriptions whose braces do
	// not define schema hierarchy; indentation does.
	type scope struct {
		path   string
		indent int
	}

	var stack []scope

	lines := strings.Split(source, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := managedIdentityInSource.FindStringSubmatch(trimmed); m != nil {
			attrName := m[1]
			indent := len(line) - len(strings.TrimLeft(line, "\t "))
			for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
				stack = stack[:len(stack)-1]
			}
			var parentPath string
			if len(stack) > 0 {
				parentPath = stack[len(stack)-1].path
			}
			fullPath := typegraph.JoinPath(parentPath, attrName)
			for _, suffix := range []string{
				"",
				"principal_id",
				"tenant_id",
				"type",
				"user_assigned_identities",
				"user_assigned_identities.client_id",
				"user_assigned_identities.principal_id",
			} {
				paths[joinPathAllowEmpty(fullPath, suffix)] = true
			}
			continue
		}

		m := attrNameInSource.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		attrName := m[1]
		indent := len(line) - len(strings.TrimLeft(line, "\t "))
		for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}

		var parentPath string
		if len(stack) > 0 {
			parentPath = stack[len(stack)-1].path
		}
		fullPath := typegraph.JoinPath(parentPath, attrName)
		paths[fullPath] = true

		if strings.Contains(trimmed, "schema.SingleNestedAttribute{") || strings.Contains(trimmed, "schema.ListNestedAttribute{") || strings.Contains(trimmed, "schema.MapNestedAttribute{") {
			stack = append(stack, scope{path: fullPath, indent: indent})
		}
	}

	return paths
}

func joinPathAllowEmpty(prefix, name string) string {
	if name == "" {
		return prefix
	}
	return typegraph.JoinPath(prefix, name)
}

func sortedStringSet[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
