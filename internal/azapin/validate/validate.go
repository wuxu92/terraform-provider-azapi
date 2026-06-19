// Package validate provides a compiled-schema validator that cross-references
// generated Terraform schema.Schema objects against the bicep type graph.
// It reuses generator.Type — no duplicate walker.
package validate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/generator"
	"github.com/Azure/terraform-provider-azapi/internal/azapin/naming"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Mismatch describes a discrepancy between a compiled schema and bicep types.
type Mismatch struct {
	Path   string
	ARM    string
	Kind   MismatchKind
	Detail string
}

// MismatchKind identifies the type of mismatch.
type MismatchKind int

const (
	MismatchExtraInSchema   MismatchKind = iota // Attribute in schema, not in bicep
	MismatchMissingInSchema                     // Property in bicep, not in schema
	MismatchTypeMismatch                        // Both exist, types don't correspond
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

// AzapinTagRegex extracts the resource type+version from a schema Description.
// Format: [azapin:Microsoft.Storage/storageAccounts@2025-01-01]
var AzapinTagRegex = regexp.MustCompile(`\[azapin:([^\]]+)\]`)

// ExtractResourceTag extracts the "ResourceType@Version" from a schema Description.
func ExtractResourceTag(description string) (string, bool) {
	m := AzapinTagRegex.FindStringSubmatch(description)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

// SchemaAgainstBicep validates a compiled schema.Schema against a bicep type tree.
// It walks both trees in parallel, matching snake_case Terraform attributes to
// camelCase ARM properties via naming.CamelToSnake. ignoreTopLevel lists
// top-level attribute names to exclude — the synthesized operational-envelope
// attributes (name / parent reference / id) are not part of the bicep body.
func SchemaAgainstBicep(s schema.Schema, body *generator.Type, ignoreTopLevel ...string) []Mismatch {
	if body == nil || body.Kind != generator.KindObject {
		return []Mismatch{{Kind: MismatchTypeMismatch, Detail: "bicep body is not an ObjectType"}}
	}
	attrs := s.Attributes
	if len(ignoreTopLevel) > 0 {
		ignore := make(map[string]bool, len(ignoreTopLevel))
		for _, n := range ignoreTopLevel {
			ignore[n] = true
		}
		attrs = make(map[string]schema.Attribute, len(s.Attributes))
		for k, v := range s.Attributes {
			if !ignore[k] {
				attrs[k] = v
			}
		}
	}
	return validateObject(attrs, body, "")
}

func validateObject(tfAttrs map[string]schema.Attribute, bicepObj *generator.Type, prefix string) []Mismatch {
	var mismatches []Mismatch

	// Build snake_case → bicep property map
	bicepBySnake := make(map[string]*generator.Property)
	for armName, prop := range bicepObj.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		snake := naming.CamelToSnake(armName)
		bicepBySnake[snake] = prop
	}

	// Check: every TF attribute exists in bicep
	for _, tfName := range sortedAttrKeys(tfAttrs) {
		tfAttr := tfAttrs[tfName]
		path := joinPath(prefix, tfName)

		bicepProp, ok := bicepBySnake[tfName]
		if !ok {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				Kind:   MismatchExtraInSchema,
				Detail: fmt.Sprintf("attribute %q in compiled schema but not in bicep types", tfName),
			})
			continue
		}

		mismatches = append(mismatches, validateType(tfAttr, bicepProp, path)...)
		delete(bicepBySnake, tfName)
	}

	// Check: every bicep property exists in TF schema
	for _, snake := range sortedPropKeys(bicepBySnake) {
		prop := bicepBySnake[snake]
		path := joinPath(prefix, snake)
		mismatches = append(mismatches, Mismatch{
			Path:   path,
			ARM:    prop.Name,
			Kind:   MismatchMissingInSchema,
			Detail: fmt.Sprintf("bicep property %q not in compiled schema", prop.Name),
		})
	}

	return mismatches
}

func validateType(tfAttr schema.Attribute, bicepProp *generator.Property, path string) []Mismatch {
	var mismatches []Mismatch
	bt := bicepProp.Type

	switch attr := tfAttr.(type) {
	case schema.StringAttribute:
		if !isStringCompat(bt) {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "StringAttribute", bt))
		}
	case schema.BoolAttribute:
		if bt.Kind != generator.KindBool {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "BoolAttribute", bt))
		}
	case schema.Int64Attribute:
		if bt.Kind != generator.KindInt {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "Int64Attribute", bt))
		}
	case schema.SingleNestedAttribute:
		if bt.Kind != generator.KindObject {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "SingleNestedAttribute", bt))
		} else {
			mismatches = append(mismatches, validateObject(attr.Attributes, bt, path)...)
		}
	case schema.ListNestedAttribute:
		if bt.Kind != generator.KindArray {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "ListNestedAttribute", bt))
		} else if bt.ElementType != nil && bt.ElementType.Kind == generator.KindObject {
			mismatches = append(mismatches, validateObject(attr.NestedObject.Attributes, bt.ElementType, path+"[*]")...)
		}
	case schema.ListAttribute:
		if bt.Kind != generator.KindArray {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "ListAttribute", bt))
		}
	case schema.DynamicAttribute:
		// Dynamic accepts anything
	}

	return mismatches
}

func isStringCompat(t *generator.Type) bool {
	switch t.Kind {
	case generator.KindString, generator.KindStringLiteral:
		return true
	case generator.KindUnion:
		return t.IsEnum()
	default:
		return false
	}
}

func typeMismatch(path string, prop *generator.Property, tfType string, bt *generator.Type) Mismatch {
	return Mismatch{
		Path:   path,
		ARM:    prop.Name,
		Kind:   MismatchTypeMismatch,
		Detail: fmt.Sprintf("compiled %s but bicep type is %s", tfType, bicepDesc(bt)),
	}
}

func bicepDesc(t *generator.Type) string {
	switch t.Kind {
	case generator.KindString:
		return "String"
	case generator.KindStringLiteral:
		return "StringLiteral"
	case generator.KindBool:
		return "Bool"
	case generator.KindInt:
		return "Int"
	case generator.KindObject:
		return "Object"
	case generator.KindArray:
		return "Array"
	case generator.KindUnion:
		if t.IsEnum() {
			return "Enum"
		}
		return "Union"
	case generator.KindAny:
		return "Any"
	default:
		return fmt.Sprintf("Unknown(%d)", t.Kind)
	}
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func sortedAttrKeys(m map[string]schema.Attribute) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedPropKeys(m map[string]*generator.Property) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FormatMismatches returns a human-readable summary.
func FormatMismatches(mismatches []Mismatch) string {
	if len(mismatches) == 0 {
		return "No mismatches found."
	}
	var b strings.Builder
	extra, missing, typeMm := 0, 0, 0
	for _, m := range mismatches {
		switch m.Kind {
		case MismatchExtraInSchema:
			extra++
		case MismatchMissingInSchema:
			missing++
		case MismatchTypeMismatch:
			typeMm++
		}
	}
	b.WriteString(fmt.Sprintf("%d mismatches: %d extra, %d missing, %d type\n",
		len(mismatches), extra, missing, typeMm))
	for _, m := range mismatches {
		b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", m.Kind, m.Path, m.Detail))
	}
	return b.String()
}
