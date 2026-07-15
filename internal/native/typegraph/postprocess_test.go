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

// TestExtractDefaultsWalksDiscriminated confirms description-mined defaults are
// extracted from a discriminated block's base properties AND its variant
// properties — before the fix these walkers stopped at the KindObject guard and
// silently skipped the whole discriminated subtree.
func TestExtractDefaultsWalksDiscriminated(t *testing.T) {
	disc := &Type{Kind: KindDiscriminated, Name: "Auth", Discriminator: "type",
		Properties: map[string]*Property{
			"tls": {Name: "tls", Type: &Type{Kind: KindBool}, Description: "Enabled by default."},
		},
		Variants: map[string]*Type{
			"Basic": {Kind: KindObject, Name: "Basic", Properties: map[string]*Property{
				"retries": {Name: "retries", Type: &Type{Kind: KindInt}, Description: "The default value is 3."},
			}},
		},
	}
	body := &Type{Kind: KindObject, Properties: map[string]*Property{
		"auth": {Name: "auth", Type: disc, Description: "Auth configuration."},
	}}
	extractDefaults(body)

	if got := disc.Properties["tls"].DefaultValue; got != "true" {
		t.Errorf("base property default: got %q, want true", got)
	}
	if got := disc.Variants["Basic"].Properties["retries"].DefaultValue; got != "3" {
		t.Errorf("variant property default: got %q, want 3", got)
	}
}

// TestExtractDescriptionValidatorsWalksDiscriminated confirms the validator walker
// descends a discriminated block's base and variant properties.
func TestExtractDescriptionValidatorsWalksDiscriminated(t *testing.T) {
	disc := &Type{Kind: KindDiscriminated, Name: "Auth", Discriminator: "type",
		Properties: map[string]*Property{},
		Variants: map[string]*Type{
			"Basic": {Kind: KindObject, Name: "Basic", Properties: map[string]*Property{
				"count": {Name: "count", Type: &Type{Kind: KindInt}, Description: "Value must be between 1 and 5."},
			}},
		},
	}
	body := &Type{Kind: KindObject, Properties: map[string]*Property{
		"auth": {Name: "auth", Type: disc, Description: "Auth configuration."},
	}}
	extractDescriptionValidators(body)

	if got := len(disc.Variants["Basic"].Properties["count"].Validators); got != 1 {
		t.Fatalf("variant validator count: got %d, want 1", got)
	}
}

// TestPromoteSingleOptionalWalksVariant confirms the sole-optional promotion runs
// inside a variant block (a KindObject) but does NOT fire against the
// discriminated node's base properties, whose variant siblings it cannot count.
func TestPromoteSingleOptionalWalksVariant(t *testing.T) {
	disc := &Type{Kind: KindDiscriminated, Name: "Auth", Discriminator: "type",
		Properties: map[string]*Property{
			"shared": {Name: "shared", Type: &Type{Kind: KindString}},
		},
		Variants: map[string]*Type{
			"Basic": {Kind: KindObject, Name: "Basic", Properties: map[string]*Property{
				"username": {Name: "username", Type: &Type{Kind: KindString}},
			}},
		},
	}
	body := &Type{Kind: KindObject, Properties: map[string]*Property{
		"auth":    {Name: "auth", Type: disc},
		"sibling": {Name: "sibling", Type: &Type{Kind: KindString}},
	}}
	promoteSingleOptional(body)

	if !disc.Variants["Basic"].Properties["username"].Flags.IsRequired() {
		t.Error("sole optional inside a variant block should be promoted to Required")
	}
	if disc.Properties["shared"].Flags.IsRequired() {
		t.Error("a discriminated block's lone base property must not be promoted (variants are the real content)")
	}
}
