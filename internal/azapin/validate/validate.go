// Package validate provides a compiled-schema validator that cross-references
// generated Terraform schema objects against their source bicep type graphs.
//
// This runs as go test, after compilation — it's the authoritative check that
// the generated schema.Schema objects correctly map to the ARM API payload.
package validate

import (
	"fmt"
	"sort"
	"strings"

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

// BicepType mirrors generator.Type but is decoupled from the generator package
// to avoid import cycles. The test constructs these from parsed types.json.
type BicepType struct {
	Kind        BicepTypeKind
	Name        string
	Properties  map[string]*BicepProperty
	ElementType *BicepType
	Elements    []*BicepType
}

// BicepProperty mirrors generator.Property.
type BicepProperty struct {
	Name  string // ARM JSON name (camelCase)
	Type  *BicepType
	Flags int // bicep flags: 1=Required, 2=ReadOnly, 4=WriteOnly, 8=SystemManaged
}

// BicepTypeKind mirrors generator.TypeKind.
type BicepTypeKind int

const (
	BicepKindString BicepTypeKind = iota
	BicepKindStringLiteral
	BicepKindInt
	BicepKindBool
	BicepKindObject
	BicepKindArray
	BicepKindUnion
	BicepKindAny
)

func (p *BicepProperty) isSystemManaged() bool { return p.Flags&8 != 0 }

func (t *BicepType) isEnum() bool {
	if t.Kind != BicepKindUnion {
		return false
	}
	for _, elem := range t.Elements {
		if elem.Kind != BicepKindStringLiteral && elem.Kind != BicepKindString {
			return false
		}
	}
	return true
}

// SchemaAgainstBicep validates a compiled schema.Schema against a bicep type tree.
// It walks both trees in parallel, matching snake_case Terraform attributes to
// camelCase ARM properties via naming.CamelToSnake.
func SchemaAgainstBicep(s schema.Schema, body *BicepType) []Mismatch {
	if body == nil || body.Kind != BicepKindObject {
		return []Mismatch{{Kind: MismatchTypeMismatch, Detail: "bicep body is not an ObjectType"}}
	}
	return validateObject(s.Attributes, body, "")
}

func validateObject(tfAttrs map[string]schema.Attribute, bicepObj *BicepType, prefix string) []Mismatch {
	var mismatches []Mismatch

	// Build snake_case → bicep property map
	bicepBySnake := make(map[string]*BicepProperty)
	for armName, prop := range bicepObj.Properties {
		if prop.isSystemManaged() {
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

func validateType(tfAttr schema.Attribute, bicepProp *BicepProperty, path string) []Mismatch {
	var mismatches []Mismatch
	bt := bicepProp.Type

	switch attr := tfAttr.(type) {
	case schema.StringAttribute:
		if !isStringCompat(bt) {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "StringAttribute", bt))
		}

	case schema.BoolAttribute:
		if bt.Kind != BicepKindBool {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "BoolAttribute", bt))
		}

	case schema.Int64Attribute:
		if bt.Kind != BicepKindInt {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "Int64Attribute", bt))
		}

	case schema.SingleNestedAttribute:
		if bt.Kind != BicepKindObject {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "SingleNestedAttribute", bt))
		} else {
			mismatches = append(mismatches, validateObject(attr.Attributes, bt, path)...)
		}

	case schema.ListNestedAttribute:
		if bt.Kind != BicepKindArray {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "ListNestedAttribute", bt))
		} else if bt.ElementType != nil && bt.ElementType.Kind == BicepKindObject {
			mismatches = append(mismatches, validateObject(attr.NestedObject.Attributes, bt.ElementType, path+"[*]")...)
		}

	case schema.ListAttribute:
		if bt.Kind != BicepKindArray {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "ListAttribute", bt))
		}

	case schema.DynamicAttribute:
		// Dynamic accepts anything

	default:
		// Float32/64, Set, Map, etc. — skip for now
	}

	return mismatches
}

func isStringCompat(t *BicepType) bool {
	switch t.Kind {
	case BicepKindString, BicepKindStringLiteral:
		return true
	case BicepKindUnion:
		return t.isEnum()
	default:
		return false
	}
}

func typeMismatch(path string, prop *BicepProperty, tfType string, bt *BicepType) Mismatch {
	return Mismatch{
		Path:   path,
		ARM:    prop.Name,
		Kind:   MismatchTypeMismatch,
		Detail: fmt.Sprintf("compiled %s but bicep type is %s", tfType, bicepDesc(bt)),
	}
}

func bicepDesc(t *BicepType) string {
	switch t.Kind {
	case BicepKindString:
		return "String"
	case BicepKindStringLiteral:
		return "StringLiteral"
	case BicepKindBool:
		return "Bool"
	case BicepKindInt:
		return "Int"
	case BicepKindObject:
		return "Object"
	case BicepKindArray:
		return "Array"
	case BicepKindUnion:
		if t.isEnum() {
			return "Enum"
		}
		return "Union"
	case BicepKindAny:
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

func sortedPropKeys(m map[string]*BicepProperty) []string {
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
