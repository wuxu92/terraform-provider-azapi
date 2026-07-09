package resource

import (
	"fmt"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// toDataSourceSchema converts a generated resource schema into the equivalent
// data-source schema. The framework keeps resource and data-source schemas in
// distinct packages (whose Attribute interface lives in an un-importable
// internal package), so the conversion type-switches on the public concrete
// resource/schema attribute types and rebuilds datasource/schema values.
//
// Semantics: a data source reads an existing resource, so the two identity
// attributes — "name" and the parent reference (parentAttr) — become Required
// inputs the practitioner supplies to locate the resource, while every other
// attribute (the whole ARM body plus "id") becomes a Computed output. Plan
// modifiers and defaults are dropped (they are meaningless for read-only data);
// element types and descriptions are preserved, and string validators on the
// identity inputs are carried over for early, clear errors.
func toDataSourceSchema(rs rschema.Schema, parentAttr string) dschema.Schema {
	attrs := make(map[string]dschema.Attribute, len(rs.Attributes))
	for name, a := range rs.Attributes {
		identity := name == "name" || name == parentAttr
		attrs[name] = toDSAttribute(a, identity)
	}
	return dschema.Schema{
		Description:         rs.Description,
		MarkdownDescription: rs.MarkdownDescription,
		DeprecationMessage:  rs.DeprecationMessage,
		Attributes:          attrs,
	}
}

// toDSAttributes converts a nested attribute map; every descendant is a Computed
// output (only top-level name/parent are practitioner inputs).
func toDSAttributes(in map[string]rschema.Attribute) map[string]dschema.Attribute {
	out := make(map[string]dschema.Attribute, len(in))
	for name, a := range in {
		out[name] = toDSAttribute(a, false)
	}
	return out
}

// toDSAttribute converts one attribute. identity=true marks a practitioner input
// (Required, validators preserved); otherwise the attribute is a Computed output.
func toDSAttribute(a rschema.Attribute, identity bool) dschema.Attribute {
	switch v := a.(type) {
	case rschema.StringAttribute:
		out := dschema.StringAttribute{
			CustomType:          v.CustomType,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
		if identity {
			out.Required = true
			out.Validators = v.Validators
		} else {
			out.Computed = true
		}
		return out
	case rschema.BoolAttribute:
		return dschema.BoolAttribute{
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.Int64Attribute:
		return dschema.Int64Attribute{
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.Float64Attribute:
		return dschema.Float64Attribute{
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.NumberAttribute:
		return dschema.NumberAttribute{
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.DynamicAttribute:
		return dschema.DynamicAttribute{
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.ListAttribute:
		return dschema.ListAttribute{
			ElementType:         v.ElementType,
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.MapAttribute:
		return dschema.MapAttribute{
			ElementType:         v.ElementType,
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.SetAttribute:
		return dschema.SetAttribute{
			ElementType:         v.ElementType,
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.ObjectAttribute:
		return dschema.ObjectAttribute{
			AttributeTypes:      v.AttributeTypes,
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.SingleNestedAttribute:
		return dschema.SingleNestedAttribute{
			Attributes:          toDSAttributes(v.Attributes),
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.ListNestedAttribute:
		return dschema.ListNestedAttribute{
			NestedObject:        dschema.NestedAttributeObject{Attributes: toDSAttributes(v.NestedObject.Attributes)},
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.SetNestedAttribute:
		return dschema.SetNestedAttribute{
			NestedObject:        dschema.NestedAttributeObject{Attributes: toDSAttributes(v.NestedObject.Attributes)},
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	case rschema.MapNestedAttribute:
		return dschema.MapNestedAttribute{
			NestedObject:        dschema.NestedAttributeObject{Attributes: toDSAttributes(v.NestedObject.Attributes)},
			CustomType:          v.CustomType,
			Computed:            true,
			Sensitive:           v.Sensitive,
			Description:         v.Description,
			MarkdownDescription: v.MarkdownDescription,
			DeprecationMessage:  v.DeprecationMessage,
		}
	default:
		// Generated schemas only use the kinds handled above (no Blocks, no
		// Int32/Float32). A new kind should surface loudly the moment it appears
		// rather than silently drop an attribute from the data source.
		panic(fmt.Sprintf("native datasource: unsupported schema attribute type %T", a))
	}
}
