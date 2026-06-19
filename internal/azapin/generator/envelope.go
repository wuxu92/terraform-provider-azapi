package generator

import (
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/naming"
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

// Envelope is the operational-envelope spec wrapped around every generated body
// schema. name maps to the ARM resource name and the parent reference to the
// parent ID; neither is part of the bicep body graph, so the generator
// synthesizes them. id is uniform (computed) and emitted directly.
type Envelope struct {
	Name   EnvelopeAttr
	Parent EnvelopeAttr
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
// synthesizes for the operational envelope (name, the parent reference, id).
// They are not part of the bicep body graph, so schema-vs-bicep validation must
// exclude them (see ValidateEmittedSchema).
func EnvelopeAttrNames(def *ResourceDefinition) []string {
	return []string{def.Envelope.Name.Name, def.Envelope.Parent.Name, "id"}
}

// FindProperty resolves an ARM dot path (e.g. "properties.minimumTlsVersion",
// "sku.name") to the Property it names within the resource body, or nil if any
// segment is missing. Customizers use it to reach a body property and set flags
// such as DefaultValue, ForceNew, Sensitive, or to append Validators.
func FindProperty(def *ResourceDefinition, armPath string) *Property {
	if def == nil {
		return nil
	}
	return navigate(def.Body, armPath)
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

// IntRangeValidator builds an inclusive int64 range validator.
func IntRangeValidator(min, max int64) DescriptionValidator {
	return DescriptionValidator{Kind: ValidatorIntRange, Min: &min, Max: &max}
}
