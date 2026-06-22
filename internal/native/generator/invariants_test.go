package generator

import (
	"os"
	"testing"
)

// TestCheckFlagInvariantsClean proves the real storage-account schema satisfies
// the framework attribute-flag invariants after the full post-processing overlay.
func TestCheckFlagInvariantsClean(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/storage/microsoft.storage/2025-01-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	var sa *ResourceDefinition
	for _, d := range defs {
		if d.Name == "Microsoft.Storage/storageAccounts@2025-01-01" {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatal("storage account definition not found")
	}
	PostProcess([]*ResourceDefinition{sa})
	if vs := CheckFlagInvariants(sa); len(vs) != 0 {
		t.Errorf("storage account has %d flag-invariant violations:\n%s", len(vs), FormatViolations(vs))
	}
}

// TestCheckFlagInvariantsDetects exercises each violation class and confirms a
// valid default is not flagged.
func TestCheckFlagInvariantsDetects(t *testing.T) {
	i64 := func(n int64) *int64 { return &n }
	enum := &Type{Kind: KindUnion, Elements: []*Type{
		{Kind: KindStringLiteral, Name: "A"},
		{Kind: KindStringLiteral, Name: "B"},
	}}
	def := &ResourceDefinition{
		Name: "Microsoft.Fake/widgets@2024-01-01",
		Body: &Type{Kind: KindObject, Properties: map[string]*Property{
			// enum default outside the allowed set
			"mode": {Name: "mode", Type: enum, DefaultValue: "C"},
			// int default outside the validated range
			"days": {Name: "days", Type: &Type{Kind: KindInt}, DefaultValue: "99",
				Validators: []DescriptionValidator{{Kind: ValidatorIntRange, Min: i64(1), Max: i64(30)}}},
			// default on a Required property
			"sku": {Name: "sku", Type: &Type{Kind: KindString}, DefaultValue: "Standard", Flags: FlagRequired},
			// default on a read-only property
			"state": {Name: "state", Type: &Type{Kind: KindString}, DefaultValue: "Ready", Flags: FlagReadOnly},
			// valid default — must NOT be flagged
			"tier": {Name: "tier", Type: enum, DefaultValue: "A"},
		}},
	}
	byPath := map[string]bool{}
	for _, v := range CheckFlagInvariants(def) {
		byPath[v.Path] = true
	}
	for _, p := range []string{"mode", "days", "sku", "state"} {
		if !byPath[p] {
			t.Errorf("expected a violation on %q, got none", p)
		}
	}
	if byPath["tier"] {
		t.Error("valid default on \"tier\" should not be flagged")
	}
}

// TestDemoteDefaultedRequired confirms a Default demotes Required to Optional while
// leaving genuinely-required and read-only properties untouched.
func TestDemoteDefaultedRequired(t *testing.T) {
	body := &Type{Kind: KindObject, Properties: map[string]*Property{
		"kind":     {Name: "kind", Type: &Type{Kind: KindString}, DefaultValue: "StorageV2", Flags: FlagRequired},
		"location": {Name: "location", Type: &Type{Kind: KindString}, Flags: FlagRequired},
		"readonly": {Name: "readonly", Type: &Type{Kind: KindString}, DefaultValue: "x", Flags: FlagReadOnly},
	}}
	demoteDefaultedRequired(body)
	if body.Properties["kind"].Flags.IsRequired() {
		t.Error("a Required property with a Default should be demoted to Optional")
	}
	if !body.Properties["location"].Flags.IsRequired() {
		t.Error("a Required property without a Default must stay Required")
	}
	if !body.Properties["readonly"].Flags.IsReadOnly() {
		t.Error("a read-only property must be left untouched")
	}
}
