package all

import (
	"context"
	"testing"

	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// TestRelationalPathsResolveInSchema guards the one thing the Go compiler cannot:
// Hooks.Relational declares cross-property constraints as plain dot-separated string
// paths (e.g. "properties.encryption.key_name"), which the compiler cannot relate to
// the generated schema. A typo would otherwise surface only when a practitioner
// validates a config for that specific resource. This walks every registered
// resource's relational paths against its schema and turns a latent typo into a fast
// CI failure.
//
// It lives here (not internal/native/resource) because the blank imports in this
// package populate services.Registry and the hook registry via each service init(),
// and placing it here avoids the service-packages -> resource import cycle.
func TestRelationalPathsResolveInSchema(t *testing.T) {
	if len(services.Registry) == 0 {
		t.Fatal("services.Registry is empty; expected generated descriptors to self-register via init()")
	}
	ctx := context.Background()
	for name := range services.Registry {
		for _, err := range nativeresource.ValidateRelationalPaths(ctx, name) {
			t.Error(err)
		}
	}
}
