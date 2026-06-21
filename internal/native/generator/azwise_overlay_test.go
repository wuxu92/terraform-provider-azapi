package generator

import (
	"os"
	"strings"
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
		got := navigate(body, c.path)
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

// TestApplyAzwiseStorageAccount verifies the azwise overlay flags real storage
// account properties as ForceNew and applies verified defaults.
func TestApplyAzwiseStorageAccount(t *testing.T) {
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
		t.Fatal("storage account not found")
	}

	// Apply azwise overlay directly (PostProcess also calls it).
	ApplyAzwise(sa)

	// is_hns_enabled (properties.isHnsEnabled) is a known ForceNew bool.
	if p := navigate(sa.Body, "properties.isHnsEnabled"); p == nil {
		t.Error("properties.isHnsEnabled not found")
	} else if !p.ForceNew {
		t.Error("expected properties.isHnsEnabled to be ForceNew via azwise")
	}

	// minimumTlsVersion has a verified azwise default (TLS1_2).
	if p := navigate(sa.Body, "properties.minimumTlsVersion"); p == nil {
		t.Error("properties.minimumTlsVersion not found")
	} else if p.DefaultValue == "" {
		t.Error("expected properties.minimumTlsVersion to have an azwise default")
	}

	// Emitted schema should contain RequiresReplace and the plan modifier import.
	src, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	if !strings.Contains(src, "RequiresReplace()") {
		t.Error("expected emitted schema to contain RequiresReplace()")
	}
	if !strings.Contains(src, "boolplanmodifier") {
		t.Error("expected emitted schema to import boolplanmodifier")
	}
}
