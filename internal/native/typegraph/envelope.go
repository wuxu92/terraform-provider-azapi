package typegraph

import (
	"fmt"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
)

// This file owns the operational-envelope generation that wraps every generated
// body schema, plus the validator-builder vocabulary the emitter knows how to
// emit. Per-resource schema customization is a separate concern and lives in the
// generator/customizers sub-package — generator core never imports it.

// EnvelopeAttr describes one operational-envelope attribute (name or the parent
// reference). The id attribute is uniform and not represented here.
type EnvelopeAttr struct {
	Name        string
	Description string
	Validators  []DescriptionValidator
}

// MetaAttr describes one synthetic, behavior-only top-level attribute that is NOT
// part of the bicep body graph and is never sent to or read from ARM: it controls
// provider-side behavior (e.g. purge_on_destroy, which drives a destroy-time purge).
// The mapper skips it on Expand (absent from the body graph) and leaves it null on
// flatten, so an Optional-only bool round-trips without drift when unset; a hook
// reads it from state. It is emitted as an Optional bool schema attribute.
type MetaAttr struct {
	Name        string
	Description string
}

// Envelope is the operational-envelope spec wrapped around every generated body
// schema. name maps to the ARM resource name and the parent reference to the
// parent ID; neither is part of the bicep body graph, so the generator
// synthesizes them. id is uniform (computed) and emitted directly. Meta holds
// synthetic behavior-only attributes a customizer attaches (never from bicep).
type Envelope struct {
	Name   EnvelopeAttr
	Parent EnvelopeAttr
	Meta   []MetaAttr
}

// ARMTypeOf returns the ARM resource type of a definition without its API
// version ("Microsoft.Storage/storageAccounts@2025-01-01" ->
// "Microsoft.Storage/storageAccounts").
func ARMTypeOf(def *ResourceDefinition) string {
	name := def.Name
	if at := strings.Index(name, "@"); at >= 0 {
		name = name[:at]
	}
	return name
}

// applyEnvelopeDefaults seeds def.Envelope from the ARM type and writable scope.
// The parent reference (name + ID-shape validator) comes from the single source
// of truth, naming.ParentReference; the name attribute has no generic default
// validator (resource-specific rules come from a customizer).
func applyEnvelopeDefaults(def *ResourceDefinition) {
	ref := naming.ParentReference(ARMTypeOf(def), def.WritableScopes)
	def.Envelope = Envelope{
		Name: EnvelopeAttr{
			Name:        "name",
			Description: "Specifies the name of the Azure resource.",
		},
		Parent: EnvelopeAttr{
			Name:        ref.Name,
			Description: "The ID of the parent resource that contains this resource.",
		},
	}
	if ref.Pattern != "" {
		def.Envelope.Parent.Validators = []DescriptionValidator{
			RegexValidator(ref.Pattern, ref.Message),
		}
	}
}

// EnvelopeAttrNames returns the top-level attribute names the generator
// synthesizes outside the bicep body graph: the operational envelope (name, the
// parent reference, id) plus any behavior-only Meta attributes a customizer
// attached. They are not part of the bicep body graph, so schema-vs-bicep
// validation must exclude them (see ValidateEmittedSchema).
func EnvelopeAttrNames(def *ResourceDefinition) []string {
	names := []string{def.Envelope.Name.Name, def.Envelope.Parent.Name, "id"}
	for _, m := range def.Envelope.Meta {
		names = append(names, m.Name)
	}
	return names
}

// FindProperty resolves an ARM dot path (e.g. "properties.minimumTlsVersion",
// "sku.name") to the Property it names within the resource body. Customizers use
// it to reach a body property and set flags such as DefaultValue, ForceNew,
// Sensitive, or to append Validators.
//
// It panics if the path does not resolve: a customizer references paths by hand,
// so an unresolved path is a developer typo that must surface at generation time,
// not be silently skipped.
func FindProperty(def *ResourceDefinition, armPath string) *Property {
	if def == nil {
		panic("native: FindProperty called with a nil ResourceDefinition")
	}
	p := Navigate(def.Body, armPath)
	if p == nil {
		panic(fmt.Sprintf("native: customizer for %s references unknown property path %q", def.Name, armPath))
	}
	return p
}

// IsolateArrayElement un-shares and returns the element object type of the array
// property at armPath (e.g. "properties.networkAcls.ipRules"). The bicep graph
// deduplicates structurally identical types, so two arrays of the same shape
// (e.g. ipRules and ipv6Rules) point at one element *Type; mutating it in place
// would leak the change to every sibling. IsolateArrayElement replaces the
// array's element type with a private copy (the element struct plus a fresh
// Properties map of per-property copies), so a customizer can attach validators
// or flags to one array's elements without touching the others.
//
// It panics if armPath does not resolve or is not an array of objects — both are
// customizer-authoring errors that must surface at generation time.
func IsolateArrayElement(def *ResourceDefinition, armPath string) *Type {
	arr := FindProperty(def, armPath)
	if arr.Type == nil || arr.Type.Kind != KindArray {
		panic(fmt.Sprintf("native: customizer for %s: IsolateArrayElement path %q is not an array", def.Name, armPath))
	}
	elem := arr.Type.ElementType
	if elem == nil || elem.Kind != KindObject {
		panic(fmt.Sprintf("native: customizer for %s: IsolateArrayElement path %q is not an array of objects", def.Name, armPath))
	}
	clone := *elem
	clone.Properties = make(map[string]*Property, len(elem.Properties))
	for name, p := range elem.Properties {
		pc := *p
		if len(p.Validators) > 0 {
			pc.Validators = append([]DescriptionValidator(nil), p.Validators...)
		}
		clone.Properties[name] = &pc
	}
	arr.Type.ElementType = &clone
	return &clone
}

// ---------------------------------------------------------------------------
// Validator constructors — ergonomic builders for the DescriptionValidator
// representation the emitter already knows how to emit. The envelope seeding
// above and the generator/customizers sub-package both build validators with
// these.
// ---------------------------------------------------------------------------

// RegexValidator builds a regex-match validator.
func RegexValidator(pattern, message string) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorRegex, Pattern: pattern, Message: message}
}

// LengthValidator builds a string-length validator. A negative bound means
// unbounded on that side (e.g. LengthValidator(3, -1) is "at least 3").
func LengthValidator(min, max int) DescriptionValidator {
	dv := DescriptionValidator{Kind: ValidatorStringLength}
	if min >= 0 {
		m := int64(min)
		dv.Min = &m
	}
	if max >= 0 {
		m := int64(max)
		dv.Max = &m
	}
	return dv
}

// OneOfValidator builds a string-enum (one-of) validator.
func OneOfValidator(message string, allowed ...string) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorStringOneOf, Allowed: allowed, Message: message}
}

// OneOfCaseInsensitiveValidator builds a case-insensitive string-enum validator.
// It is emitted for list-element permission enums where ARM accepts any casing.
func OneOfCaseInsensitiveValidator(message string, allowed ...string) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorStringOneOfCaseInsensitive, Allowed: allowed, Message: message}
}

// IntRangeValidator builds an inclusive int64 range validator.
func IntRangeValidator(min, max int64) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorIntRange, Min: &min, Max: &max}
}

// CustomValidator builds a reference to a resource-/service-specific validator
// that lives with the generated schema in services/<service>/validators. The
// call is emitted qualified with that package (validators.<call>), so it must name
// an exported constructor there returning a validator.String. Use it for a
// semantic rule unique to one resource, e.g. CustomValidator("StorageAccountIPRule()").
func CustomValidator(call string) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorCustom, Call: call}
}

// SharedValidator builds a reference to a generic, cross-resource validator in the
// shared native schema package (internal/native/schema). The call is emitted
// qualified as nativeschema.<call>, so it must name an exported constructor there
// returning a validator.String. Use it for reusable rules like
// SharedValidator("UUID()") or SharedValidator("AzureResourceID()").
func SharedValidator(call string) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorShared, Call: call}
}
