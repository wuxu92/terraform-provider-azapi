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
