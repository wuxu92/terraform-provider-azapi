// Package generator walks bicep types.json files and produces an in-memory
// type graph suitable for Terraform schema emission.
package typegraph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
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
	// KindDiscriminated is a lowered bicep DiscriminatedObjectType (polymorphic
	// body): a single flat ARM object tagged by a discriminator property whose
	// value selects one of several variant shapes. Properties holds the shared
	// base properties (discriminator excluded); Variants maps each discriminator
	// value to its variant object (discriminator excluded); Discriminator is the
	// ARM discriminator property name. In Terraform it is modelled as a nested
	// block per variant (one SingleNestedAttribute each, mutually exclusive); the
	// mapper flattens the chosen variant back into the tagged ARM object.
	KindDiscriminated
)

// Type represents a resolved node in the bicep type graph.
type Type struct {
	Kind        TypeKind
	Name        string // For ObjectType: struct name. For StringLiteral: the literal value.
	Properties  map[string]*Property
	ElementType *Type   // For ArrayType: the item type
	Elements    []*Type // For UnionType: the member types
	// Discriminator is the ARM discriminator property name (camelCase, e.g.
	// "kind") of a KindDiscriminated type; empty for every other kind.
	Discriminator string
	// Variants maps a KindDiscriminated type's discriminator value (e.g.
	// "EventHub") to its variant object (a KindObject with the variant's own
	// properties, the redundant discriminator literal excluded). Nil for every
	// other kind. Properties holds the shared base properties.
	Variants map[string]*Type
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
	// SuppressStateReuse opts an Optional+Computed attribute out of the default
	// UseStateForUnknown plan modifier. Set by ApplyAzwise on every member of a
	// mutual-exclusion / choice constraint (ExactlyOneOf, ConflictsWith,
	// AtLeastOneOf). Such a member must be un-settable by omission: with
	// UseStateForUnknown, dropping it from config pins its stale prior value into the
	// plan (so it is re-sent and never clears), while an explicit empty value trips
	// the constraint validator's non-null "is set" test. Dropping the state-reuse
	// lets an omitted member plan as "(known after apply)" so it is left out of the
	// PUT body and the server re-owns it, letting callers switch sides of the
	// constraint. A member explicitly set in config is known and unaffected.
	SuppressStateReuse bool
	// VariantSiblings, when non-empty, marks this property as a discriminated
	// variant block and lists the snake_case attribute names of its sibling
	// variant blocks (every other variant of the same discriminated type). The
	// emitter renders a planmodifiers.DiscriminatedVariant object plan modifier
	// instead of the plain UseStateForUnknown: an unselected variant reuses prior
	// state (so it stops planning as "(known after apply)") but clears to null when
	// a sibling variant is selected in config, preserving the variant-switch
	// semantics an omitted mutual-exclusion member needs. Set by the emitter's
	// variantProperty; mutually exclusive with SuppressStateReuse.
	VariantSiblings []string
	// PlanModifiers holds hand-written plan modifiers a customizer attaches to
	// this property by constructor func reference (see PlanModifier / the
	// AddPlanModifiersFor customizer method). They are appended AFTER the built-in
	// modifiers the emitter derives from flags (UseStateForUnknown / RequiresReplace
	// / the location and discriminated-variant modifiers), in the same typed
	// PlanModifiers slice. Use for schema-level plan behavior the built-in
	// vocabulary cannot express — value normalization / drift suppression,
	// conditional replace, or a bespoke modifier. Each ref's Func must be an
	// uncalled constructor returning the framework plan-modifier type matching this
	// attribute's kind (planmodifier.String for a string attr, .Object for an
	// object, …); a mismatch is a compile error in the regenerated _gen.go.
	PlanModifiers []PlanModifierRef
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
	// Func is the validator constructor referenced by a customizer, e.g.
	// validators.UUID (an uncalled func() validator.String). Stored as any so
	// typegraph stays free of the terraform-plugin-framework dependency; the
	// emitter reflects it (runtime.FuncForPC) to recover the qualified call and
	// the package import. Set only for ValidatorFunc.
	Func any
}

// PlanModifierRef references a hand-written plan-modifier constructor by func
// value, e.g. nativeschema.NormalizeResourceID (an uncalled func returning a
// planmodifier.String). It mirrors DescriptionValidator's ValidatorFunc: the Func
// is stored as any so typegraph stays free of the terraform-plugin-framework
// dependency, and the emitter reflects it (runtime.FuncForPC) to recover the
// qualified call and the package import. Build one with PlanModifier(fn).
type PlanModifierRef struct {
	Func any
}

// PlanModifier builds a PlanModifierRef from a plan-modifier constructor func
// value — e.g. PlanModifier(nativeschema.NormalizeResourceID), where fn is the
// uncalled constructor. Referencing the constructor by symbol means a rename or
// deletion is a compile error at the customizer callsite instead of a silently
// wrong string. The func must return the framework plan-modifier type matching
// the target attribute's kind; typegraph never calls fn or imports its package.
func PlanModifier(fn any) PlanModifierRef {
	return PlanModifierRef{Func: fn}
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
	ValidatorFunc                                            // Hand-written validator referenced by func value: validators.X()
	ValidatorListSizeAtLeast                                 // Minimum list length (listvalidator.SizeAtLeast); uses Min
	ValidatorListSizeAtMost                                  // Maximum list length (listvalidator.SizeAtMost); uses Max
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
	// Timeouts holds the per-operation timeout defaults lowered from azwise.
	// ApplyAzwise populates it; the emitter writes it into the generated
	// services.Descriptor. A zero field means "no override".
	Timeouts Timeouts
}

// Timeouts holds the per-operation timeout defaults lowered from azwise for
// emission into a generated services.Descriptor. A zero field means the runtime
// falls back to its built-in default for that operation.
type Timeouts struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
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

	var defs []*ResourceDefinition
	for _, entry := range entries {
		if entry.Type != "ResourceType" {
			continue
		}
		bodyIdx, err := parseRef(entry.Body)
		if err != nil {
			continue
		}
		// Fresh cache per resource: Type nodes are deduped WITHIN a resource (DAG
		// sharing breaks reference cycles) but never shared ACROSS resources. A
		// shared cache leaks per-resource mutations — PostProcess/ApplyAzwise and
		// the customizers write onto Property/Type nodes — from one resource onto
		// another that references the same types.json index (e.g. Microsoft.Web
		// serverfarms and staticSites share the SkuDescription "sku.name" node, so
		// staticSites' azwise default would otherwise land on serverfarms).
		w := &walker{entries: entries, cache: make(map[int]*Type), inProgress: make(map[int]bool)}
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

// maxDiscriminatedVariants caps how many variants a discriminated body may carry
// before it degrades to a dynamic (KindAny) attribute. One SingleNestedAttribute
// is emitted per variant, so a wide union (the ARM tail reaches 121 variants)
// would produce an unusably large schema. The common case is 1-12 variants; a
// per-resource customizer can still force typing when a wide union is worth it.
const maxDiscriminatedVariants = 25

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
		// Properties resolve in a stable order (see resolveProps).
		for name, prop := range w.resolveProps(entry.Properties, depth) {
			obj.Properties[name] = prop
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
	case "DiscriminatedObjectType":
		// A polymorphic body: a discriminator property tags one of several variant
		// shapes. baseProperties is the shared set; elements is a JSON object
		// (discriminator value -> variant $ref), unlike UnionType's array. Model it
		// as KindDiscriminated with the discriminator excluded from both the base
		// properties and each variant (it is synthesized by the mapper from the
		// selected variant, never surfaced as an attribute).
		disc := &Type{Kind: KindDiscriminated, Name: entry.Name, Discriminator: entry.Discriminator, Properties: make(map[string]*Property), Variants: make(map[string]*Type)}
		for name, prop := range w.resolveProps(entry.BaseProperties, depth) {
			if name == entry.Discriminator {
				continue
			}
			disc.Properties[name] = prop
		}
		var variantRefs map[string]rawRef
		if len(entry.Elements) > 0 {
			_ = json.Unmarshal(entry.Elements, &variantRefs)
		}
		if len(variantRefs) > maxDiscriminatedVariants {
			// Wide union: fall back to the dynamic representation rather than emit
			// one nested block per variant (see maxDiscriminatedVariants).
			t = &Type{Kind: KindAny, Name: entry.Name}
			break
		}
		for value, ref := range variantRefs {
			vIdx, err := parseRef(&ref)
			if err != nil {
				continue
			}
			resolved := w.resolve(vIdx, depth+1)
			if resolved == nil || resolved.Kind != KindObject {
				continue
			}
			// Copy the variant object so stripping the discriminator does not mutate
			// a shared cached node.
			variant := &Type{Kind: KindObject, Name: resolved.Name, Properties: make(map[string]*Property, len(resolved.Properties))}
			for pName, p := range resolved.Properties {
				if pName == entry.Discriminator {
					continue
				}
				variant.Properties[pName] = p
			}
			disc.Variants[value] = variant
		}
		t = disc
	default:
		// ResourceType, ResourceFunctionType, etc. — non-body categories that never
		// appear as a settable property type. Emitted as a dynamic attribute.
		t = &Type{Kind: KindAny, Name: entry.Name}
	}
	w.cache[idx] = t
	return t
}

// resolveProps resolves a bicep property map into Property nodes in a stable
// (sorted) order. entry.Properties/baseProperties is a Go map whose iteration
// order is randomized per run; for mutually-recursive ARM types that randomness
// decides which reference realizes the full ObjectType and which becomes the
// KindAny cycle-break sentinel, so a stable order keeps generation reproducible
// and keeps the emitter and the compiled-schema validator in agreement.
func (w *walker) resolveProps(props map[string]*rawProperty, depth int) map[string]*Property {
	out := make(map[string]*Property, len(props))
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		prop := props[name]
		propIdx, err := parseRef(&prop.Type)
		if err != nil {
			continue
		}
		resolved := w.resolve(propIdx, depth+1)
		if resolved == nil {
			resolved = &Type{Kind: KindAny}
		}
		out[name] = &Property{
			Name:        name,
			Type:        resolved,
			Flags:       PropertyFlag(prop.Flags),
			Description: prop.Description,
		}
	}
	return out
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
