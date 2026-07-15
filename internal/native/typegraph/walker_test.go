package typegraph

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParseKeyVault(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/keyvault/microsoft.keyvault/2023-07-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	// Find vault
	var vault *ResourceDefinition
	for _, d := range defs {
		if d.APIVersion == "2023-07-01" && !containsSubResource(d.Name) {
			vault = d
			t.Logf("Found: %s", d.Name)
			break
		}
	}
	if vault == nil {
		// Try any vault definition
		for _, d := range defs {
			t.Logf("Available: %s", d.Name)
		}
		t.Skip("no vault definition found")
	}

	if vault.Body.Kind != KindObject {
		t.Fatalf("body kind = %d, want KindObject", vault.Body.Kind)
	}
	t.Logf("Vault properties: %d top-level", len(vault.Body.Properties))
}

func containsSubResource(name string) bool {
	// Count slashes in the resource type part (before @)
	parts := name
	if at := len(name) - 1; at >= 0 {
		for i, c := range name {
			if c == '@' {
				parts = name[:i]
				break
			}
		}
	}
	slashes := 0
	for _, c := range parts {
		if c == '/' {
			slashes++
		}
	}
	return slashes > 1
}

// buildDiscriminatedTypesJSON emits a minimal types.json whose resource body is a
// single property `variant` of a DiscriminatedObjectType with n variants.
func buildDiscriminatedTypesJSON(n int) []byte {
	var entries []string
	entries = append(entries, `{"$type":"StringType"}`) // idx 0
	var elements []string
	for i := range n {
		idx := len(entries) // variant object index
		entries = append(entries, fmt.Sprintf(`{"$type":"ObjectType","name":"V%d","properties":{"p%d":{"type":{"$ref":"#/0"},"flags":0},"kind":{"type":{"$ref":"#/0"},"flags":1}}}`, i, i))
		elements = append(elements, fmt.Sprintf(`"K%d":{"$ref":"#/%d"}`, i, idx))
	}
	discIdx := len(entries)
	entries = append(entries, fmt.Sprintf(`{"$type":"DiscriminatedObjectType","name":"Poly","discriminator":"kind","baseProperties":{"shared":{"type":{"$ref":"#/0"},"flags":0}},"elements":{%s}}`, strings.Join(elements, ",")))
	propsIdx := len(entries)
	entries = append(entries, fmt.Sprintf(`{"$type":"ObjectType","name":"props","properties":{"variant":{"type":{"$ref":"#/%d"},"flags":0}}}`, discIdx))
	entries = append(entries, fmt.Sprintf(`{"$type":"ResourceType","name":"Microsoft.Test/poly@2024-01-01","body":{"$ref":"#/%d"}}`, propsIdx))
	return []byte("[" + strings.Join(entries, ",") + "]")
}

// TestDiscriminatedVariantCountThreshold locks risk #3 from discriminator-report.md:
// a discriminated body at or under the variant cap is typed (KindDiscriminated);
// one past the cap degrades to a dynamic KindAny attribute so a pathological wide
// union (the ARM tail reaches 121 variants) never emits an unusable schema.
func TestDiscriminatedVariantCountThreshold(t *testing.T) {
	for _, tc := range []struct {
		name     string
		variants int
		wantKind TypeKind
	}{
		{"at cap types", maxDiscriminatedVariants, KindDiscriminated},
		{"over cap degrades", maxDiscriminatedVariants + 1, KindAny},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defs, err := ParseTypesJSON(buildDiscriminatedTypesJSON(tc.variants))
			if err != nil {
				t.Fatalf("ParseTypesJSON: %v", err)
			}
			if len(defs) != 1 {
				t.Fatalf("expected 1 def, got %d", len(defs))
			}
			got := defs[0].Body.Properties["variant"].Type.Kind
			if got != tc.wantKind {
				t.Errorf("variant kind for %d variants = %d, want %d", tc.variants, got, tc.wantKind)
			}
		})
	}
}
