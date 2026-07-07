// Package generator walks bicep types.json files and produces an in-memory
// type graph suitable for Terraform schema emission.
package typegraph

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PropertyFlag represents bicep type property flags.
type PropertyFlag int

const (
	FlagNone          PropertyFlag = 0
	FlagRequired      PropertyFlag = 1
	FlagReadOnly      PropertyFlag = 2
	FlagWriteOnly     PropertyFlag = 4
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
	// KindMap is an open dictionary: a bicep ObjectType with no declared
	// properties but an additionalProperties schema (e.g. ARM "tags"). The keys
	// are arbitrary strings; ElementType is the resolved value type.
	KindMap
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
	Name         string // ARM JSON name (camelCase)
	Type         *Type  // Resolved type
	Flags        PropertyFlag
	Description  string
	DefaultValue string                 // Extracted from description or azwise, empty if none
	Validators   []DescriptionValidator // Extracted from description or azwise
	// azwise-derived overlay (set by ApplyAzwise; empty/false otherwise)
	ForceNew      bool // emit a RequiresReplace plan modifier
	Sensitive     bool // emit Sensitive: true
	ForceComputed bool // force Computed-only (azwise ComputedFields)
	// NonNullStateForUnknown opts a computed attribute out of the default
	// UseStateForUnknown in favour of UseNonNullStateForUnknown: a null prior
	// plans as "(known after apply)" so the server may populate it, instead of
	// pinning the stale null over a value the server later supplies. Set by a
	// customizer for server-controlled read-only fields that transition
	// null -> non-null when a sibling/parent is configured (e.g. blobServices
	// lastAccessTimeTrackingPolicy.name). Off by default: a blanket switch
	// breaks idempotency for fields the server legitimately leaves null.
	NonNullStateForUnknown bool
	// UseSet emits a primitive array as schema.SetAttribute instead of ListAttribute.
	// Use it only when ARM treats the collection as unordered (e.g. CORS header names)
	// so API reordering does not create drift. Set by customizers.
	UseSet bool
	// DefaultEmptyList emits a schema Default of an empty typed list for an
	// Optional+Computed array attribute (Default: listdefault.StaticValue of an
	// empty list). Use it when ARM always echoes an omitted list back as [] rather
	// than absent: the empty-list default makes the omitted-config plan value ([])
	// match the API's [] on apply/read/import, while an explicitly-configured list
	// (including an explicit []) is left untouched because the default only fires on
	// a null config value. Set by customizers; ignored for non-array properties.
	DefaultEmptyList bool
}

// DescriptionValidator is a validation rule extracted from a property description
// or from azwise knowledge.
type DescriptionValidator struct {
	Kind    ValidatorKind
	Pattern string   // For regex/format validators
	Min     *int64   // For numeric range / length validators
	Max     *int64   // For numeric range / length validators
	Allowed []string // For OneOf validators
	Message string   // Human-readable description
	Call    string   // For ValidatorCustom: an nativeschema constructor call, e.g. "StorageAccountIPRule()"
}

// ValidatorKind identifies the type of validator.
type ValidatorKind int

const (
	ValidatorArmResourceID              ValidatorKind = iota // ARM resource ID format
	ValidatorRegex                                           // Regex pattern
	ValidatorIntRange                                        // Numeric min/max
	ValidatorStringOneOf                                     // String enum (azwise AllowedValues)
	ValidatorStringOneOfCaseInsensitive                      // Case-insensitive string enum (list-element permissions)
	ValidatorStringLength                                    // String length min/max (azwise)
	ValidatorCustom                                          // Service-specific validator: services/<service>/validators (validators.X())
	ValidatorShared                                          // Generic shared validator: internal/native/schema (nativeschema.X())
)

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
	// WritableScopes is the bicep scope bitmask for the resource (Tenant=1,
	// ManagementGroup=2, Subscription=4, ResourceGroup=8, Extension=16). It seeds
	// the default operational envelope (see Envelope) at generation time.
	WritableScopes int
	// Envelope is the operational-envelope spec (name + parent reference) baked
	// into the generated schema. PostProcess populates it from the ARM type and
	// WritableScopes; a customizer plugin may override it before emission.
	Envelope Envelope
	// Relational holds cross-property constraints lowered from azwise (ConflictsWith /
	// RequiredWith / ExactlyOneOf / AtLeastOneOf). ApplyAzwise populates it; the
	// emitter writes it into the generated services.Descriptor.
	Relational []RelationalConstraintDef
}

// RelationalConstraintDef is a lowered cross-property constraint ready for
// emission. Kind is the services.RelationalKind constant name ("ConflictsWith",
// "RequiredWith", "ExactlyOneOf", "AtLeastOneOf"). Paths are snake_case schema
// segments, absolute from the schema root; Paths[0] is the subject for
// ConflictsWith/RequiredWith.
type RelationalConstraintDef struct {
	Kind    string
	Paths   [][]string
	Message string
}

// rawEntry is a JSON entry in types.json.
type rawEntry struct {
	Type           string                  `json:"$type"`
	Name           string                  `json:"name,omitempty"`
	WritableScopes int                     `json:"writableScopes,omitempty"`
	Value          string                  `json:"value,omitempty"`
	Body           *rawRef                 `json:"body,omitempty"`
	Properties     map[string]*rawProperty `json:"properties,omitempty"`
	ItemType       *rawRef                 `json:"itemType,omitempty"`
	// AdditionalProperties is the value schema of an open dictionary ObjectType
	// (a map[string]T such as ARM "tags"); nil for a closed object.
	AdditionalProperties *rawRef `json:"additionalProperties,omitempty"`
	// Elements is json.RawMessage because its shape is type-dependent:
	// UnionType uses an array ([]rawRef); DiscriminatedObjectType uses an
	// object (map[string]rawRef). Decoding it lazily in resolve() avoids
	// failing the entire document's unmarshal on the shape mismatch.
	Elements json.RawMessage `json:"elements,omitempty"`
	// BaseProperties is the shared property set of a DiscriminatedObjectType.
	BaseProperties map[string]*rawProperty `json:"baseProperties,omitempty"`
	Discriminator  string                  `json:"discriminator,omitempty"`
	MinValue       *int64                  `json:"minValue,omitempty"`
	MaxValue       *int64                  `json:"maxValue,omitempty"`
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

	w := &walker{entries: entries, cache: make(map[int]*Type), inProgress: make(map[int]bool)}

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
			Name:           entry.Name,
			APIVersion:     apiVersion,
			Body:           body,
			WritableScopes: entry.WritableScopes,
		})
	}
	return defs, nil
}

// walker resolves $ref indices into Type nodes. It breaks reference cycles
// (self-referential ARM types like the ErrorEntity/ErrorDetail pattern) by
// tracking in-progress indices: a back-edge to a node still being built is
// replaced with a KindAny sentinel, keeping the produced graph a DAG so that
// downstream consumers (emitter, postprocess, validators) terminate.
type walker struct {
	entries    []rawEntry
	cache      map[int]*Type
	inProgress map[int]bool
}

const maxDepth = 50

func (w *walker) resolve(idx int, depth int) *Type {
	if idx < 0 || idx >= len(w.entries) {
		return nil
	}
	if t, ok := w.cache[idx]; ok {
		return t // fully resolved node — safe to share (DAG)
	}
	if w.inProgress[idx] {
		// Back-edge: this node is an ancestor still being built. Returning the
		// partial node here would create a cycle; emit a sentinel instead.
		return &Type{Kind: KindAny, Name: "recursive"}
	}
	if depth > maxDepth {
		return &Type{Kind: KindAny, Name: "any"}
	}
	entry := w.entries[idx]
	w.inProgress[idx] = true
	defer delete(w.inProgress, idx)
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
		obj := &Type{Kind: KindObject, Name: entry.Name, Properties: make(map[string]*Property)}
		for name, prop := range entry.Properties {
			propIdx, err := parseRef(&prop.Type)
			if err != nil {
				continue
			}
			resolved := w.resolve(propIdx, depth+1)
			if resolved == nil {
				resolved = &Type{Kind: KindAny}
			}
			obj.Properties[name] = &Property{
				Name:        name,
				Type:        resolved,
				Flags:       PropertyFlag(prop.Flags),
				Description: prop.Description,
			}
		}
		// An ObjectType with no declared properties but an additionalProperties
		// schema is an open dictionary (map[string]T) — e.g. ARM "tags". Model it
		// as a KindMap keyed by string with the resolved value type. Emitting it
		// as an empty KindObject (the old behavior) produced an unusable
		// SingleNestedAttribute with zero attributes.
		if len(obj.Properties) == 0 && entry.AdditionalProperties != nil {
			elem := &Type{Kind: KindAny}
			if elemIdx, err := parseRef(entry.AdditionalProperties); err == nil {
				if resolved := w.resolve(elemIdx, depth+1); resolved != nil {
					elem = resolved
				}
			}
			t = &Type{Kind: KindMap, Name: entry.Name, ElementType: elem}
		} else {
			t = obj
		}
	case "ArrayType":
		t = &Type{Kind: KindArray}
		if entry.ItemType != nil {
			if itemIdx, err := parseRef(entry.ItemType); err == nil {
				t.ElementType = w.resolve(itemIdx, depth+1)
			}
		}
	case "UnionType":
		t = &Type{Kind: KindUnion}
		var elems []rawRef
		if len(entry.Elements) > 0 {
			_ = json.Unmarshal(entry.Elements, &elems)
		}
		for i := range elems {
			elemIdx, err := parseRef(&elems[i])
			if err != nil {
				continue
			}
			if resolved := w.resolve(elemIdx, depth+1); resolved != nil {
				t.Elements = append(t.Elements, resolved)
			}
		}
	default:
		// DiscriminatedObjectType, ResourceType, ResourceFunctionType, etc.
		// Polymorphic/discriminated bodies are emitted as a dynamic attribute.
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
