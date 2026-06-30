package generator

import (
	"testing"
)

func TestValidateStorageAccountSchema(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	tag := "Microsoft.Storage/storageAccounts@" + ver

	PostProcess(defs)

	var sa *ResourceDefinition
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
	mismatches := ValidateEmittedSchema(source, sa.Body, EnvelopeAttrNames(sa)...)

	if len(mismatches) > 0 {
		t.Logf("Mismatches:\n%s", FormatMismatches(mismatches))

		errors := 0
		for _, m := range mismatches {
			if m.Kind != MismatchMissingInSchema {
				errors++
			}
		}
		if errors > 0 {
			t.Errorf("%d schema errors (extra or type mismatch)", errors)
		}
	} else {
		// Count validated paths
		paths := make(map[string]*Property)
		CollectExpectedPaths(sa.Body, "", paths)
		t.Logf("Schema-bicep validation passed: %d properties validated, 0 mismatches", len(paths))
	}
}

func TestValidateWebServerFarmSchema(t *testing.T) {
	defs, ver := latestWebServerFarmDefs(t)
	tag := "Microsoft.Web/serverfarms@" + ver

	PostProcess(defs)

	var farm *ResourceDefinition
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

	mismatches := ValidateEmittedSchema(source, farm.Body, EnvelopeAttrNames(farm)...)
	for _, m := range mismatches {
		if m.Kind != MismatchMissingInSchema {
			t.Fatalf("unexpected schema mismatch for web server farm: %s", FormatMismatches(mismatches))
		}
	}
}

func TestValidateWebSiteSchema(t *testing.T) {
	defs, ver := latestWebSiteDefs(t)
	tag := "Microsoft.Web/sites@" + ver

	PostProcess(defs)

	var site *ResourceDefinition
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

	mismatches := ValidateEmittedSchema(source, site.Body, EnvelopeAttrNames(site)...)
	for _, m := range mismatches {
		if m.Kind != MismatchMissingInSchema {
			t.Fatalf("unexpected schema mismatch for web site: %s", FormatMismatches(mismatches))
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

func TestValidateDetectsMismatches(t *testing.T) {
	// Build a bicep type with known properties
	bicepBody := &Type{
		Kind: KindObject,
		Properties: map[string]*Property{
			"name":     {Name: "name", Type: &Type{Kind: KindString}, Flags: FlagSystemManaged},
			"location": {Name: "location", Type: &Type{Kind: KindString}, Flags: FlagRequired},
			"sku": {Name: "sku", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"name": {Name: "name", Type: &Type{Kind: KindString}, Flags: FlagRequired},
				"tier": {Name: "tier", Type: &Type{Kind: KindString}},
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
		case MismatchExtraInSchema:
			extraCount++
			if m.Path != "phantom" {
				t.Errorf("unexpected extra: %s", m.Path)
			}
		case MismatchMissingInSchema:
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

	t.Logf("Mismatches:\n%s", FormatMismatches(mismatches))
}
