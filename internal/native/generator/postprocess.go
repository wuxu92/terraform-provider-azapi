package generator

import (
	"fmt"
	"regexp"
	"strings"
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
			applyEnvelopeDefaults(def)
		}
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
	if typ.Kind != KindObject {
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

		// Recurse into nested objects
		if prop.Type.Kind == KindObject {
			extractDefaults(prop.Type)
		}
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
	if typ.Kind != KindObject {
		return
	}

	var required, optional, readonly []*Property
	for _, prop := range typ.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		if prop.Flags.IsRequired() {
			required = append(required, prop)
		} else if prop.Flags.IsReadOnly() {
			readonly = append(readonly, prop)
		} else {
			optional = append(optional, prop)
		}
	}

	// Promote: exactly 1 optional, 0 required, any number of readonly
	if len(optional) == 1 && len(required) == 0 {
		optional[0].Flags = optional[0].Flags | FlagRequired
	}

	// Recurse into all nested objects
	for _, prop := range typ.Properties {
		if prop.Type.Kind == KindObject {
			promoteSingleOptional(prop.Type)
		}
		if prop.Type.Kind == KindArray && prop.Type.ElementType != nil && prop.Type.ElementType.Kind == KindObject {
			promoteSingleOptional(prop.Type.ElementType)
		}
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
	if typ.Kind != KindObject {
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

		// Recurse into nested objects
		if prop.Type.Kind == KindObject {
			extractDescriptionValidators(prop.Type)
		}
		if prop.Type.Kind == KindArray && prop.Type.ElementType != nil && prop.Type.ElementType.Kind == KindObject {
			extractDescriptionValidators(prop.Type.ElementType)
		}
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
