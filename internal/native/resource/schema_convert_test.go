package resource

import (
	"context"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// syntheticResourceSchema builds a resource schema exercising every conversion
// path the focused tests care about: two identity string inputs (name +
// parent_id, one carrying a validator), a Computed leaf that also carries a
// validator (to prove validators are dropped on outputs), a ListAttribute with
// an ElementType, an "id" output, and a SingleNested body holding an
// Optional+Computed leaf and a ListNested grandchild. parentAttr is "parent_id".
func syntheticResourceSchema() rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"name": rschema.StringAttribute{
				Required:            true,
				Description:         "The resource name.",
				MarkdownDescription: "The resource **name**.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"parent_id": rschema.StringAttribute{
				Required: true,
			},
			// Computed body leaf. It deliberately carries a validator in the
			// source so the "outputs drop validators" contract has teeth.
			"location": rschema.StringAttribute{
				Computed:            true,
				Description:         "The Azure region.",
				MarkdownDescription: "The Azure `region`.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"zones": rschema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"properties": rschema.SingleNestedAttribute{
				Computed:            true,
				Description:         "The resource properties.",
				MarkdownDescription: "The resource **properties**.",
				Attributes: map[string]rschema.Attribute{
					// Optional+Computed in the source: must collapse to
					// Computed-only (never Optional) in the data source.
					"state": rschema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Provisioning state.",
					},
					"rules": rschema.ListNestedAttribute{
						Computed: true,
						NestedObject: rschema.NestedAttributeObject{
							Attributes: map[string]rschema.Attribute{
								"kind": rschema.StringAttribute{Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

// walkDSDescendants visits every nested descendant attribute (not the top level)
// of the given data-source attribute, threading a dotted path for diagnostics.
func walkDSDescendants(a dsschema.Attribute, prefix string, visit func(path string, a dsschema.Attribute)) {
	var children map[string]dsschema.Attribute
	switch v := a.(type) {
	case dsschema.SingleNestedAttribute:
		children = v.Attributes
	case dsschema.ListNestedAttribute:
		children = v.NestedObject.Attributes
	case dsschema.SetNestedAttribute:
		children = v.NestedObject.Attributes
	case dsschema.MapNestedAttribute:
		children = v.NestedObject.Attributes
	default:
		return
	}
	for name, child := range children {
		p := prefix + "." + name
		visit(p, child)
		walkDSDescendants(child, p, visit)
	}
}

// TestToDataSourceSchemaInputOutputFlags defends the core contract: exactly the
// two identity attributes (name + parentAttr) are Required inputs, while every
// other top-level attribute, "id", and ALL nested descendants are Computed
// outputs that are never Required or Optional.
func TestToDataSourceSchemaInputOutputFlags(t *testing.T) {
	ds := toDataSourceSchema(syntheticResourceSchema(), "parent_id")

	// Identity inputs.
	for _, name := range []string{"name", "parent_id"} {
		a := ds.Attributes[name]
		if a == nil {
			t.Fatalf("identity attribute %q missing after conversion", name)
		}
		if !a.IsRequired() {
			t.Errorf("identity %q: IsRequired = false, want true", name)
		}
		if a.IsComputed() || a.IsOptional() {
			t.Errorf("identity %q: IsComputed=%v IsOptional=%v, want both false", name, a.IsComputed(), a.IsOptional())
		}
	}

	// Top-level Computed outputs (including id and the nested container itself).
	for _, name := range []string{"location", "zones", "id", "properties"} {
		a := ds.Attributes[name]
		if a == nil {
			t.Fatalf("output attribute %q missing after conversion", name)
		}
		if !a.IsComputed() {
			t.Errorf("output %q: IsComputed = false, want true", name)
		}
		if a.IsRequired() || a.IsOptional() {
			t.Errorf("output %q: IsRequired=%v IsOptional=%v, want both false", name, a.IsRequired(), a.IsOptional())
		}
	}

	// Every nested descendant is a Computed output — never Required, never
	// Optional (the source made "state" Optional+Computed; conversion must strip
	// Optional).
	descendants := 0
	walkDSDescendants(ds.Attributes["properties"], "properties", func(path string, a dsschema.Attribute) {
		descendants++
		if !a.IsComputed() {
			t.Errorf("descendant %q: IsComputed = false, want true", path)
		}
		if a.IsRequired() || a.IsOptional() {
			t.Errorf("descendant %q: IsRequired=%v IsOptional=%v, want both false", path, a.IsRequired(), a.IsOptional())
		}
	})
	if descendants != 3 { // state, rules, rules.kind
		t.Fatalf("walked %d nested descendants, want 3 (state, rules, rules.kind)", descendants)
	}
}

// TestToDataSourceSchemaValidatorHandling proves string validators are carried
// over onto identity inputs and dropped from Computed outputs.
func TestToDataSourceSchemaValidatorHandling(t *testing.T) {
	ds := toDataSourceSchema(syntheticResourceSchema(), "parent_id")

	name, ok := ds.Attributes["name"].(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want dsschema.StringAttribute", ds.Attributes["name"])
	}
	if len(name.Validators) == 0 {
		t.Errorf("identity name: validators dropped (len 0), want preserved (>0)")
	}

	loc, ok := ds.Attributes["location"].(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("location attribute is %T, want dsschema.StringAttribute", ds.Attributes["location"])
	}
	if len(loc.Validators) != 0 {
		t.Errorf("output location: %d validators carried over, want 0 (dropped)", len(loc.Validators))
	}
}

// TestToDataSourceSchemaTypeFidelity proves GetType() is preserved through
// conversion for a leaf, a ListAttribute (ElementType), and a nested object
// (recursive object type), compared via attr.Type.Equal.
func TestToDataSourceSchemaTypeFidelity(t *testing.T) {
	rs := syntheticResourceSchema()
	ds := toDataSourceSchema(rs, "parent_id")

	for _, name := range []string{"name", "zones", "properties"} {
		want := rs.Attributes[name].GetType()
		got := ds.Attributes[name].GetType()
		if !got.Equal(want) {
			t.Errorf("attr %q: GetType() = %s, want %s", name, got, want)
		}
	}
}

// TestToDataSourceSchemaDescriptionsPreserved proves Description and
// MarkdownDescription survive conversion for a top-level leaf and a nested
// container.
func TestToDataSourceSchemaDescriptionsPreserved(t *testing.T) {
	ds := toDataSourceSchema(syntheticResourceSchema(), "parent_id")

	cases := []struct {
		name       string
		wantDesc   string
		wantMdDesc string
	}{
		{"location", "The Azure region.", "The Azure `region`."},
		{"properties", "The resource properties.", "The resource **properties**."},
	}
	for _, tc := range cases {
		a := ds.Attributes[tc.name]
		if got := a.GetDescription(); got != tc.wantDesc {
			t.Errorf("attr %q: Description = %q, want %q", tc.name, got, tc.wantDesc)
		}
		if got := a.GetMarkdownDescription(); got != tc.wantMdDesc {
			t.Errorf("attr %q: MarkdownDescription = %q, want %q", tc.name, got, tc.wantMdDesc)
		}
	}
}

// TestToDataSourceSchemaParentAttrParameterization proves only "name" and the
// literally-passed parentAttr are treated as identity. A body attribute named
// "parent_id" (a suggestive name that is NOT the passed parentAttr) stays a
// Computed output.
func TestToDataSourceSchemaParentAttrParameterization(t *testing.T) {
	rs := rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"name":              rschema.StringAttribute{Required: true},
			"resource_group_id": rschema.StringAttribute{Required: true},
			// Named like a parent reference but not the passed parentAttr: must
			// remain a Computed output, proving the identity check keys off the
			// argument, not a hardcoded "parent_id".
			"parent_id": rschema.StringAttribute{Computed: true},
			"location":  rschema.StringAttribute{Computed: true},
		},
	}
	ds := toDataSourceSchema(rs, "resource_group_id")

	for _, name := range []string{"name", "resource_group_id"} {
		if !ds.Attributes[name].IsRequired() {
			t.Errorf("identity %q: IsRequired = false, want true", name)
		}
	}
	for _, name := range []string{"parent_id", "location"} {
		a := ds.Attributes[name]
		if !a.IsComputed() {
			t.Errorf("output %q: IsComputed = false, want true", name)
		}
		if a.IsRequired() {
			t.Errorf("output %q: IsRequired = true, want false (only name + passed parentAttr are identity)", name)
		}
	}
}

// TestToDataSourceSchemaRegistrySweep converts every real registered descriptor
// and asserts the framework accepts the emitted schema (proving every attribute
// kind the generators emit is handled — the panic/coverage guard), that the
// attribute count is preserved, and that exactly name+ParentAttr are Required
// while all other top-level attributes are Computed.
func TestToDataSourceSchemaRegistrySweep(t *testing.T) {
	ctx := context.Background()
	if len(services.Registry) == 0 {
		t.Fatal("services.Registry is empty; the test binary must import internal/native/services/all (see base_test.go)")
	}
	for name, desc := range services.Registry {
		name, desc := name, desc
		t.Run(name, func(t *testing.T) {
			ds := NewDataSource(name).(*DataSource)
			var resp datasource.SchemaResponse
			ds.Schema(ctx, datasource.SchemaRequest{}, &resp)

			if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
				t.Fatalf("ValidateImplementation reported errors: %v", diags)
			}

			rs := desc.Schema()
			if got, want := len(resp.Schema.Attributes), len(rs.Attributes); got != want {
				t.Fatalf("data source has %d attributes, want %d (resource schema)", got, want)
			}

			required := 0
			for an, a := range resp.Schema.Attributes {
				identity := an == "name" || an == desc.ParentAttr
				if identity {
					required++
					if !a.IsRequired() {
						t.Errorf("identity attr %q: IsRequired = false, want true", an)
					}
					if a.IsComputed() || a.IsOptional() {
						t.Errorf("identity attr %q: IsComputed=%v IsOptional=%v, want both false", an, a.IsComputed(), a.IsOptional())
					}
				} else {
					if !a.IsComputed() {
						t.Errorf("output attr %q: IsComputed = false, want true", an)
					}
					if a.IsRequired() || a.IsOptional() {
						t.Errorf("output attr %q: IsRequired=%v IsOptional=%v, want both false", an, a.IsRequired(), a.IsOptional())
					}
				}
			}
			// name + ParentAttr must both be present and Required — nothing else.
			if required != 2 {
				t.Errorf("Required attribute count = %d, want 2 (name + %q)", required, desc.ParentAttr)
			}
		})
	}
}
