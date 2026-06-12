package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/naming"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Mismatch describes a discrepancy between the Terraform schema and bicep types.
type Mismatch struct {
	Path     string // Dot-separated Terraform attribute path
	ARM      string // Corresponding ARM path (if known)
	Kind     MismatchKind
	Detail   string
}

// MismatchKind identifies the type of schema-bicep mismatch.
type MismatchKind int

const (
	// MismatchExtraInSchema means the Terraform schema has an attribute
	// not present in the bicep type graph.
	MismatchExtraInSchema MismatchKind = iota

	// MismatchMissingInSchema means the bicep type graph has a property
	// not present in the Terraform schema.
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

// ValidateSchemaAgainstBicep cross-references a generated Terraform schema
// against the bicep type graph it was generated from. Returns all mismatches.
//
// The validator walks both trees in parallel, matching Terraform snake_case
// attribute names to ARM camelCase property names via naming.CamelToSnake.
func ValidateSchemaAgainstBicep(s schema.Schema, body *Type) []Mismatch {
	if body == nil || body.Kind != KindObject {
		return []Mismatch{{
			Path:   "",
			Kind:   MismatchTypeMismatch,
			Detail: "bicep body is not an ObjectType",
		}}
	}
	return validateObject(s.Attributes, body, "")
}

func validateObject(tfAttrs map[string]schema.Attribute, bicepObj *Type, prefix string) []Mismatch {
	var mismatches []Mismatch

	// Build a map from snake_case → ARM name for all bicep properties
	armBySnake := make(map[string]string)   // snake_case → ARM camelCase
	bicepBySnake := make(map[string]*Property) // snake_case → bicep property
	for armName, prop := range bicepObj.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		snake := naming.CamelToSnake(armName)
		armBySnake[snake] = armName
		bicepBySnake[snake] = prop
	}

	// Check: every Terraform attribute should exist in bicep
	tfNames := sortedKeys(tfAttrs)
	for _, tfName := range tfNames {
		tfAttr := tfAttrs[tfName]
		path := joinPath(prefix, tfName)

		bicepProp, ok := bicepBySnake[tfName]
		if !ok {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				Kind:   MismatchExtraInSchema,
				Detail: fmt.Sprintf("attribute %q exists in Terraform schema but not in bicep types", tfName),
			})
			continue
		}

		// Type correspondence check + recurse into nested
		mismatches = append(mismatches, validateTypeCorrespondence(tfAttr, bicepProp, path)...)

		// Remove from the map so we can find what's left (missing in schema)
		delete(bicepBySnake, tfName)
	}

	// Check: every non-SystemManaged bicep property should exist in schema
	missingNames := sortedMapKeys(bicepBySnake)
	for _, snake := range missingNames {
		prop := bicepBySnake[snake]
		path := joinPath(prefix, snake)
		mismatches = append(mismatches, Mismatch{
			Path:   path,
			ARM:    prop.Name,
			Kind:   MismatchMissingInSchema,
			Detail: fmt.Sprintf("bicep property %q (snake: %q) exists in types but not in Terraform schema", prop.Name, snake),
		})
	}

	return mismatches
}

func validateTypeCorrespondence(tfAttr schema.Attribute, bicepProp *Property, path string) []Mismatch {
	var mismatches []Mismatch
	bicepType := bicepProp.Type

	switch attr := tfAttr.(type) {
	case schema.StringAttribute:
		if !isStringCompatible(bicepType) {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    bicepProp.Name,
				Kind:   MismatchTypeMismatch,
				Detail: fmt.Sprintf("Terraform StringAttribute but bicep type is %s", bicepTypeDesc(bicepType)),
			})
		}

	case schema.BoolAttribute:
		if bicepType.Kind != KindBool {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    bicepProp.Name,
				Kind:   MismatchTypeMismatch,
				Detail: fmt.Sprintf("Terraform BoolAttribute but bicep type is %s", bicepTypeDesc(bicepType)),
			})
		}

	case schema.Int64Attribute:
		if bicepType.Kind != KindInt {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    bicepProp.Name,
				Kind:   MismatchTypeMismatch,
				Detail: fmt.Sprintf("Terraform Int64Attribute but bicep type is %s", bicepTypeDesc(bicepType)),
			})
		}

	case schema.SingleNestedAttribute:
		if bicepType.Kind != KindObject {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    bicepProp.Name,
				Kind:   MismatchTypeMismatch,
				Detail: fmt.Sprintf("Terraform SingleNestedAttribute but bicep type is %s", bicepTypeDesc(bicepType)),
			})
		} else {
			// Recurse into nested attributes
			mismatches = append(mismatches, validateObject(attr.Attributes, bicepType, path)...)
		}

	case schema.ListNestedAttribute:
		if bicepType.Kind != KindArray {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    bicepProp.Name,
				Kind:   MismatchTypeMismatch,
				Detail: fmt.Sprintf("Terraform ListNestedAttribute but bicep type is %s", bicepTypeDesc(bicepType)),
			})
		} else if bicepType.ElementType != nil && bicepType.ElementType.Kind == KindObject {
			// Recurse into element attributes
			mismatches = append(mismatches, validateObject(attr.NestedObject.Attributes, bicepType.ElementType, path+"[*]")...)
		}

	case schema.ListAttribute:
		if bicepType.Kind != KindArray {
			mismatches = append(mismatches, Mismatch{
				Path:   path,
				ARM:    bicepProp.Name,
				Kind:   MismatchTypeMismatch,
				Detail: fmt.Sprintf("Terraform ListAttribute but bicep type is %s", bicepTypeDesc(bicepType)),
			})
		}

	case schema.DynamicAttribute:
		// Dynamic accepts anything — always compatible

	default:
		// Unknown attribute type — skip (could be Float64Attribute, etc.)
	}

	return mismatches
}

// isStringCompatible returns true if a bicep type can be represented as a Terraform string.
func isStringCompatible(t *Type) bool {
	switch t.Kind {
	case KindString, KindStringLiteral:
		return true
	case KindUnion:
		// Enum (union of string literals) → string
		return t.IsEnum()
	default:
		return false
	}
}

func bicepTypeDesc(t *Type) string {
	switch t.Kind {
	case KindString:
		return "String"
	case KindStringLiteral:
		return "StringLiteral(" + t.Name + ")"
	case KindBool:
		return "Bool"
	case KindInt:
		return "Int"
	case KindObject:
		return "Object(" + t.Name + ")"
	case KindArray:
		if t.ElementType != nil {
			return "Array(" + bicepTypeDesc(t.ElementType) + ")"
		}
		return "Array"
	case KindUnion:
		if t.IsEnum() {
			return "Enum"
		}
		return "Union"
	case KindAny:
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

func sortedKeys(m map[string]schema.Attribute) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapKeys(m map[string]*Property) []string {
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
