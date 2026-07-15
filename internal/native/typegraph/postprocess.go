package typegraph

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
)

// PostProcess applies semantic rules to a parsed type graph that can't be
// derived from bicep flags alone. Call after ParseTypesJSON.
//
// Rules applied (in order):
//   - Extract default values from property descriptions
//   - Extract validators from property descriptions (ARM ID, datetime, numeric ranges)
//   - Promote single-optional-child block properties to Required
//   - Overlay azwise knowledge (ForceNew, computed, sensitive, verified defaults,
//     validation) so curated AzureRM knowledge is the base for the schema
//   - Mark writable top-level location as ForceNew: ARM tracked-resource locations
//     are creation-time placement, not in-place mutable state
//   - Seed the operational-envelope spec (name + parent reference) from the ARM
//     type and writable scope
//
// Per-resource developer customizers run separately, after PostProcess, via the
// generator/customizers sub-package (customizers.Apply); generator core does not
// depend on them so the runtime never pulls customizers in.
func PostProcess(defs []*ResourceDefinition) {
	for _, def := range defs {
		if def.Body != nil {
			extractDefaults(def.Body)
			extractDescriptionValidators(def.Body)
			promoteSingleOptional(def.Body)
			ApplyAzwise(def)
			markTopLevelLocationForceNew(def)
			demoteDefaultedRequired(def.Body)
			applyEnvelopeDefaults(def)
			synthesizeDiscriminatorConstraints(def)
		}
	}
}

func markTopLevelLocationForceNew(def *ResourceDefinition) {
	if def == nil || def.Body == nil || def.Body.Kind != KindObject {
		return
	}
	loc := def.Body.Properties["location"]
	if loc == nil || loc.Type == nil || loc.Type.Kind != KindString || EffectiveComputed(loc) {
		return
	}
	loc.ForceNew = true
}

// synthesizeDiscriminatorConstraints walks the body for discriminated properties
// reachable by an object-only path and appends a mutual-exclusion constraint over
// their variant blocks: ExactlyOneOf when the discriminated block is Required,
// AtMostOneOf otherwise. The runtime enforces these as resource-level
// ConfigValidators (see resource/configvalidators.go), the same machinery azwise
// relational rules use. A discriminated type nested inside an array/map is skipped
// because the config-path addressing is object-only (matches resolveObjectPathSegments).
func synthesizeDiscriminatorConstraints(def *ResourceDefinition) {
	if def == nil || def.Body == nil {
		return
	}
	walkDiscriminated(def.Body, nil, func(disc *Type, required bool, prefix []string) {
		if len(disc.Variants) < 2 {
			return // one (or zero) variant: nothing is mutually exclusive
		}
		values := make([]string, 0, len(disc.Variants))
		for value := range disc.Variants {
			values = append(values, value)
		}
		sort.Strings(values)
		paths := make([][]string, 0, len(values))
		for _, value := range values {
			path := make([]string, 0, len(prefix)+1)
			path = append(path, prefix...)
			paths = append(paths, append(path, naming.CamelToSnake(value)))
		}
		kind := "AtMostOneOf"
		msg := "at most one variant of the discriminated block may be set"
		if required {
			kind = "ExactlyOneOf"
			msg = "exactly one variant of the discriminated block must be set"
		}
		def.Relational = append(def.Relational, RelationalConstraintDef{Kind: kind, Paths: paths, Message: msg})
	})
}

// walkDiscriminated visits every discriminated type reachable from typ by an
// object-only path, invoking fn with the discriminated type, whether the property
// holding it is Required, and the snake_case path prefix to the block. It descends
// object properties (and into each discriminated block's base properties and
// variants) but not arrays or maps, whose elements cannot be addressed by the
// object-only relational config-path walker.
func walkDiscriminated(typ *Type, prefix []string, fn func(disc *Type, required bool, prefix []string)) {
	if typ == nil {
		return
	}
	switch typ.Kind {
	case KindObject, KindDiscriminated:
		for armName, prop := range typ.Properties {
			if prop == nil || prop.Flags.IsSystemManaged() {
				continue
			}
			childPrefix := append(append([]string{}, prefix...), naming.CamelToSnake(armName))
			if prop.Type != nil && prop.Type.Kind == KindDiscriminated {
				fn(prop.Type, prop.Flags.IsRequired(), childPrefix)
			}
			walkDiscriminated(prop.Type, childPrefix, fn)
		}
		// A discriminated block's variants are addressed as nested blocks; descend
		// them so a variant that itself holds a discriminated property is covered.
		for _, value := range sortedKeys(typ.Variants) {
			variantPrefix := append(append([]string{}, prefix...), naming.CamelToSnake(value))
			walkDiscriminated(typ.Variants[value], variantPrefix, fn)
		}
	}
}

func sortedKeys(m map[string]*Type) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// demoteDefaultedRequired makes a Default override the Required flag: a property
// with a known default value is, by definition, omittable, so it is emitted as
// Optional+Computed+Default rather than Required. The framework forbids
// Required+Default, and forcing the user to set a field that already has a curated
// default is poor UX (azurerm makes e.g. account_kind / network default_action
// optional-with-default for the same reason). Read-only/computed-only properties
// are left untouched — a default there is a knowledge conflict that
// CheckFlagInvariants surfaces instead.
func demoteDefaultedRequired(typ *Type) {
	if typ == nil {
		return
	}
	switch typ.Kind {
	case KindObject:
		for _, prop := range typ.Properties {
			if prop == nil {
				continue
			}
			if prop.DefaultValue != "" && prop.Flags.IsRequired() && !EffectiveComputed(prop) {
				prop.Flags &^= FlagRequired
			}
			demoteDefaultedRequired(prop.Type)
		}
	case KindArray:
		demoteDefaultedRequired(typ.ElementType)
	}
}

// ---------------------------------------------------------------------------
// Default value extraction from descriptions
// ---------------------------------------------------------------------------

// defaultPatterns matches description text that indicates a default value.
// Ordered by specificity — first match wins.
var defaultPatterns = []*regexp.Regexp{
	// "The default value is true" / "default value is false"
	regexp.MustCompile(`(?i)(?:the\s+)?default\s+value\s+is\s+['"]?(\w+)['"]?`),
	// "defaults to true" / "Defaults to TLS1_2"
	regexp.MustCompile(`(?i)defaults?\s+to\s+['"]?([A-Za-z0-9_.]+)['"]?`),
	// "default is NoRootSquash"
	regexp.MustCompile(`(?i)(?:the\s+)?default\s+is\s+['"]?([A-Za-z0-9_.]+)['"]?`),
	// "enabled by default" → true
	regexp.MustCompile(`(?i)enabled\s+by\s+default`),
	// "disabled by default" → false
	regexp.MustCompile(`(?i)disabled\s+by\s+default`),
}

func extractDefaults(typ *Type) {
	if typ.Kind != KindObject && typ.Kind != KindDiscriminated {
		return
	}
	for _, prop := range typ.Properties {
		if prop.Flags.IsRequired() || prop.Flags.IsReadOnly() || prop.Flags.IsSystemManaged() {
			continue
		}
		if prop.Description == "" {
			continue
		}

		defaultVal := parseDefaultFromDescription(prop.Description, prop.Type)
		if defaultVal != "" {
			prop.DefaultValue = defaultVal
		}

		// Recurse into nested objects and discriminated blocks.
		if prop.Type.Kind == KindObject || prop.Type.Kind == KindDiscriminated {
			extractDefaults(prop.Type)
		}
	}
	// A discriminated block's variants are addressed as nested blocks; each is a
	// KindObject, so extract their defaults through the same object path.
	for _, value := range sortedKeys(typ.Variants) {
		extractDefaults(typ.Variants[value])
	}
}

// parseDefaultFromDescription extracts a default value from a property description.
// Returns empty string if no default is found or the value can't be validated.
func parseDefaultFromDescription(desc string, typ *Type) string {
	for _, pat := range defaultPatterns {
		m := pat.FindStringSubmatch(desc)
		if m == nil {
			continue
		}

		// Special cases for "enabled/disabled by default"
		patStr := pat.String()
		if strings.Contains(patStr, "enabled\\s+by") {
			if typ.Kind == KindBool {
				return "true"
			}
			continue
		}
		if strings.Contains(patStr, "disabled\\s+by") {
			if typ.Kind == KindBool {
				return "false"
			}
			continue
		}

		if len(m) < 2 {
			continue
		}
		val := m[1]

		// Validate the extracted value against the property type
		if validateDefault(val, typ) {
			return normalizeDefault(val, typ)
		}
	}
	return ""
}

// validateDefault checks if a candidate default value is valid for the type.
func validateDefault(val string, typ *Type) bool {
	if val == "" || val == "null" || val == "undefined" {
		return false
	}

	switch typ.Kind {
	case KindBool:
		lower := strings.ToLower(val)
		return lower == "true" || lower == "false"
	case KindString:
		return true
	case KindInt:
		// Only simple numeric values
		for _, c := range val {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	default:
		if typ.IsEnum() {
			// Check if val matches one of the enum values (case-insensitive)
			for _, ev := range typ.EnumValues() {
				if strings.EqualFold(ev, val) {
					return true
				}
			}
			return false
		}
	}
	return false
}

// normalizeDefault normalizes a default value to its canonical form.
func normalizeDefault(val string, typ *Type) string {
	switch typ.Kind {
	case KindBool:
		return strings.ToLower(val)
	default:
		if typ.IsEnum() {
			// Return the exact enum value (correct case)
			for _, ev := range typ.EnumValues() {
				if strings.EqualFold(ev, val) {
					return ev
				}
			}
		}
	}
	return val
}

// ---------------------------------------------------------------------------
// Single-optional-child promotion
// ---------------------------------------------------------------------------

// promoteSingleOptional walks the type graph and promotes the sole Optional
// property in a block to Required when the block has zero Required properties
// and zero non-SystemManaged ReadOnly properties.
//
// Rationale: if a block has only one settable property and nothing else, the
// user would be creating an empty block with no purpose. The property should
// be Required so the block is meaningful when present.
func promoteSingleOptional(typ *Type) {
	// The single-optional promotion counts the block's own settable properties. A
	// discriminated node's variant blocks are synthesized at emit and absent from
	// Properties, so counting its base props alone would misfire — only recurse
	// through a discriminated node, never promote against it.
	if typ.Kind == KindObject {
		var required, optional []*Property
		for _, prop := range typ.Properties {
			if prop.Flags.IsSystemManaged() {
				continue
			}
			if prop.Flags.IsRequired() {
				required = append(required, prop)
			} else if !prop.Flags.IsReadOnly() {
				optional = append(optional, prop)
			}
		}
		// Promote: exactly 1 optional, 0 required, any number of readonly
		if len(optional) == 1 && len(required) == 0 {
			optional[0].Flags = optional[0].Flags | FlagRequired
		}
	} else if typ.Kind != KindDiscriminated {
		return
	}

	// Recurse into all nested objects, discriminated blocks, and their variants.
	for _, prop := range typ.Properties {
		if prop.Type.Kind == KindObject || prop.Type.Kind == KindDiscriminated {
			promoteSingleOptional(prop.Type)
		}
		if prop.Type.Kind == KindArray && prop.Type.ElementType != nil &&
			(prop.Type.ElementType.Kind == KindObject || prop.Type.ElementType.Kind == KindDiscriminated) {
			promoteSingleOptional(prop.Type.ElementType)
		}
	}
	for _, value := range sortedKeys(typ.Variants) {
		promoteSingleOptional(typ.Variants[value])
	}
}

// ---------------------------------------------------------------------------
// Description-based validator extraction
// ---------------------------------------------------------------------------

// ARM resource ID pattern in descriptions: /subscriptions/{subscriptionId}/...
var armIDPattern = regexp.MustCompile(`/subscriptions/\{[^}]+\}/resourceGroups/`)

// Datetime format mentions
var datetimePattern = regexp.MustCompile(`(?i)(?:datetime|date/time)\s+format\s+['\"]?([^'\".\s]+)|(?:ISO\s*8601|RFC\s*3339)|format[:\s]+['"]?(yyyy-MM-dd[^'".\s]*)`)

// Numeric range: "must be greater than 0", "less than or equal to 5120", "between 1 and 100"
var (
	greaterThanPattern = regexp.MustCompile(`(?i)(?:must be|value is)\s+greater than\s+(?:or equal to\s+)?(\d+)`)
	lessThanPattern    = regexp.MustCompile(`(?i)(?:must be|value is)\s+less than\s+(?:or equal to\s+)?(\d+)`)
	betweenPattern     = regexp.MustCompile(`(?i)between\s+(\d+)\s+and\s+(\d+)`)
	minimumPattern     = regexp.MustCompile(`(?i)minimum\s+(?:value\s+)?(?:is\s+|of\s+)?(\d+)`)
	maximumPattern     = regexp.MustCompile(`(?i)maximum\s+(?:value\s+)?(?:is\s+|of\s+)?(\d+)`)
)

func extractDescriptionValidators(typ *Type) {
	if typ.Kind != KindObject && typ.Kind != KindDiscriminated {
		return
	}
	for _, prop := range typ.Properties {
		if prop.Flags.IsReadOnly() || prop.Flags.IsSystemManaged() {
			continue
		}
		if prop.Description == "" {
			continue
		}

		desc := prop.Description

		// ARM resource ID format
		if prop.Type.Kind == KindString && armIDPattern.MatchString(desc) {
			prop.Validators = append(prop.Validators, DescriptionValidator{
				Kind:    ValidatorArmResourceID,
				Message: "must be a valid ARM resource ID",
			})
		}

		// Datetime format
		if prop.Type.Kind == KindString {
			if m := datetimePattern.FindStringSubmatch(desc); m != nil {
				format := "yyyy-MM-ddTHH:mm:ssZ"
				for _, g := range m[1:] {
					if g != "" {
						format = g
						break
					}
				}
				prop.Validators = append(prop.Validators, DescriptionValidator{
					Kind:    ValidatorRegex,
					Pattern: datetimeFormatToRegex(format),
					Message: "must be in datetime format " + format,
				})
			}
		}

		// Numeric range
		if prop.Type.Kind == KindInt {
			v := extractNumericRange(desc)
			if v != nil {
				prop.Validators = append(prop.Validators, *v)
			}
		}

		// Recurse into nested objects, discriminated blocks, and array elements.
		if prop.Type.Kind == KindObject || prop.Type.Kind == KindDiscriminated {
			extractDescriptionValidators(prop.Type)
		}
		if prop.Type.Kind == KindArray && prop.Type.ElementType != nil &&
			(prop.Type.ElementType.Kind == KindObject || prop.Type.ElementType.Kind == KindDiscriminated) {
			extractDescriptionValidators(prop.Type.ElementType)
		}
	}
	// A discriminated block's variants are addressed as nested blocks; each is a
	// KindObject, so extract their validators through the same object path.
	for _, value := range sortedKeys(typ.Variants) {
		extractDescriptionValidators(typ.Variants[value])
	}
}

func extractNumericRange(desc string) *DescriptionValidator {
	var min, max *int64

	if m := betweenPattern.FindStringSubmatch(desc); m != nil {
		v1 := parseInt64(m[1])
		v2 := parseInt64(m[2])
		min = &v1
		max = &v2
	} else {
		if m := greaterThanPattern.FindStringSubmatch(desc); m != nil {
			v := parseInt64(m[1])
			min = &v
		}
		if m := minimumPattern.FindStringSubmatch(desc); m != nil && min == nil {
			v := parseInt64(m[1])
			min = &v
		}
		if m := lessThanPattern.FindStringSubmatch(desc); m != nil {
			v := parseInt64(m[1])
			max = &v
		}
		if m := maximumPattern.FindStringSubmatch(desc); m != nil && max == nil {
			v := parseInt64(m[1])
			max = &v
		}
	}

	if min == nil && max == nil {
		return nil
	}

	msg := "must be"
	if min != nil {
		msg += fmt.Sprintf(" >= %d", *min)
	}
	if min != nil && max != nil {
		msg += " and"
	}
	if max != nil {
		msg += fmt.Sprintf(" <= %d", *max)
	}

	return &DescriptionValidator{
		Kind:    ValidatorIntRange,
		Min:     min,
		Max:     max,
		Message: msg,
	}
}

func parseInt64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

// datetimeFormatToRegex converts a datetime format string to a basic regex.
func datetimeFormatToRegex(format string) string {
	// Common datetime format: yyyy-MM-ddTHH:mm:ssZ
	r := strings.NewReplacer(
		"yyyy", `\d{4}`,
		"MM", `\d{2}`,
		"dd", `\d{2}`,
		"HH", `\d{2}`,
		"mm", `\d{2}`,
		"ss", `\d{2}`,
		"T", `T`,
		"Z", `Z`,
	)
	return "^" + r.Replace(format) + "$"
}
