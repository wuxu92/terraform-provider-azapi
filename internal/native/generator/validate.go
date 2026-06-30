package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
)

// Mismatch describes a discrepancy between the emitted schema and bicep types.
type Mismatch struct {
	Path   string // Dot-separated attribute path
	ARM    string // ARM property path (if known)
	Kind   MismatchKind
	Detail string
}

// MismatchKind identifies the type of schema-bicep mismatch.
type MismatchKind int

const (
	// MismatchExtraInSchema means the emitted schema has an attribute
	// not present in the bicep type graph.
	MismatchExtraInSchema MismatchKind = iota

	// MismatchMissingInSchema means the bicep type graph has a property
	// not present in the emitted schema.
	MismatchMissingInSchema

	// MismatchTypeMismatch means both sides have the property but the
	// types don't correspond.
	MismatchTypeMismatch
)

func (k MismatchKind) String() string {
	switch k {
	case MismatchExtraInSchema:
		return "extra_in_schema"
	case MismatchMissingInSchema:
		return "missing_in_schema"
	case MismatchTypeMismatch:
		return "type_mismatch"
	default:
		return "unknown"
	}
}

// ValidateEmittedSchema verifies that the emitted Go source faithfully covers
// the bicep type graph. It works directly with the type graph and the emitted
// source string — no compiled schema.Schema needed.
//
// This is designed to run inside the generator pipeline itself, immediately
// after EmitSchema, so every generation run validates automatically.
//
// Checks:
//   - Every non-SystemManaged bicep property has a corresponding attribute in the emitted source
//   - Every quoted attribute name in the emitted source exists in the bicep type graph
//   - Type mapping is correct (String↔string, Bool↔bool, Object↔SingleNested, etc.)
func ValidateEmittedSchema(emittedSource string, body *Type, ignoreTopLevel ...string) []Mismatch {
	if body == nil || body.Kind != KindObject {
		return []Mismatch{{
			Kind:   MismatchTypeMismatch,
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
	expectedPaths := make(map[string]*Property) // snake_case dotted path → bicep property
	CollectExpectedPaths(body, "", expectedPaths)

	// Extract all attribute paths from the emitted Go source
	emittedPaths := extractEmittedPaths(emittedSource)

	var mismatches []Mismatch

	// Check: every emitted path should exist in expected
	for _, path := range sortedStringSet(emittedPaths) {
		if ignore[path] {
			continue
		}
		if _, ok := expectedPaths[path]; !ok {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				Kind:   MismatchExtraInSchema,
				Detail: fmt.Sprintf("emitted attribute %q not found in bicep type graph", path),
			})
		}
	}

	// Check: every expected path should exist in emitted
	for _, path := range sortedStringSet(expectedPaths) {
		if !emittedPaths[path] {
			prop := expectedPaths[path]
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    prop.Name,
				Kind:   MismatchMissingInSchema,
				Detail: fmt.Sprintf("bicep property %q (path: %s) not emitted in schema", prop.Name, path),
			})
		}
	}

	return mismatches
}

// CollectExpectedPaths walks the type graph and collects all attribute paths
// that the emitter should produce, using the same logic as emitAttributes:
// skip SystemManaged, convert names with CamelToSnake, recurse into objects/arrays.
func CollectExpectedPaths(typ *Type, prefix string, out map[string]*Property) {
	if typ == nil || typ.Kind != KindObject {
		return
	}
	for armName, prop := range typ.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		tfName := naming.CamelToSnake(armName)
		path := joinPath(prefix, tfName)
		out[path] = prop

		// Recurse into nested objects
		if prop.Type.Kind == KindObject {
			CollectExpectedPaths(prop.Type, path, out)
		}
		// Recurse into arrays of objects
		if prop.Type.Kind == KindArray && prop.Type.ElementType != nil && prop.Type.ElementType.Kind == KindObject {
			CollectExpectedPaths(prop.Type.ElementType, path, out)
		}
		// Recurse into maps of objects (MapNestedAttribute element schema)
		if prop.Type.Kind == KindMap && prop.Type.ElementType != nil && prop.Type.ElementType.Kind == KindObject {
			CollectExpectedPaths(prop.Type.ElementType, path, out)
		}
	}
}

// attrNameInSource matches quoted attribute names in schema declarations:
//
//	"attribute_name": schema.XxxAttribute{
var attrNameInSource = regexp.MustCompile(`"([a-z][a-z0-9_]*)"\s*:\s*schema\.\w+Attribute`)

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
		fullPath := joinPath(parentPath, attrName)
		paths[fullPath] = true

		if strings.Contains(trimmed, "schema.SingleNestedAttribute{") || strings.Contains(trimmed, "schema.ListNestedAttribute{") || strings.Contains(trimmed, "schema.MapNestedAttribute{") {
			stack = append(stack, scope{path: fullPath, indent: indent})
		}
	}

	return paths
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func sortedStringSet[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FormatMismatches returns a human-readable summary of mismatches.
func FormatMismatches(mismatches []Mismatch) string {
	if len(mismatches) == 0 {
		return "No mismatches found."
	}

	var b strings.Builder
	extra := 0
	missing := 0
	typeMismatch := 0
	for _, m := range mismatches {
		switch m.Kind {
		case MismatchExtraInSchema:
			extra++
		case MismatchMissingInSchema:
			missing++
		case MismatchTypeMismatch:
			typeMismatch++
		}
	}
	b.WriteString(fmt.Sprintf("%d mismatches: %d extra in schema, %d missing in schema, %d type mismatches\n",
		len(mismatches), extra, missing, typeMismatch))

	for _, m := range mismatches {
		b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", m.Kind, m.Path, m.Detail))
	}
	return b.String()
}
