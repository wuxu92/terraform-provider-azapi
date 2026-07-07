package keyvault

import (
	"testing"

	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
)

// TestEnsureAccessPolicies exercises the ensureAccessPolicies BeforeCreate/BeforeUpdate
// hook directly (white-box) over the three body shapes that define its contract: an
// injected empty array when the user omits access policies, an untouched slice when the
// user supplies one, and a no-op when properties is absent entirely.
func TestEnsureAccessPolicies(t *testing.T) {
	t.Run("absent accessPolicies is injected as an empty non-nil slice", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"tenantId": "x",
					"sku":      map[string]interface{}{"name": "standard", "family": "A"},
				},
			},
		}

		ensureAccessPolicies(c)

		props, ok := c.Body["properties"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected Body[\"properties\"] to remain a map[string]interface{}, got %T", c.Body["properties"])
		}
		raw, present := props["accessPolicies"]
		if !present {
			t.Fatalf("expected accessPolicies to be injected, but the key is absent")
		}
		if raw == nil {
			t.Fatalf("expected injected accessPolicies to be a non-nil slice, got nil")
		}
		got, ok := raw.([]interface{})
		if !ok {
			t.Fatalf("expected injected accessPolicies to be []interface{}, got %T", raw)
		}
		if len(got) != 0 {
			t.Fatalf("expected injected accessPolicies to be empty, got len %d: %#v", len(got), got)
		}
	})

	t.Run("user-supplied accessPolicies is left untouched", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"accessPolicies": []interface{}{
						map[string]interface{}{"objectId": "abc"},
					},
				},
			},
		}

		ensureAccessPolicies(c)

		props := c.Body["properties"].(map[string]interface{})
		got, ok := props["accessPolicies"].([]interface{})
		if !ok {
			t.Fatalf("expected accessPolicies to remain []interface{}, got %T", props["accessPolicies"])
		}
		if len(got) != 1 {
			t.Fatalf("expected user-supplied accessPolicies to be unchanged (len 1), got len %d: %#v", len(got), got)
		}
		policy, ok := got[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected accessPolicies[0] to remain map[string]interface{}, got %T", got[0])
		}
		if policy["objectId"] != "abc" {
			t.Fatalf("expected accessPolicies[0].objectId to be preserved as \"abc\", got %#v", policy["objectId"])
		}
	})

	t.Run("missing properties is a no-op", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			Body: map[string]interface{}{},
		}

		ensureAccessPolicies(c)

		if _, present := c.Body["properties"]; present {
			t.Fatalf("expected Body to still have no \"properties\" key, but it was added: %#v", c.Body)
		}
	})
}
