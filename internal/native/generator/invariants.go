package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// FlagViolation is a single terraform-plugin-framework attribute-flag invariant
// breach at an ARM dot path within a resource body.
type FlagViolation struct {
	Path    string
	Message string
}

func (v FlagViolation) String() string { return v.Path + ": " + v.Message }

// CheckFlagInvariants walks the post-processed body (after azwise + customizers)
// and returns every attribute-flag invariant violation. An empty slice means the
// body is framework-valid.
//
// These checks run at generation time because neither the bicep types nor the
// framework's startup ValidateImplementation catch them all:
//
//   - Restraint 1: a property with a Default must be Optional+Computed — never
//     Required and never read-only/computed-only. The framework requires Computed
//     whenever a Default is set (string_attribute.go: a non-Computed Default is a
//     startup diagnostic) and forbids Required+Default; the emitter silently drops
//     a Default that lands on a Required/computed-only property, so the knowledge
//     is lost without warning. (This is the correct framing of "a property with a
//     default does not need the Computed *safe-default* — but the framework still
//     requires the Computed *flag*".)
//
//   - Restraint 2: a Default must satisfy the attribute's own validators. The
//     framework never cross-checks this, so an enum/range default outside its
//     allowed set makes the provider reject the value it itself supplies when the
//     field is omitted from config.
func CheckFlagInvariants(def *typegraph.ResourceDefinition) []FlagViolation {
	if def == nil || def.Body == nil {
		return nil
	}
	var out []FlagViolation
	checkObjectInvariants(def.Body, "", &out)
	return out
}

// checkObjectInvariants recurses through an object's settable properties and
// object-array element schemas, checking each.
func checkObjectInvariants(obj *typegraph.Type, prefix string, out *[]FlagViolation) {
	if obj == nil || obj.Kind != typegraph.KindObject {
		return
	}
	for name, prop := range obj.Properties {
		if prop == nil || prop.Flags.IsSystemManaged() {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		checkPropInvariants(prop, path, out)

		switch {
		case prop.Type != nil && prop.Type.Kind == typegraph.KindObject:
			checkObjectInvariants(prop.Type, path, out)
		case prop.Type != nil && prop.Type.Kind == typegraph.KindArray &&
			prop.Type.ElementType != nil && prop.Type.ElementType.Kind == typegraph.KindObject:
			checkObjectInvariants(prop.Type.ElementType, path+"[]", out)
		}
	}
}

func checkPropInvariants(prop *typegraph.Property, path string, out *[]FlagViolation) {
	if prop.DefaultValue == "" {
		return
	}
	add := func(msg string) { *out = append(*out, FlagViolation{Path: path, Message: msg}) }

	// Restraint 1: a Default requires Optional+Computed.
	if prop.Flags.IsRequired() {
		add(fmt.Sprintf("has a Default (%q) but is Required; the framework forbids Required+Default and the default is dropped", prop.DefaultValue))
	}
	if typegraph.EffectiveComputed(prop) {
		add(fmt.Sprintf("has a Default (%q) but is read-only/computed; the default is dropped", prop.DefaultValue))
	}

	// Restraint 2: a Default must satisfy the attribute's own validators. The
	// emitted OneOf is case-sensitive (stringvalidator.OneOf), so membership is
	// checked exactly.
	if prop.Type != nil && prop.Type.IsEnum() {
		if vals := prop.Type.EnumValues(); !contains(vals, prop.DefaultValue) {
			add(fmt.Sprintf("Default %q is not one of the enum values %v", prop.DefaultValue, vals))
		}
	}
	for _, v := range prop.Validators {
		switch v.Kind {
		case typegraph.ValidatorStringOneOf:
			if !contains(v.Allowed, prop.DefaultValue) {
				add(fmt.Sprintf("Default %q is not one of the allowed values %v", prop.DefaultValue, v.Allowed))
			}
		case typegraph.ValidatorIntRange:
			n, err := strconv.ParseInt(prop.DefaultValue, 10, 64)
			if err != nil {
				add(fmt.Sprintf("Default %q is not an integer but the field has an int-range validator", prop.DefaultValue))
			} else if (v.Min != nil && n < *v.Min) || (v.Max != nil && n > *v.Max) {
				add(fmt.Sprintf("Default %d is outside the validated range %s", n, rangeStr(v.Min, v.Max)))
			}
		case typegraph.ValidatorStringLength:
			l := int64(len(prop.DefaultValue))
			if (v.Min != nil && l < *v.Min) || (v.Max != nil && l > *v.Max) {
				add(fmt.Sprintf("Default %q has length %d outside the validated length %s", prop.DefaultValue, l, rangeStr(v.Min, v.Max)))
			}
		}
	}
}

func contains(set []string, val string) bool {
	for _, s := range set {
		if s == val {
			return true
		}
	}
	return false
}

func rangeStr(min, max *int64) string {
	switch {
	case min != nil && max != nil:
		return fmt.Sprintf("[%d, %d]", *min, *max)
	case min != nil:
		return fmt.Sprintf(">= %d", *min)
	case max != nil:
		return fmt.Sprintf("<= %d", *max)
	}
	return "(unbounded)"
}

// FormatViolations renders violations as one indented line each.
func FormatViolations(vs []FlagViolation) string {
	var b strings.Builder
	for _, v := range vs {
		b.WriteString("  - ")
		b.WriteString(v.String())
		b.WriteString("\n")
	}
	return b.String()
}
