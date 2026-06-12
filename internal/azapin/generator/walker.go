// Package generator walks bicep types.json files and produces an in-memory
// type graph suitable for Terraform schema emission.
package generator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PropertyFlag represents bicep type property flags.
type PropertyFlag int

const (
	FlagNone         PropertyFlag = 0
	FlagRequired     PropertyFlag = 1
	FlagReadOnly     PropertyFlag = 2
	FlagWriteOnly    PropertyFlag = 4
	FlagSystemManaged PropertyFlag = 8
)

func (f PropertyFlag) IsRequired() bool      { return f&FlagRequired != 0 }
func (f PropertyFlag) IsReadOnly() bool      { return f&FlagReadOnly != 0 }
func (f PropertyFlag) IsWriteOnly() bool     { return f&FlagWriteOnly != 0 }
func (f PropertyFlag) IsSystemManaged() bool { return f&FlagSystemManaged != 0 }

// TypeKind identifies what kind of type a graph node is.
type TypeKind int

const (
	KindString TypeKind = iota
	KindStringLiteral
	KindInt
	KindBool
	KindObject
	KindArray
	KindUnion
	KindAny
	KindResource
)

// Type represents a resolved node in the bicep type graph.
type Type struct {
	Kind        TypeKind
	Name        string // For ObjectType: struct name. For StringLiteral: the literal value.
	Properties  map[string]*Property
	ElementType *Type   // For ArrayType: the item type
	Elements    []*Type // For UnionType: the member types
}

// Property represents one property within an ObjectType.
type Property struct {
	Name        string       // ARM JSON name (camelCase)
	Type        *Type        // Resolved type
	Flags       PropertyFlag
	Description string
}

// IsEnum returns true if this type is a union of string literals (enum).
func (t *Type) IsEnum() bool {
	if t.Kind != KindUnion {
		return false
	}
	for _, elem := range t.Elements {
		if elem.Kind != KindStringLiteral && elem.Kind != KindString {
			return false
		}
	}
	return true
}

// EnumValues returns the allowed string values if this is an enum type.
func (t *Type) EnumValues() []string {
	if !t.IsEnum() {
		return nil
	}
	var values []string
	for _, elem := range t.Elements {
		if elem.Kind == KindStringLiteral {
			values = append(values, elem.Name)
		}
	}
	return values
}

// ResourceDefinition is the top-level parsed result for one ARM resource type.
type ResourceDefinition struct {
	Name       string // e.g., "Microsoft.Storage/storageAccounts@2025-01-01"
	APIVersion string
	Body       *Type
}

// rawEntry is a JSON entry in types.json.
type rawEntry struct {
	Type     string                     `json:"$type"`
	Name     string                     `json:"name,omitempty"`
	Value    string                     `json:"value,omitempty"`
	Body     *rawRef                    `json:"body,omitempty"`
	Properties map[string]*rawProperty  `json:"properties,omitempty"`
	ItemType *rawRef                    `json:"itemType,omitempty"`
	Elements []rawRef                   `json:"elements,omitempty"`
	MinValue *int64                     `json:"minValue,omitempty"`
	MaxValue *int64                     `json:"maxValue,omitempty"`
}

type rawRef struct {
	Ref string `json:"$ref"`
}

type rawProperty struct {
	Type        rawRef `json:"type"`
	Flags       int    `json:"flags"`
	Description string `json:"description,omitempty"`
}

// ParseTypesJSON parses a types.json file and returns all ResourceType definitions found.
func ParseTypesJSON(data []byte) ([]*ResourceDefinition, error) {
	var entries []rawEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal types.json: %w", err)
	}

	w := &walker{entries: entries, cache: make(map[int]*Type)}

	var defs []*ResourceDefinition
	for _, entry := range entries {
		if entry.Type != "ResourceType" {
			continue
		}
		bodyIdx, err := parseRef(entry.Body)
		if err != nil {
			continue
		}
		body := w.resolve(bodyIdx, 0)
		if body == nil {
			continue
		}
		name := entry.Name
		apiVersion := ""
		if at := strings.Index(name, "@"); at >= 0 {
			apiVersion = name[at+1:]
			name = name[:at]
		}
		defs = append(defs, &ResourceDefinition{
			Name:       entry.Name,
			APIVersion: apiVersion,
			Body:       body,
		})
	}
	return defs, nil
}

// walker resolves $ref indices into Type nodes with cycle detection.
type walker struct {
	entries []rawEntry
	cache   map[int]*Type
}

const maxDepth = 20

func (w *walker) resolve(idx int, depth int) *Type {
	if depth > maxDepth {
		return &Type{Kind: KindAny, Name: "any"}
	}
	if idx < 0 || idx >= len(w.entries) {
		return nil
	}
	if t, ok := w.cache[idx]; ok {
		return t
	}

	entry := w.entries[idx]

	// Pre-create to break cycles
	var t *Type
	switch entry.Type {
	case "StringType":
		t = &Type{Kind: KindString}
	case "StringLiteralType":
		t = &Type{Kind: KindStringLiteral, Name: entry.Value}
	case "IntegerType":
		t = &Type{Kind: KindInt}
	case "BooleanType":
		t = &Type{Kind: KindBool}
	case "AnyType":
		t = &Type{Kind: KindAny}
	case "ObjectType":
		t = &Type{Kind: KindObject, Name: entry.Name, Properties: make(map[string]*Property)}
		w.cache[idx] = t // cache early for cycles
		for name, prop := range entry.Properties {
			propIdx, err := parseRef(&prop.Type)
			if err != nil {
				continue
			}
			resolved := w.resolve(propIdx, depth+1)
			if resolved == nil {
				resolved = &Type{Kind: KindAny}
			}
			t.Properties[name] = &Property{
				Name:        name,
				Type:        resolved,
				Flags:       PropertyFlag(prop.Flags),
				Description: prop.Description,
			}
		}
		return t
	case "ArrayType":
		t = &Type{Kind: KindArray}
		w.cache[idx] = t
		if entry.ItemType != nil {
			itemIdx, err := parseRef(entry.ItemType)
			if err == nil {
				t.ElementType = w.resolve(itemIdx, depth+1)
			}
		}
		return t
	case "UnionType":
		t = &Type{Kind: KindUnion}
		w.cache[idx] = t
		for _, elem := range entry.Elements {
			elemIdx, err := parseRef(&elem)
			if err != nil {
				continue
			}
			resolved := w.resolve(elemIdx, depth+1)
			if resolved != nil {
				t.Elements = append(t.Elements, resolved)
			}
		}
		return t
	default:
		// ResourceType, ResourceFunctionType, DiscriminatedObjectType, etc.
		// For discriminated objects, treat as Any for now
		t = &Type{Kind: KindAny, Name: entry.Name}
	}

	w.cache[idx] = t
	return t
}

func parseRef(ref *rawRef) (int, error) {
	if ref == nil || ref.Ref == "" {
		return -1, fmt.Errorf("nil ref")
	}
	// Format: "#/123"
	s := ref.Ref
	if !strings.HasPrefix(s, "#/") {
		return -1, fmt.Errorf("unexpected ref format: %s", s)
	}
	var idx int
	if _, err := fmt.Sscanf(s[2:], "%d", &idx); err != nil {
		return -1, err
	}
	return idx, nil
}
