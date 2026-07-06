package generator

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

func TestValidateStorageAccountSchema(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	tag := "Microsoft.Storage/storageAccounts@" + ver
	typegraph.PostProcess(defs)

	var sa *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatal("storage account not found")
	}

	// Generate the schema source
	source, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	// Validate: emitted source covers all bicep properties and vice versa. The
	// synthesized envelope attributes (name / parent reference / id) are excluded
	// since they are not part of the bicep body graph.
	mismatches := ValidateEmittedSchema(source, sa.Body, typegraph.EnvelopeAttrNames(sa)...)

	if len(mismatches) > 0 {
		t.Logf("Mismatches:\n%s", typegraph.FormatMismatches(mismatches))

		errors := 0
		for _, m := range mismatches {
			if m.Kind != typegraph.MismatchMissingInSchema {
				errors++
			}
		}
		if errors > 0 {
			t.Errorf("%d schema errors (extra or type mismatch)", errors)
		}
	} else {
		// Count validated paths
		paths := make(map[string]*typegraph.Property)
		CollectExpectedPaths(sa.Body, "", paths)
		t.Logf("Schema-bicep validation passed: %d properties validated, 0 mismatches", len(paths))
	}
}

func TestValidateWebServerFarmSchema(t *testing.T) {
	defs, ver := latestWebServerFarmDefs(t)
	tag := "Microsoft.Web/serverfarms@" + ver
	typegraph.PostProcess(defs)

	var farm *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			farm = d
			break
		}
	}
	if farm == nil {
		t.Fatal("web server farm not found")
	}

	source, err := EmitSchema(farm)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	mismatches := ValidateEmittedSchema(source, farm.Body, typegraph.EnvelopeAttrNames(farm)...)
	for _, m := range mismatches {
		if m.Kind != typegraph.MismatchMissingInSchema {
			t.Fatalf("unexpected schema mismatch for web server farm: %s", typegraph.FormatMismatches(mismatches))
		}
	}
}

func TestValidateWebSiteSchema(t *testing.T) {
	defs, ver := latestWebSiteDefs(t)
	tag := "Microsoft.Web/sites@" + ver
	typegraph.PostProcess(defs)

	var site *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			site = d
			break
		}
	}
	if site == nil {
		t.Fatal("web site not found")
	}

	source, err := EmitSchema(site)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	mismatches := ValidateEmittedSchema(source, site.Body, typegraph.EnvelopeAttrNames(site)...)
	for _, m := range mismatches {
		if m.Kind != typegraph.MismatchMissingInSchema {
			t.Fatalf("unexpected schema mismatch for web site: %s", typegraph.FormatMismatches(mismatches))
		}
	}
}

func TestExtractEmittedPathsKeepsNestedSiblingsSeparate(t *testing.T) {
	source := `
		"identity": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
		"properties": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"server_farm_id": schema.StringAttribute{
					Required: true,
				},
			},
		},
	`

	paths := extractEmittedPaths(source)
	for _, p := range []string{"identity", "identity.type", "properties", "properties.server_farm_id"} {
		if !paths[p] {
			t.Fatalf("missing extracted path %q in %#v", p, paths)
		}
	}
	if paths["identity.properties"] || paths["identity.properties.server_farm_id"] {
		t.Fatalf("sibling nested object was incorrectly attributed under identity: %#v", paths)
	}
}

func TestExtractEmittedPathsExpandsManagedIdentityHelper(t *testing.T) {
	source := `
		"identity": nativeschema.ManagedServiceIdentity(false, ""),
	`

	paths := extractEmittedPaths(source)
	for _, p := range []string{
		"identity",
		"identity.principal_id",
		"identity.tenant_id",
		"identity.type",
		"identity.user_assigned_identities",
		"identity.user_assigned_identities.client_id",
		"identity.user_assigned_identities.principal_id",
	} {
		if !paths[p] {
			t.Fatalf("missing extracted path %q in %#v", p, paths)
		}
	}
}

func TestValidateDetectsMismatches(t *testing.T) {
	// Build a bicep type with known properties
	bicepBody := &typegraph.Type{
		Kind: typegraph.KindObject,
		Properties: map[string]*typegraph.Property{
			"name":     {Name: "name", Type: &typegraph.Type{Kind: typegraph.KindString}, Flags: typegraph.FlagSystemManaged},
			"location": {Name: "location", Type: &typegraph.Type{Kind: typegraph.KindString}, Flags: typegraph.FlagRequired},
			"sku": {Name: "sku", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"name": {Name: "name", Type: &typegraph.Type{Kind: typegraph.KindString}, Flags: typegraph.FlagRequired},
				"tier": {Name: "tier", Type: &typegraph.Type{Kind: typegraph.KindString}},
			}}},
		},
	}

	// Generate a schema source that has an extra attribute and misses "tier"
	fakeSource := `
		"location": schema.StringAttribute{
			Required: true,
		},
		"sku": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required: true,
				},
			},
		},
		"phantom": schema.StringAttribute{
			Optional: true,
		},
	`

	mismatches := ValidateEmittedSchema(fakeSource, bicepBody)

	var extraCount, missingCount int
	for _, m := range mismatches {
		switch m.Kind {
		case typegraph.MismatchExtraInSchema:
			extraCount++
			if m.Path != "phantom" {
				t.Errorf("unexpected extra: %s", m.Path)
			}
		case typegraph.MismatchMissingInSchema:
			missingCount++
			if m.Path != "sku.tier" {
				t.Errorf("unexpected missing: %s", m.Path)
			}
		}
	}

	if extraCount != 1 {
		t.Errorf("expected 1 extra in schema, got %d", extraCount)
	}
	if missingCount != 1 {
		t.Errorf("expected 1 missing in schema, got %d", missingCount)
	}

	t.Logf("Mismatches:\n%s", typegraph.FormatMismatches(mismatches))
}
