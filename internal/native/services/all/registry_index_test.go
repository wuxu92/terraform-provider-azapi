package all

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// TestRegistryDescriptorsResolveInEmbeddedIndex enforces the load-bearing
// invariant behind ADR-0002 (no static property map): every registered
// descriptor's pinned ARMType@APIVersion MUST exist in the embedded schema index,
// because the runtime Base resolves ARM camelCase property names by loading that
// exact types.json at CRUD time. Without this guard, a stale pin left behind by a
// bicep-types refresh compiles and registers cleanly, then fails only at
// `terraform apply` with "no types.json location for X@Y". This test turns that
// runtime apply failure into a fast-lane CI failure.
//
// Blank imports in this package populate services.Registry via each generated
// service's init(); placing the test here (rather than in internal/native/resource)
// avoids the import cycle service packages -> resource.
func TestRegistryDescriptorsResolveInEmbeddedIndex(t *testing.T) {
	if len(services.Registry) == 0 {
		t.Fatal("services.Registry is empty; expected generated descriptors to self-register via init()")
	}
	for name, d := range services.Registry {
		if _, err := azure.GetResourceTypeLocation(d.ARMType, d.APIVersion); err != nil {
			t.Errorf("resource %q pins %s@%s but it is absent from the embedded schema index: %v",
				name, d.ARMType, d.APIVersion, err)
		}
	}
}
