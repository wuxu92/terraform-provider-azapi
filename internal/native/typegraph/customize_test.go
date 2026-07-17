package typegraph

import "testing"

// fixtureDef builds a resource definition with a nested body and an array of
// objects for exercising the customizer helper methods.
func fixtureDef() *ResourceDefinition {
	elem := &Type{Kind: KindObject, Properties: map[string]*Property{
		"count": {Name: "count", Type: &Type{Kind: KindString}},
	}}
	body := &Type{Kind: KindObject, Properties: map[string]*Property{
		"sku": {Name: "sku", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
			"name": {Name: "name", Type: &Type{Kind: KindString}},
		}}},
		"properties": {Name: "properties", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
			"tlsVersion": {Name: "tlsVersion", Type: &Type{Kind: KindString}},
			"secret":     {Name: "secret", Type: &Type{Kind: KindString}},
			"tags":       {Name: "tags", Type: &Type{Kind: KindArray, ElementType: &Type{Kind: KindString}}},
			"rules":      {Name: "rules", Type: &Type{Kind: KindArray, ElementType: elem}},
			"probe":      {Name: "probe", Type: &Type{Kind: KindString}},
		}}},
	}}
	return &ResourceDefinition{Name: "Microsoft.Fake/things@2024-01-01", Body: body}
}

func TestRequiredComputedForceNewSensitive(t *testing.T) {
	def := fixtureDef()
	def.Required("sku", "sku.name").
		Computed("properties.probe").
		ForceNew("properties.tlsVersion").
		Sensitive("properties.secret")

	if !FindProperty(def, "sku").Flags.IsRequired() || !FindProperty(def, "sku.name").Flags.IsRequired() {
		t.Error("Required should set FlagRequired on every path")
	}
	if !FindProperty(def, "properties.probe").ForceComputed {
		t.Error("Computed should set ForceComputed")
	}
	if !FindProperty(def, "properties.tlsVersion").ForceNew {
		t.Error("ForceNew should set ForceNew")
	}
	if !FindProperty(def, "properties.secret").Sensitive {
		t.Error("Sensitive should set Sensitive")
	}
}

func TestArrayAndDefaultHelpers(t *testing.T) {
	def := fixtureDef()
	def.AsSet("properties.tags").
		WithEmptyListDefault("properties.rules").
		NonNullStateForUnknown("properties.probe").
		Default("properties.tlsVersion", "TLS1_2")

	if !FindProperty(def, "properties.tags").UseSet {
		t.Error("AsSet should set UseSet")
	}
	if !FindProperty(def, "properties.rules").DefaultEmptyList {
		t.Error("WithEmptyListDefault should set DefaultEmptyList")
	}
	if !FindProperty(def, "properties.probe").NonNullStateForUnknown {
		t.Error("NonNullStateForUnknown should set NonNullStateForUnknown")
	}
	if got := FindProperty(def, "properties.tlsVersion").DefaultValue; got != "TLS1_2" {
		t.Errorf("Default = %q, want TLS1_2", got)
	}
}

func TestAddValidatorsForAppends(t *testing.T) {
	def := fixtureDef()
	FindProperty(def, "properties.tlsVersion").Validators = []DescriptionValidator{LengthValidator(1, 4)}

	def.AddValidatorsFor("properties.tlsVersion", RegexValidator("^T", "must start with T"), Validator(func() {}))

	got := FindProperty(def, "properties.tlsVersion").Validators
	if len(got) != 3 {
		t.Fatalf("validators = %d, want 3 (append, not replace)", len(got))
	}
	if got[0].Kind != ValidatorStringLength {
		t.Error("AddValidatorsFor must append after the existing validator, not replace it")
	}

	// Empty variadic is a no-op that neither panics nor mutates.
	def.AddValidatorsFor("properties.tlsVersion")
	if len(FindProperty(def, "properties.tlsVersion").Validators) != 3 {
		t.Error("AddValidatorsFor with no validators should be a no-op")
	}
}

func TestEnvelopeHelpers(t *testing.T) {
	def := fixtureDef()
	def.SetNameValidators(LengthValidator(3, 24), RegexValidator("^[a-z]+$", "lowercase")).
		SetParent("scope_id", "the scope").
		AddMetaAttr(MetaAttr{Name: "purge_on_destroy", Description: "purge"}).
		AddMetaAttr(MetaAttr{Name: "second", Description: "another"})

	if len(def.Envelope.Name.Validators) != 2 {
		t.Errorf("name validators = %d, want 2", len(def.Envelope.Name.Validators))
	}
	// SetNameValidators replaces rather than appends.
	def.SetNameValidators(OneOfValidator("one", "default"))
	if len(def.Envelope.Name.Validators) != 1 {
		t.Errorf("SetNameValidators should replace: got %d, want 1", len(def.Envelope.Name.Validators))
	}
	if def.Envelope.Parent.Name != "scope_id" || def.Envelope.Parent.Description != "the scope" {
		t.Errorf("SetParent = %q/%q", def.Envelope.Parent.Name, def.Envelope.Parent.Description)
	}
	if len(def.Envelope.Meta) != 2 || def.Envelope.Meta[0].Name != "purge_on_destroy" || def.Envelope.Meta[1].Name != "second" {
		t.Errorf("AddMetaAttr should append in order: %#v", def.Envelope.Meta)
	}
}

func TestHelperPanicsOnUnknownPath(t *testing.T) {
	def := fixtureDef()
	defer func() {
		if recover() == nil {
			t.Error("Required on an unknown path should panic (a customizer typo must fail generation)")
		}
	}()
	def.Required("properties.doesNotExist")
}
