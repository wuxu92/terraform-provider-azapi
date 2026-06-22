// Package mapper converts between the typed Terraform value tree (driven by a
// native-generated schema) and ARM JSON. It is generic — there is no per-resource
// Go struct. The bicep type graph supplies the authoritative ARM property names
// (camelCase), so snake_case ↔ camelCase is a lookup, never a reversal.
package mapper

import (
	"context"
	"encoding/json"

	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
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
func Expand(obj types.Object, body *generator.Type) map[string]interface{} {
	if obj.IsNull() || obj.IsUnknown() || body == nil || body.Kind != generator.KindObject {
		return map[string]interface{}{}
	}
	return expandObject(obj, body)
}

func expandObject(obj types.Object, t *generator.Type) map[string]interface{} {
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

func expandValue(v attr.Value, t *generator.Type) (interface{}, bool) {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil, false
	}
	switch t.Kind {
	case generator.KindString, generator.KindStringLiteral:
		if s, ok := v.(types.String); ok {
			return s.ValueString(), true
		}
	case generator.KindBool:
		if b, ok := v.(types.Bool); ok {
			return b.ValueBool(), true
		}
	case generator.KindInt:
		if n, ok := v.(types.Int64); ok {
			return n.ValueInt64(), true
		}
	case generator.KindUnion:
		// enum and non-enum unions are emitted as strings
		if s, ok := v.(types.String); ok {
			return s.ValueString(), true
		}
	case generator.KindObject:
		if o, ok := v.(types.Object); ok {
			return expandObject(o, t), true
		}
	case generator.KindArray:
		if l, ok := v.(types.List); ok {
			elems := l.Elements()
			arr := make([]interface{}, 0, len(elems))
			for _, e := range elems {
				if ev, ok := expandValue(e, t.ElementType); ok {
					arr = append(arr, ev)
				}
			}
			return arr, true
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
func Flatten(ctx context.Context, arm map[string]interface{}, objType basetypes.ObjectType, body *generator.Type, envelope map[string]attr.Value) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Index the body's settable properties by snake_case attribute name.
	index := map[string]*generator.Property{}
	if body != nil && body.Kind == generator.KindObject {
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

func flattenValue(ctx context.Context, armVal interface{}, at attr.Type, bicep *generator.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

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
			var nested *generator.Type
			if bicep != nil && bicep.Kind == generator.KindObject {
				nested = bicep
			}
			return Flatten(ctx, m, t, nested, nil)
		}
	case basetypes.ListType:
		if arr, ok := armVal.([]interface{}); ok {
			elemType := t.ElementType()
			var elemBicep *generator.Type
			if bicep != nil && bicep.Kind == generator.KindArray {
				elemBicep = bicep.ElementType
			}
			elems := make([]attr.Value, 0, len(arr))
			for _, item := range arr {
				ev, d := flattenValue(ctx, item, elemType, elemBicep)
				diags.Append(d...)
				elems = append(elems, ev)
			}
			lv, d := types.ListValue(elemType, elems)
			diags.Append(d...)
			return lv, diags
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
	default:
		return v
	}
}

// FlattenInto refreshes the body attributes of an existing object (plan for
// create, prior state for read/update) from an ARM JSON response, preserving the
// envelope (name, parent_id, id, timeouts) and any body attribute the response
// omits — recursively, so a nested optional the user set but the server drops keeps
// the planned value instead of collapsing to null (which the framework rejects as
// an inconsistent apply result). Only attributes present in the response are
// overwritten, avoiding wipes of write-only fields the server never echoes.
func FlattenInto(ctx context.Context, arm map[string]interface{}, base types.Object, body *generator.Type) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if base.IsNull() || base.IsUnknown() {
		return base, diags
	}

	attrTypes := base.AttributeTypes(ctx)
	values := make(map[string]attr.Value, len(attrTypes))
	for k, v := range base.Attributes() {
		values[k] = v
	}

	if body != nil && body.Kind == generator.KindObject {
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
			v, d := flattenValueInto(ctx, armVal, values[name], at, prop.Type)
			diags.Append(d...)
			values[name] = v
		}
	}

	obj, d := types.ObjectValue(attrTypes, values)
	diags.Append(d...)
	return obj, diags
}

// flattenValueInto is the base-aware counterpart of flattenValue: for a nested
// object it recurses through FlattenInto, threading the corresponding base value
// so fields the response omits are preserved at every depth. Non-object kinds have
// no per-element base to thread and delegate to flattenValue unchanged.
func flattenValueInto(ctx context.Context, armVal interface{}, base attr.Value, at attr.Type, bicep *generator.Type) (attr.Value, diag.Diagnostics) {
	if _, ok := at.(basetypes.ObjectType); ok {
		if m, ok := armVal.(map[string]interface{}); ok {
			if baseObj, ok := base.(types.Object); ok && !baseObj.IsNull() && !baseObj.IsUnknown() {
				var nested *generator.Type
				if bicep != nil && bicep.Kind == generator.KindObject {
					nested = bicep
				}
				return FlattenInto(ctx, m, baseObj, nested)
			}
		}
	}
	return flattenValue(ctx, armVal, at, bicep)
}
