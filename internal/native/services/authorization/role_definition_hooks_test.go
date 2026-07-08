package authorization

import (
	"testing"

	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
)

// TestEnsureAssignableScopes exercises the ensureAssignableScopes BeforeCreate/BeforeUpdate
// hook directly (white-box) over the body shapes that define its contract: the role's own
// scope (parent_id) is injected as the sole assignable scope when the user omits or empties
// the list, a user-supplied list is left untouched, and absent properties is a no-op.
func TestEnsureAssignableScopes(t *testing.T) {
	const parentScope = "/subscriptions/00000000-0000-0000-0000-000000000000"

	t.Run("absent assignableScopes defaults to the parent scope", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			ID: parse.ResourceId{ParentId: parentScope},
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"roleName": "x",
				},
			},
		}

		ensureAssignableScopes(c)

		props, ok := c.Body["properties"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected Body[\"properties\"] to remain a map[string]interface{}, got %T", c.Body["properties"])
		}
		raw, present := props["assignableScopes"]
		if !present {
			t.Fatalf("expected assignableScopes to be injected, but the key is absent")
		}
		if raw == nil {
			t.Fatalf("expected injected assignableScopes to be a non-nil slice, got nil")
		}
		got, ok := raw.([]interface{})
		if !ok {
			t.Fatalf("expected injected assignableScopes to be []interface{}, got %T", raw)
		}
		if len(got) != 1 {
			t.Fatalf("expected injected assignableScopes to hold the single parent scope (len 1), got len %d: %#v", len(got), got)
		}
		if got[0] != parentScope {
			t.Fatalf("expected assignableScopes[0] to default to the parent scope %q, got %#v", parentScope, got[0])
		}
	})

	t.Run("empty assignableScopes defaults to the parent scope", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			ID: parse.ResourceId{ParentId: parentScope},
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"assignableScopes": []interface{}{},
				},
			},
		}

		ensureAssignableScopes(c)

		props := c.Body["properties"].(map[string]interface{})
		got, ok := props["assignableScopes"].([]interface{})
		if !ok {
			t.Fatalf("expected assignableScopes to be []interface{}, got %T", props["assignableScopes"])
		}
		if len(got) != 1 {
			t.Fatalf("expected an empty assignableScopes to default to the single parent scope (len 1), got len %d: %#v", len(got), got)
		}
		if got[0] != parentScope {
			t.Fatalf("expected assignableScopes[0] to default to the parent scope %q, got %#v", parentScope, got[0])
		}
	})

	t.Run("user-supplied assignableScopes is left untouched", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			ID: parse.ResourceId{ParentId: parentScope},
			Body: map[string]interface{}{
				"properties": map[string]interface{}{
					"assignableScopes": []interface{}{
						"/subscriptions/aaa",
						"/subscriptions/bbb",
					},
				},
			},
		}

		ensureAssignableScopes(c)

		props := c.Body["properties"].(map[string]interface{})
		got, ok := props["assignableScopes"].([]interface{})
		if !ok {
			t.Fatalf("expected assignableScopes to remain []interface{}, got %T", props["assignableScopes"])
		}
		if len(got) != 2 {
			t.Fatalf("expected user-supplied assignableScopes to be unchanged (len 2), got len %d: %#v", len(got), got)
		}
		if got[0] != "/subscriptions/aaa" || got[1] != "/subscriptions/bbb" {
			t.Fatalf("expected user-supplied assignableScopes to be preserved in order, got %#v", got)
		}
	})

	t.Run("missing properties is a no-op", func(t *testing.T) {
		c := &nativeresource.CrudCtx{
			ID:   parse.ResourceId{ParentId: parentScope},
			Body: map[string]interface{}{},
		}

		ensureAssignableScopes(c)

		if _, present := c.Body["properties"]; present {
			t.Fatalf("expected Body to still have no \"properties\" key, but it was added: %#v", c.Body)
		}
	})
}
