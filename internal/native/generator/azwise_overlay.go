package generator

import (
	"fmt"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/azure/azwise"
)

// ApplyAzwise overlays curated AzureRM-derived knowledge (azwise) onto a parsed
// type graph. azwise rules are hand-verified and more reliable than the
// bicep-flag/description-mined heuristics, so they take precedence:
//
//   - ForceNew paths      → prop.ForceNew (RequiresReplace plan modifier)
//   - ComputedFields      → prop.ForceComputed (Computed-only)
//   - SensitiveFields     → prop.Sensitive (Sensitive: true)
//   - DefaultValues       → prop.DefaultValue (overrides description-mined default)
//   - StringRules/IntRules→ prop.Validators (appended)
//
// It is a no-op when no azwise knowledge is registered for the resource type.
// Property paths are ARM dot paths (e.g. "properties.accessTier", "sku.name");
// paths that do not resolve in the type graph (e.g. ListKeys-only sensitive
// fields that are not part of the body) are silently skipped.
func ApplyAzwise(def *ResourceDefinition) {
	if def == nil || def.Body == nil || def.Body.Kind != KindObject {
		return
	}

	armType := def.Name
	apiVersion := def.APIVersion
	if at := strings.Index(armType, "@"); at >= 0 {
		armType = armType[:at]
	}

	azwise.EnsureRegistered()
	k := azwise.Get(armType, apiVersion)
	if k == nil {
		return
	}

	// Flag-level overlays available on the base interface.
	for _, p := range k.GetComputedFields() {
		if prop := navigate(def.Body, p); prop != nil {
			prop.ForceComputed = true
		}
	}
	for _, dv := range k.GetDefaultValues() {
		if dv.Value == nil {
			continue // Optional+Computed with no explicit default — leave to schema
		}
		if prop := navigate(def.Body, dv.PropertyPath); prop != nil {
			if s, ok := formatAzwiseDefault(dv.Value, prop.Type); ok {
				prop.DefaultValue = s
			}
		}
	}

	// Structured rules via the SchemaKnowledge accessor interface.
	sk, ok := k.(azwise.SchemaKnowledge)
	if !ok {
		return
	}
	for _, p := range sk.GetForceNewPaths() {
		if prop := navigate(def.Body, p); prop != nil {
			prop.ForceNew = true
		}
	}
	for _, p := range sk.GetSensitiveFields() {
		if prop := navigate(def.Body, p); prop != nil {
			prop.Sensitive = true
		}
	}
	for _, r := range sk.GetStringRules() {
		if r.PropertyPath == "" || strings.Contains(r.PropertyPath, "[*]") {
			continue // name rule or array-element rule — not a single attribute
		}
		prop := navigate(def.Body, r.PropertyPath)
		if prop == nil || prop.Type.Kind != KindString {
			continue
		}
		applyStringRule(prop, r)
	}
	for _, r := range sk.GetIntRules() {
		if strings.Contains(r.PropertyPath, "[*]") {
			continue
		}
		prop := navigate(def.Body, r.PropertyPath)
		if prop == nil || prop.Type.Kind != KindInt {
			continue
		}
		if r.MinValue != nil || r.MaxValue != nil {
			dropValidators(prop, ValidatorIntRange)
			prop.Validators = append(prop.Validators, DescriptionValidator{
				Kind:    ValidatorIntRange,
				Min:     r.MinValue,
				Max:     r.MaxValue,
				Message: r.Message,
			})
		}
	}
}

// navigate resolves an ARM dot path (e.g. "properties.networkAcls.defaultAction")
// to the Property it names, walking ObjectType properties and descending through
// single arrays of objects. Returns nil if any segment is missing.
func navigate(body *Type, path string) *Property {
	cur := body
	segments := strings.Split(path, ".")
	for i, seg := range segments {
		seg = strings.TrimSuffix(seg, "[*]")
		if cur == nil || cur.Kind != KindObject {
			return nil
		}
		prop, ok := cur.Properties[seg]
		if !ok {
			return nil
		}
		if i == len(segments)-1 {
			return prop
		}
		// Descend: into the object, or into an array's element object.
		switch prop.Type.Kind {
		case KindObject:
			cur = prop.Type
		case KindArray:
			cur = prop.Type.ElementType
		default:
			return nil
		}
	}
	return nil
}

// applyStringRule appends OneOf / length / regex validators from an azwise StringRule.
// AllowedValues are skipped when the property is already a bicep enum (the emitter
// emits a OneOf for enums) to avoid duplicate validators.
func applyStringRule(prop *Property, r azwise.StringRule) {
	if len(r.AllowedValues) > 0 && !prop.Type.IsEnum() {
		dropValidators(prop, ValidatorStringOneOf)
		prop.Validators = append(prop.Validators, DescriptionValidator{
			Kind:    ValidatorStringOneOf,
			Allowed: r.AllowedValues,
			Message: r.Message,
		})
	}
	if r.MinLength > 0 || r.MaxLength > 0 {
		var min, max *int64
		if r.MinLength > 0 {
			v := int64(r.MinLength)
			min = &v
		}
		if r.MaxLength > 0 {
			v := int64(r.MaxLength)
			max = &v
		}
		dropValidators(prop, ValidatorStringLength)
		prop.Validators = append(prop.Validators, DescriptionValidator{
			Kind:    ValidatorStringLength,
			Min:     min,
			Max:     max,
			Message: r.Message,
		})
	}
	if r.Regex != "" {
		dropValidators(prop, ValidatorRegex)
		prop.Validators = append(prop.Validators, DescriptionValidator{
			Kind:    ValidatorRegex,
			Pattern: r.Regex,
			Message: r.Message,
		})
	}
}

// dropValidators removes every validator of the given kind from the property so
// an azwise rule replaces (rather than stacks onto) a description-mined or
// bicep-flag validator of the same kind. azwise rules are authoritative, so a
// duplicate or differing heuristic validator must not survive alongside them.
func dropValidators(prop *Property, kind ValidatorKind) {
	kept := prop.Validators[:0]
	for _, v := range prop.Validators {
		if v.Kind != kind {
			kept = append(kept, v)
		}
	}
	prop.Validators = kept
}

// formatAzwiseDefault renders an azwise default value (ARM wire format:
// string / bool / float64) as the Go literal the emitter expects for the
// property type. Returns ok=false when the value is incompatible.
func formatAzwiseDefault(v interface{}, typ *Type) (string, bool) {
	switch typ.Kind {
	case KindBool:
		if b, ok := v.(bool); ok {
			if b {
				return "true", true
			}
			return "false", true
		}
	case KindInt:
		switch n := v.(type) {
		case float64:
			return fmt.Sprintf("%d", int64(n)), true
		case int:
			return fmt.Sprintf("%d", n), true
		case int64:
			return fmt.Sprintf("%d", n), true
		}
	case KindString:
		if s, ok := v.(string); ok {
			return s, true
		}
	default:
		if typ.IsEnum() {
			if s, ok := v.(string); ok {
				return s, true
			}
		}
	}
	return "", false
}
