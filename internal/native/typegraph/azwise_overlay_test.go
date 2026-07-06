package typegraph

import (
	"testing"
)

func TestNavigate(t *testing.T) {
	body := &Type{
		Kind: KindObject,
		Properties: map[string]*Property{
			"sku": {Name: "sku", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"name": {Name: "name", Type: &Type{Kind: KindString}},
			}}},
			"properties": {Name: "properties", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"networkAcls": {Name: "networkAcls", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
					"defaultAction": {Name: "defaultAction", Type: &Type{Kind: KindString}},
				}}},
			}}},
		},
	}

	cases := []struct {
		path string
		want string // expected resolved property ARM name, "" = nil
	}{
		{"sku.name", "name"},
		{"properties.networkAcls.defaultAction", "defaultAction"},
		{"sku", "sku"},
		{"sku.missing", ""},
		{"properties.networkAcls.missing", ""},
		{"missing.path", ""},
	}
	for _, c := range cases {
		got := Navigate(body, c.path)
		if c.want == "" {
			if got != nil {
				t.Errorf("navigate(%q) = %q, want nil", c.path, got.Name)
			}
			continue
		}
		if got == nil || got.Name != c.want {
			t.Errorf("navigate(%q) = %v, want %q", c.path, got, c.want)
		}
	}
}

func TestFormatAzwiseDefault(t *testing.T) {
	cases := []struct {
		v    interface{}
		typ  *Type
		want string
		ok   bool
	}{
		{true, &Type{Kind: KindBool}, "true", true},
		{false, &Type{Kind: KindBool}, "false", true},
		{float64(5), &Type{Kind: KindInt}, "5", true},
		{"TLS1_2", &Type{Kind: KindString}, "TLS1_2", true},
		{"Hot", &Type{Kind: KindUnion, Elements: []*Type{{Kind: KindStringLiteral, Name: "Hot"}}}, "Hot", true},
		{"x", &Type{Kind: KindBool}, "", false}, // wrong type
	}
	for _, c := range cases {
		got, ok := formatAzwiseDefault(c.v, c.typ)
		if ok != c.ok || got != c.want {
			t.Errorf("formatAzwiseDefault(%v, %d) = (%q, %v), want (%q, %v)", c.v, c.typ.Kind, got, ok, c.want, c.ok)
		}
	}
}

// TestDropValidators verifies the helper removes only same-kind validators so an
// authoritative azwise rule can replace a heuristic one without stacking.
func TestDropValidators(t *testing.T) {
	prop := &Property{Validators: []DescriptionValidator{
		{Kind: ValidatorIntRange},
		{Kind: ValidatorRegex},
		{Kind: ValidatorIntRange},
	}}
	dropValidators(prop, ValidatorIntRange)
	if len(prop.Validators) != 1 || prop.Validators[0].Kind != ValidatorRegex {
		t.Fatalf("dropValidators left %d validators, want 1 Regex: %+v", len(prop.Validators), prop.Validators)
	}
}
