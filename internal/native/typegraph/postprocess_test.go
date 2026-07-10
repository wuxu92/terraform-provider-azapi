package typegraph

import "testing"

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

func TestMarkTopLevelLocationForceNew(t *testing.T) {
	writable := &ResourceDefinition{Body: &Type{Kind: KindObject, Properties: map[string]*Property{
		"location": {Name: "location", Type: &Type{Kind: KindString}},
	}}}
	markTopLevelLocationForceNew(writable)
	if !writable.Body.Properties["location"].ForceNew {
		t.Fatal("writable top-level location should require replacement")
	}

	computed := &ResourceDefinition{Body: &Type{Kind: KindObject, Properties: map[string]*Property{
		"location": {Name: "location", Type: &Type{Kind: KindString}, Flags: FlagReadOnly},
	}}}
	markTopLevelLocationForceNew(computed)
	if computed.Body.Properties["location"].ForceNew {
		t.Fatal("computed/read-only location should not get a replacement modifier")
	}
}
