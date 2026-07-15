// Package mapper converts between the typed Terraform value tree (driven by a
// native-generated schema) and ARM JSON. It is generic — there is no per-resource
// Go struct. The bicep type graph supplies the authoritative ARM property names
// (camelCase), so snake_case ↔ camelCase is a lookup, never a reversal.
package mapper

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
	nativeschema "github.com/Azure/terraform-provider-azapi/internal/native/schema"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
	"github.com/Azure/terraform-provider-azapi/internal/services/dynamic"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Expand converts the body attributes of a typed object (plan or config) into
// ARM JSON, keyed by the ARM property names from the bicep type graph. Attributes
// that are null/unknown, system-managed, or absent from the type graph (the
// envelope: name, parent_id, id, timeouts) are skipped.
func Expand(obj types.Object, body *typegraph.Type) map[string]interface{} {
	if obj.IsNull() || obj.IsUnknown() || body == nil {
		return map[string]interface{}{}
	}
	if body.Kind == typegraph.KindDiscriminated {
		// A discriminated root flattens directly into the tagged ARM body; the
		// envelope attributes (name, parent_id, id, timeouts) are ignored here just
		// as they are for an object root (they are not in the type graph).
		out, _ := expandDiscriminated(obj, body)
		return out
	}
	if body.Kind != typegraph.KindObject {
		return map[string]interface{}{}
	}
	return expandObject(obj, body)
}

func expandObject(obj types.Object, t *typegraph.Type) map[string]interface{} {
	out := make(map[string]interface{})
	attrs := obj.Attributes()
	for armName, prop := range t.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		v, ok := attrs[naming.CamelToSnake(armName)]
		if !ok {
			continue
		}
		jv, ok := expandValue(v, prop.Type)
		if !ok {
			continue
		}
		out[armName] = jv
	}
	return out
}

// expandDiscriminated flattens a discriminated block object into a single tagged
// ARM object: the shared base properties, the properties of the one selected
// variant block, and the synthesized discriminator property set to that variant's
// value. Variant blocks are mutually exclusive (enforced by a ConfigValidator);
// the sorted-first non-null variant wins if config somehow carries more than one.
func expandDiscriminated(obj types.Object, t *typegraph.Type) (map[string]interface{}, bool) {
	out := make(map[string]interface{})
	attrs := obj.Attributes()

	for armName, prop := range t.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		if v, ok := attrs[naming.CamelToSnake(armName)]; ok {
			if jv, ok := expandValue(v, prop.Type); ok {
				out[armName] = jv
			}
		}
	}

	// Emit the selected variant's properties plus the synthesized discriminator.
	// Variants are sorted so a config that somehow set more than one (the
	// ConfigValidator normally forbids it) resolves deterministically.
	//
	// When NO variant block is set, the block is optional (AtMostOneOf) and
	// present for its base properties alone: emit just those, with no
	// discriminator tag. flattenDiscriminatedInto's discriminator-absent guard is
	// the exact inverse, so the base-only shape round-trips. A required block
	// (ExactlyOneOf) always has a variant, so an untagged body cannot reach ARM.
	for _, value := range sortedVariantKeys(t.Variants) {
		v, ok := attrs[naming.CamelToSnake(value)]
		if !ok || v == nil || v.IsNull() || v.IsUnknown() {
			continue
		}
		vo, ok := v.(types.Object)
		if !ok {
			continue
		}
		for k, jv := range expandObject(vo, t.Variants[value]) {
			out[k] = jv
		}
		if t.Discriminator != "" {
			out[t.Discriminator] = value
		}
		break // exactly one variant is active
	}

	return out, true
}

// discriminatedBaseIndex maps each settable base property of a discriminated type
// to its snake_case attribute name (the shared shape read/written on both the
// expand and flatten paths).
func discriminatedBaseIndex(t *typegraph.Type) map[string]*typegraph.Property {
	index := make(map[string]*typegraph.Property, len(t.Properties))
	for armName, prop := range t.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		index[naming.CamelToSnake(armName)] = prop
	}
	return index
}

// sortedVariantKeys returns a discriminated type's variant values in a stable
// order so expansion is deterministic when config carries more than one block.
func sortedVariantKeys(variants map[string]*typegraph.Type) []string {
	keys := make([]string, 0, len(variants))
	for value := range variants {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func expandValue(v attr.Value, t *typegraph.Type) (interface{}, bool) {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil, false
	}
	switch t.Kind {
	case typegraph.KindString, typegraph.KindStringLiteral:
		if s, ok := v.(types.String); ok {
			return s.ValueString(), true
		}
	case typegraph.KindBool:
		if b, ok := v.(types.Bool); ok {
			return b.ValueBool(), true
		}
	case typegraph.KindInt:
		if n, ok := v.(types.Int64); ok {
			return n.ValueInt64(), true
		}
	case typegraph.KindUnion:
		// enum and non-enum unions are emitted as strings
		if s, ok := v.(types.String); ok {
			return s.ValueString(), true
		}
	case typegraph.KindObject:
		if o, ok := v.(types.Object); ok {
			return expandObject(o, t), true
		}
	case typegraph.KindDiscriminated:
		if o, ok := v.(types.Object); ok {
			return expandDiscriminated(o, t)
		}
	case typegraph.KindArray:
		var elems []attr.Value
		switch c := v.(type) {
		case types.List:
			elems = c.Elements()
		case types.Set:
			elems = c.Elements()
		default:
			return nil, false
		}
		arr := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			if ev, ok := expandValue(e, t.ElementType); ok {
				arr = append(arr, ev)
			}
		}
		return arr, true
	case typegraph.KindMap:
		if m, ok := v.(types.Map); ok {
			elems := m.Elements()
			out := make(map[string]interface{}, len(elems))
			for k, e := range elems {
				if ev, ok := expandValue(e, t.ElementType); ok {
					out[k] = ev
				}
			}
			return out, true
		}
	default:
		// KindAny → dynamic attribute
		if d, ok := v.(types.Dynamic); ok {
			b, err := dynamic.ToJSON(d)
			if err == nil {
				var anyVal interface{}
				if json.Unmarshal(b, &anyVal) == nil {
					return anyVal, true
				}
			}
		}
	}
	return nil, false
}

// Flatten builds a state object value from an ARM JSON response. envelope provides
// values for the non-body (envelope) attributes — id, name, parent_id, timeouts —
// keyed by their Terraform attribute name; those win over the response. The result
// conforms to objType, which must be schema.Type() as a basetypes.ObjectType.
func Flatten(ctx context.Context, arm map[string]interface{}, objType basetypes.ObjectType, body *typegraph.Type, envelope map[string]attr.Value) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if body != nil && body.Kind == typegraph.KindDiscriminated {
		// A discriminated root: base props + the active variant come from the flat
		// tagged ARM body, the envelope attributes win over the response.
		return flattenDiscriminated(ctx, arm, objType, body, envelope)
	}

	// Index the body's settable properties by snake_case attribute name.
	index := map[string]*typegraph.Property{}
	if body != nil && body.Kind == typegraph.KindObject {
		for armName, prop := range body.Properties {
			if prop.Flags.IsSystemManaged() {
				continue
			}
			index[naming.CamelToSnake(armName)] = prop
		}
	}

	attrTypes := objType.AttributeTypes()
	values := make(map[string]attr.Value, len(attrTypes))
	for name, at := range attrTypes {
		if env, ok := envelope[name]; ok && env != nil {
			values[name] = env
			continue
		}
		prop, ok := index[name]
		if !ok {
			values[name] = nullOf(ctx, at)
			continue
		}
		armVal, present := arm[prop.Name]
		if !present || armVal == nil {
			values[name] = nullOf(ctx, at)
			continue
		}
		v, d := flattenValue(ctx, armVal, at, prop.Type)
		diags.Append(d...)
		values[name] = v
	}

	obj, d := types.ObjectValue(attrTypes, values)
	diags.Append(d...)
	return obj, diags
}

// flattenDiscriminated builds a discriminated block state object from a flat
// tagged ARM object. The discriminator property selects the active variant: the
// base properties and the active variant's properties are read from the same flat
// ARM object; every inactive variant block is null. When the discriminator is
// absent from the response, no variant is populated (all variant blocks null).
func flattenDiscriminated(ctx context.Context, arm map[string]interface{}, objType basetypes.ObjectType, disc *typegraph.Type, envelope map[string]attr.Value) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	baseIndex := discriminatedBaseIndex(disc)
	variantIndex := map[string]string{} // snake attr name -> discriminator value
	for value := range disc.Variants {
		variantIndex[naming.CamelToSnake(value)] = value
	}
	kind, _ := arm[disc.Discriminator].(string)

	attrTypes := objType.AttributeTypes()
	values := make(map[string]attr.Value, len(attrTypes))
	for name, at := range attrTypes {
		if env, ok := envelope[name]; ok && env != nil {
			values[name] = env // envelope attr (root only); nil map for a nested block
			continue
		}
		if value, ok := variantIndex[name]; ok {
			// The active variant reads its properties from the same flat ARM object;
			// inactive variants are null.
			if value == kind {
				if vt, ok := at.(basetypes.ObjectType); ok {
					v, d := Flatten(ctx, arm, vt, disc.Variants[value], nil)
					diags.Append(d...)
					values[name] = v
					continue
				}
			}
			values[name] = nullOf(ctx, at)
			continue
		}
		prop, ok := baseIndex[name]
		if !ok {
			values[name] = nullOf(ctx, at)
			continue
		}
		armVal, present := arm[prop.Name]
		if !present || armVal == nil {
			values[name] = nullOf(ctx, at)
			continue
		}
		v, d := flattenValue(ctx, armVal, at, prop.Type)
		diags.Append(d...)
		values[name] = v
	}

	obj, d := types.ObjectValue(attrTypes, values)
	diags.Append(d...)
	return obj, diags
}

func flattenValue(ctx context.Context, armVal interface{}, at attr.Type, bicep *typegraph.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ot, ok := at.(basetypes.ObjectType); ok && bicep != nil && bicep.Kind == typegraph.KindDiscriminated {
		if m, ok := armVal.(map[string]interface{}); ok {
			return flattenDiscriminated(ctx, m, ot, bicep, nil)
		}
		return nullOf(ctx, at), diags
	}
	switch t := at.(type) {
	case basetypes.StringType:
		if s, ok := armVal.(string); ok {
			return types.StringValue(s), diags
		}
	case basetypes.BoolType:
		if b, ok := armVal.(bool); ok {
			return types.BoolValue(b), diags
		}
	case basetypes.Int64Type:
		switch n := armVal.(type) {
		case float64:
			return types.Int64Value(int64(n)), diags
		case int64:
			return types.Int64Value(n), diags
		}
	case basetypes.ObjectType:
		if m, ok := armVal.(map[string]interface{}); ok {
			var nested *typegraph.Type
			if bicep != nil && bicep.Kind == typegraph.KindObject {
				nested = bicep
			}
			return Flatten(ctx, m, t, nested, nil)
		}
	case basetypes.ListType:
		if arr, ok := armVal.([]interface{}); ok {
			elemType := t.ElementType()
			elems, d := flattenArrayElements(ctx, arr, elemType, bicep)
			diags.Append(d...)
			lv, d := types.ListValue(elemType, elems)
			diags.Append(d...)
			return lv, diags
		}
	case basetypes.SetType:
		if arr, ok := armVal.([]interface{}); ok {
			elemType := t.ElementType()
			elems, d := flattenArrayElements(ctx, arr, elemType, bicep)
			diags.Append(d...)
			sv, d := types.SetValue(elemType, elems)
			diags.Append(d...)
			return sv, diags
		}
	case basetypes.MapType:
		if m, ok := armVal.(map[string]interface{}); ok {
			elemType := t.ElementType()
			var elemBicep *typegraph.Type
			if bicep != nil && bicep.Kind == typegraph.KindMap {
				elemBicep = bicep.ElementType
			}
			elems := make(map[string]attr.Value, len(m))
			for k, item := range m {
				ev, d := flattenValue(ctx, item, elemType, elemBicep)
				diags.Append(d...)
				elems[canonicalMapKey(k)] = ev
			}
			mv, d := types.MapValue(elemType, elems)
			diags.Append(d...)
			return mv, diags
		}
	case basetypes.DynamicType:
		b, err := json.Marshal(armVal)
		if err == nil {
			if d, derr := dynamic.FromJSONImplied(b); derr == nil {
				return d, diags
			}
		}
	}

	// Type mismatch or unsupported → null of the target type.
	return nullOf(ctx, at), diags
}

func flattenArrayElements(ctx context.Context, arr []interface{}, elemType attr.Type, bicep *typegraph.Type) ([]attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	var elemBicep *typegraph.Type
	if bicep != nil && bicep.Kind == typegraph.KindArray {
		elemBicep = bicep.ElementType
	}
	elems := make([]attr.Value, 0, len(arr))
	for _, item := range arr {
		ev, d := flattenValue(ctx, item, elemType, elemBicep)
		diags.Append(d...)
		elems = append(elems, ev)
	}
	return elems, diags
}

// nullOf returns the null value for an arbitrary attribute type.
func nullOf(ctx context.Context, at attr.Type) attr.Value {
	v, err := at.ValueFromTerraform(ctx, tftypes.NewValue(at.TerraformType(ctx), nil))
	if err != nil {
		// Should not happen for well-formed schema types.
		return types.DynamicNull()
	}
	return v
}

// ResolveUnknowns recursively replaces any remaining unknown value with the null
// of its type. Apply results must be fully known: Optional+Computed attributes the
// user did not set and the ARM response did not populate stay unknown after
// FlattenInto; nulling them satisfies the framework while leaving known values and
// known siblings intact.
func ResolveUnknowns(ctx context.Context, v attr.Value) attr.Value {
	if v == nil {
		return v
	}
	if v.IsUnknown() {
		return nullOf(ctx, v.Type(ctx))
	}
	if v.IsNull() {
		return v
	}
	switch t := v.(type) {
	case types.Object:
		attrs := t.Attributes()
		nv := make(map[string]attr.Value, len(attrs))
		for k, av := range attrs {
			nv[k] = ResolveUnknowns(ctx, av)
		}
		obj, _ := types.ObjectValue(t.AttributeTypes(ctx), nv)
		return obj
	case types.List:
		elems := t.Elements()
		ne := make([]attr.Value, len(elems))
		for i, e := range elems {
			ne[i] = ResolveUnknowns(ctx, e)
		}
		lv, _ := types.ListValue(t.ElementType(ctx), ne)
		return lv
	case types.Set:
		elems := t.Elements()
		ne := make([]attr.Value, len(elems))
		for i, e := range elems {
			ne[i] = ResolveUnknowns(ctx, e)
		}
		sv, _ := types.SetValue(t.ElementType(ctx), ne)
		return sv
	case types.Map:
		elems := t.Elements()
		ne := make(map[string]attr.Value, len(elems))
		for k, e := range elems {
			ne[k] = ResolveUnknowns(ctx, e)
		}
		mv, _ := types.MapValue(t.ElementType(ctx), ne)
		return mv
	default:
		return v
	}
}

// FlattenInto refreshes body attributes of an existing state object from an ARM
// JSON response, preserving the envelope (name, parent_id, id, timeouts) and any
// body attribute the response omits. It is used for read/import refresh, so normal
// echoed properties overwrite prior state while sensitive/write-only and semantic
// location values are kept stable.
func FlattenInto(ctx context.Context, arm map[string]interface{}, base types.Object, body *typegraph.Type) (types.Object, diag.Diagnostics) {
	return flattenInto(ctx, arm, base, body, false)
}

// FlattenApplyInto refreshes the planned object from an ARM write response for
// Create/Update. Terraform requires apply results to match every planned known
// value exactly; values configured by the practitioner cannot be replaced with
// Azure's echo/default representation during apply. Unknown planned values may
// still be populated from the response.
func FlattenApplyInto(ctx context.Context, arm map[string]interface{}, base types.Object, body *typegraph.Type) (types.Object, diag.Diagnostics) {
	return flattenInto(ctx, arm, base, body, true)
}

func flattenInto(ctx context.Context, arm map[string]interface{}, base types.Object, body *typegraph.Type, preserveKnown bool) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if base.IsNull() || base.IsUnknown() {
		return base, diags
	}

	attrTypes := base.AttributeTypes(ctx)
	values := make(map[string]attr.Value, len(attrTypes))
	for k, v := range base.Attributes() {
		values[k] = v
	}

	if body != nil && body.Kind == typegraph.KindDiscriminated {
		// A discriminated root: refresh base props + the active variant from the flat
		// ARM body, preserving the envelope (name, parent_id, id, timeouts) already
		// copied above. flattenDiscriminatedInto leaves non-base/variant attrs intact.
		v, d := flattenDiscriminatedInto(ctx, arm, base, body, preserveKnown)
		diags.Append(d...)
		if obj, ok := v.(types.Object); ok {
			return obj, diags
		}
		return base, diags
	}

	if body != nil && body.Kind == typegraph.KindObject {
		for armName, prop := range body.Properties {
			if prop.Flags.IsSystemManaged() {
				continue
			}
			name := naming.CamelToSnake(armName)
			at, ok := attrTypes[name]
			if !ok {
				continue
			}
			armVal, present := arm[prop.Name]
			if !present || armVal == nil {
				continue // keep the base value
			}
			if preserveKnown && shouldPreserveKnownApplyValue(values[name]) {
				continue
			}
			if shouldPreserveSensitiveValue(prop) {
				continue
			}
			if shouldPreserveEquivalentLocation(prop, values[name], armVal) {
				continue
			}
			v, d := flattenValueInto(ctx, armVal, values[name], at, prop.Type, preserveKnown)
			diags.Append(d...)
			values[name] = v
		}
	}

	obj, d := types.ObjectValue(attrTypes, values)
	diags.Append(d...)
	return obj, diags
}

// flattenValueInto is the base-aware counterpart of flattenValue: for a nested
// object it recurses through flattenInto, threading the corresponding base value
// so fields the response omits are preserved at every depth. Non-object kinds have
// no per-element base to thread and delegate to flattenValue unchanged.
func flattenValueInto(ctx context.Context, armVal interface{}, base attr.Value, at attr.Type, bicep *typegraph.Type, preserveKnown bool) (attr.Value, diag.Diagnostics) {
	if _, ok := at.(basetypes.ObjectType); ok && bicep != nil && bicep.Kind == typegraph.KindDiscriminated {
		if m, ok := armVal.(map[string]interface{}); ok {
			if baseObj, ok := base.(types.Object); ok && !baseObj.IsNull() && !baseObj.IsUnknown() {
				return flattenDiscriminatedInto(ctx, m, baseObj, bicep, preserveKnown)
			}
		}
	}
	if _, ok := at.(basetypes.ObjectType); ok {
		if m, ok := armVal.(map[string]interface{}); ok {
			if baseObj, ok := base.(types.Object); ok && !baseObj.IsNull() && !baseObj.IsUnknown() {
				var nested *typegraph.Type
				if bicep != nil && bicep.Kind == typegraph.KindObject {
					nested = bicep
				}
				return flattenInto(ctx, m, baseObj, nested, preserveKnown)
			}
		}
	}
	if mt, ok := at.(basetypes.MapType); ok {
		if m, ok := armVal.(map[string]interface{}); ok {
			if baseMap, ok := base.(types.Map); ok && !baseMap.IsNull() && !baseMap.IsUnknown() {
				return flattenMapInto(ctx, m, baseMap, mt, bicep, preserveKnown)
			}
		}
	}
	return flattenValue(ctx, armVal, at, bicep)
}

// flattenDiscriminatedInto is the base-aware refresh of a discriminated block: it
// refreshes the base properties and the active variant (threading the prior state
// so omitted/sensitive fields are preserved), and clears every inactive variant
// block whenever the discriminator selects a different (or no longer matching)
// variant. When the response omits the discriminator, the prior block is returned
// unchanged so a discriminator-less GET does not wipe a configured variant.
func flattenDiscriminatedInto(ctx context.Context, arm map[string]interface{}, base types.Object, disc *typegraph.Type, preserveKnown bool) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	kind, hasKind := arm[disc.Discriminator].(string)
	if !hasKind || kind == "" {
		return base, diags // no discriminator in response: keep prior block intact
	}

	attrTypes := base.AttributeTypes(ctx)
	baseAttrs := base.Attributes()
	values := make(map[string]attr.Value, len(attrTypes))
	for k, v := range baseAttrs {
		values[k] = v
	}

	baseIndex := discriminatedBaseIndex(disc)

	for value := range disc.Variants {
		name := naming.CamelToSnake(value)
		at, ok := attrTypes[name]
		if !ok {
			continue
		}
		if value != kind {
			values[name] = nullOf(ctx, at) // inactive variant clears
			continue
		}
		vt, ok := at.(basetypes.ObjectType)
		if !ok {
			continue
		}
		if priorObj, ok := baseAttrs[name].(types.Object); ok && !priorObj.IsNull() && !priorObj.IsUnknown() {
			v, d := flattenInto(ctx, arm, priorObj, disc.Variants[value], preserveKnown)
			diags.Append(d...)
			values[name] = v
		} else {
			v, d := Flatten(ctx, arm, vt, disc.Variants[value], nil)
			diags.Append(d...)
			values[name] = v
		}
	}

	for name, prop := range baseIndex {
		at, ok := attrTypes[name]
		if !ok {
			continue
		}
		armVal, present := arm[prop.Name]
		if !present || armVal == nil {
			continue // keep base value
		}
		if preserveKnown && shouldPreserveKnownApplyValue(values[name]) {
			continue
		}
		if shouldPreserveSensitiveValue(prop) {
			continue
		}
		v, d := flattenValueInto(ctx, armVal, values[name], at, prop.Type, preserveKnown)
		diags.Append(d...)
		values[name] = v
	}

	obj, d := types.ObjectValue(attrTypes, values)
	diags.Append(d...)
	return obj, diags
}

// flattenMapInto rebuilds an open map (e.g. identity.user_assigned_identities,
// tags) from the ARM response while preserving each prior key's exact casing when
// the response echoes a case-insensitively-equal key. ARM resource-ID map keys are
// case-insensitive, but Terraform map keys are case-sensitive, so a server that
// lowercases a segment (e.g. Web returns .../resourcegroups/... for a key configured
// as .../resourceGroups/...) would otherwise churn the map as a drop+add every plan.
// Each value is threaded through flattenValueInto against its prior element so nested
// object fields the response omits are still preserved. New keys absent from the base
// keep the response casing.
func flattenMapInto(ctx context.Context, arm map[string]interface{}, base types.Map, mt basetypes.MapType, bicep *typegraph.Type, preserveKnown bool) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := mt.ElementType()
	var elemBicep *typegraph.Type
	if bicep != nil && bicep.Kind == typegraph.KindMap {
		elemBicep = bicep.ElementType
	}
	baseElems := base.Elements()
	lowerToBaseKey := make(map[string]string, len(baseElems))
	for bk := range baseElems {
		lowerToBaseKey[strings.ToLower(bk)] = bk
	}
	elems := make(map[string]attr.Value, len(arm))
	for k, item := range arm {
		var outKey string
		var baseVal attr.Value
		if bk, ok := lowerToBaseKey[strings.ToLower(k)]; ok {
			outKey = bk
			baseVal = baseElems[bk]
		} else {
			// No prior key to preserve the practitioner's exact casing against;
			// fall back to the canonical ARM-ID casing so an imported/refreshed
			// key matches the case-canonical ID written in config.
			outKey = canonicalMapKey(k)
		}
		ev, d := flattenValueInto(ctx, item, baseVal, elemType, elemBicep, preserveKnown)
		diags.Append(d...)
		elems[outKey] = ev
	}
	mv, d := types.MapValue(elemType, elems)
	diags.Append(d...)
	return mv, diags
}

// canonicalMapKey normalizes an open-map key that is an ARM resource ID to its
// canonical Azure casing (e.g. a lowercased ".../resourcegroups/..." segment ->
// ".../resourceGroups/..."). Some resource providers echo an assigned resource ID
// with altered segment casing; ARM IDs are case-insensitive but Terraform map keys
// are not, so without this an imported or freshly-flattened state key would drift
// from the case-canonical ID the practitioner writes in config (Terraform resource
// ID references are always canonical). Keys that are not resource IDs — e.g. tag
// names — fail to parse and are returned verbatim.
func canonicalMapKey(k string) string {
	if strings.HasPrefix(k, "/") {
		if id, err := arm.ParseResourceID(k); err == nil {
			return id.String()
		}
	}
	return k
}

func shouldPreserveKnownApplyValue(v attr.Value) bool {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return false
	}
	_, isObject := v.(types.Object)
	return !isObject
}

// isLocationFieldName reports whether an ARM property carries an Azure region name
// (e.g. "eastus2" vs "East US 2"): the top-level "location" and the nested geo
// "locationName" (Cosmos geo-replication, and similar structs). Both are echoed by
// Azure in a normalized display form and must be preserved when normalize-equal.
func isLocationFieldName(name string) bool {
	return name == "location" || name == "locationName"
}

func shouldPreserveEquivalentLocation(prop *typegraph.Property, base attr.Value, armVal interface{}) bool {
	if prop == nil || !isLocationFieldName(prop.Name) || prop.Type == nil || prop.Type.Kind != typegraph.KindString {
		return false
	}
	baseString, ok := base.(types.String)
	if !ok || baseString.IsNull() || baseString.IsUnknown() {
		return false
	}
	armString, ok := armVal.(string)
	if !ok {
		return false
	}
	return nativeschema.NormalizeLocation(baseString.ValueString()) == nativeschema.NormalizeLocation(armString)
}

func shouldPreserveSensitiveValue(prop *typegraph.Property) bool {
	return prop != nil && (prop.Sensitive || prop.Flags.IsWriteOnly())
}
