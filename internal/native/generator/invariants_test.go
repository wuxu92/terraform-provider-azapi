package generator

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// TestCheckFlagInvariantsClean proves the real storage-account schema satisfies
// the framework attribute-flag invariants after the full post-processing overlay.
func TestCheckFlagInvariantsClean(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	tag := "Microsoft.Storage/storageAccounts@" + ver
	var sa *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatal("storage account definition not found")
	}
	typegraph.PostProcess([]*typegraph.ResourceDefinition{sa})
	if vs := CheckFlagInvariants(sa); len(vs) != 0 {
		t.Errorf("storage account has %d flag-invariant violations:\n%s", len(vs), FormatViolations(vs))
	}
}

// TestCheckFlagInvariantsDetects exercises each violation class and confirms a
// valid default is not flagged.
func TestCheckFlagInvariantsDetects(t *testing.T) {
	i64 := func(n int64) *int64 { return &n }
	enum := &typegraph.Type{Kind: typegraph.KindUnion, Elements: []*typegraph.Type{
		{Kind: typegraph.KindStringLiteral, Name: "A"},
		{Kind: typegraph.KindStringLiteral, Name: "B"},
	}}
	def := &typegraph.ResourceDefinition{
		Name: "Microsoft.Fake/widgets@2024-01-01",
		Body: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
			// enum default outside the allowed set
			"mode": {Name: "mode", Type: enum, DefaultValue: "C"},
			// int default outside the validated range
			"days": {Name: "days", Type: &typegraph.Type{Kind: typegraph.KindInt}, DefaultValue: "99",
				Validators: []typegraph.DescriptionValidator{{Kind: typegraph.ValidatorIntRange, Min: i64(1), Max: i64(30)}}},
			// default on a Required property
			"sku": {Name: "sku", Type: &typegraph.Type{Kind: typegraph.KindString}, DefaultValue: "Standard", Flags: typegraph.FlagRequired},
			// default on a read-only property
			"state": {Name: "state", Type: &typegraph.Type{Kind: typegraph.KindString}, DefaultValue: "Ready", Flags: typegraph.FlagReadOnly},
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
