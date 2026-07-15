// Package validate provides a compiled-schema validator that cross-references
// generated Terraform schema.Schema objects against the bicep type graph.
// It reuses typegraph.Type — no duplicate walker — and the shared verification
// model (typegraph.Mismatch, ADR-0008).
package validate

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

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
// camelCase ARM properties via the shared verification leaf rule. ignoreTopLevel
// lists top-level attribute names to exclude — the synthesized operational-envelope
// attributes (name / parent reference / id) are not part of the bicep body.
func SchemaAgainstBicep(s schema.Schema, body *typegraph.Type, ignoreTopLevel ...string) []typegraph.Mismatch {
	if body == nil || body.Kind != typegraph.KindObject {
		return []typegraph.Mismatch{{Kind: typegraph.MismatchTypeMismatch, Detail: "bicep body is not an ObjectType"}}
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

func validateObject(tfAttrs map[string]schema.Attribute, bicepObj *typegraph.Type, prefix string) []typegraph.Mismatch {
	var mismatches []typegraph.Mismatch

	// Build snake_case → bicep property map using the shared leaf rule so this
	// adapter skips and renames identically to the source adapter.
	bicepBySnake := make(map[string]*typegraph.Property)
	for armName, prop := range bicepObj.Properties {
		snake, skip := typegraph.SchemaAttrName(armName, prop)
		if skip {
			continue
		}
		bicepBySnake[snake] = prop
	}

	// Check: every TF attribute exists in bicep
	for _, tfName := range sortedAttrKeys(tfAttrs) {
		tfAttr := tfAttrs[tfName]
		path := typegraph.JoinPath(prefix, tfName)

		bicepProp, ok := bicepBySnake[tfName]
		if !ok {
			mismatches = append(mismatches, typegraph.Mismatch{
				Path:   path,
				Kind:   typegraph.MismatchExtraInSchema,
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
		path := typegraph.JoinPath(prefix, snake)
		mismatches = append(mismatches, typegraph.Mismatch{
			Path:   path,
			ARM:    prop.Name,
			Kind:   typegraph.MismatchMissingInSchema,
			Detail: fmt.Sprintf("bicep property %q not in compiled schema", prop.Name),
		})
	}

	return mismatches
}

func validateType(tfAttr schema.Attribute, bicepProp *typegraph.Property, path string) []typegraph.Mismatch {
	var mismatches []typegraph.Mismatch
	bt := bicepProp.Type

	switch attr := tfAttr.(type) {
	case schema.StringAttribute:
		if !isStringCompat(bt) {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "StringAttribute", bt))
		}
	case schema.BoolAttribute:
		if bt.Kind != typegraph.KindBool {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "BoolAttribute", bt))
		}
	case schema.Int64Attribute:
		if bt.Kind != typegraph.KindInt {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "Int64Attribute", bt))
		}
	case schema.SingleNestedAttribute:
		switch bt.Kind {
		case typegraph.KindObject:
			mismatches = append(mismatches, validateObject(attr.Attributes, bt, path)...)
		case typegraph.KindDiscriminated:
			mismatches = append(mismatches, validateObject(attr.Attributes, discriminatedAsObject(bt), path)...)
		default:
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "SingleNestedAttribute", bt))
		}
	case schema.ListNestedAttribute:
		if bt.Kind != typegraph.KindArray {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "ListNestedAttribute", bt))
		} else if bt.ElementType != nil && bt.ElementType.Kind == typegraph.KindObject {
			mismatches = append(mismatches, validateObject(attr.NestedObject.Attributes, bt.ElementType, path+"[*]")...)
		} else if bt.ElementType != nil && bt.ElementType.Kind == typegraph.KindDiscriminated {
			mismatches = append(mismatches, validateObject(attr.NestedObject.Attributes, discriminatedAsObject(bt.ElementType), path+"[*]")...)
		}
	case schema.ListAttribute:
		if bt.Kind != typegraph.KindArray {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "ListAttribute", bt))
		}
	case schema.MapNestedAttribute:
		if bt.Kind != typegraph.KindMap {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "MapNestedAttribute", bt))
		} else if bt.ElementType != nil && bt.ElementType.Kind == typegraph.KindObject {
			mismatches = append(mismatches, validateObject(attr.NestedObject.Attributes, bt.ElementType, path)...)
		}
	case schema.MapAttribute:
		if bt.Kind != typegraph.KindMap {
			mismatches = append(mismatches, typeMismatch(path, bicepProp, "MapAttribute", bt))
		}
	case schema.DynamicAttribute:
		// Dynamic accepts anything
	}

	return mismatches
}

func isStringCompat(t *typegraph.Type) bool {
	switch t.Kind {
	case typegraph.KindString, typegraph.KindStringLiteral:
		return true
	case typegraph.KindUnion:
		// The generator emits enum unions and non-enum unions as StringAttribute.
		return true
	case typegraph.KindAny, typegraph.KindMap:
		// Dynamic/Map cannot be nested inside collection elements in the framework;
		// the emitter intentionally degrades those shapes to StringAttribute there.
		return true
	default:
		return false
	}
}

func typeMismatch(path string, prop *typegraph.Property, tfType string, bt *typegraph.Type) typegraph.Mismatch {
	return typegraph.Mismatch{
		Path:   path,
		ARM:    prop.Name,
		Kind:   typegraph.MismatchTypeMismatch,
		Detail: fmt.Sprintf("compiled %s but bicep type is %s", tfType, bicepDesc(bt)),
	}
}

func bicepDesc(t *typegraph.Type) string {
	switch t.Kind {
	case typegraph.KindString:
		return "String"
	case typegraph.KindStringLiteral:
		return "StringLiteral"
	case typegraph.KindBool:
		return "Bool"
	case typegraph.KindInt:
		return "Int"
	case typegraph.KindObject:
		return "Object"
	case typegraph.KindArray:
		return "Array"
	case typegraph.KindUnion:
		if t.IsEnum() {
			return "Enum"
		}
		return "Union"
	case typegraph.KindAny:
		return "Any"
	case typegraph.KindMap:
		return "Map"
	case typegraph.KindDiscriminated:
		return "Discriminated"
	default:
		return fmt.Sprintf("Unknown(%d)", t.Kind)
	}
}

// discriminatedAsObject projects a KindDiscriminated type onto the KindObject
// shape the emitter produces: the base properties plus one nested object property
// per variant (keyed by the discriminator value). This lets validateObject check a
// discriminated block's emitted SingleNestedAttribute against the same attribute
// set the emitter wrote, so parity holds without a discriminated-specific walker.
func discriminatedAsObject(t *typegraph.Type) *typegraph.Type {
	obj := &typegraph.Type{Kind: typegraph.KindObject, Name: t.Name, Properties: make(map[string]*typegraph.Property, len(t.Properties)+len(t.Variants))}
	for name, prop := range t.Properties {
		obj.Properties[name] = prop
	}
	for value, variant := range t.Variants {
		obj.Properties[value] = &typegraph.Property{Name: value, Type: variant}
	}
	return obj
}

func sortedAttrKeys(m map[string]schema.Attribute) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedPropKeys(m map[string]*typegraph.Property) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
