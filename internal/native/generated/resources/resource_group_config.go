package resources

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
)

// ResourceGroupCfg builds azapi_resource_group acceptance-test configurations.
// Construct it with NewResourceGroupCfg. The returned HCL is a template —
// {{.RandomInteger}}, {{.Location}} and {{.SubscriptionID}} are filled by the
// acceptance framework's renderer. A dependent resource (e.g. a storage account)
// reuses a resource group as its base by calling Basic here and injecting the
// group's Terraform address (the acceptance Resource's IDRef) as its parent
// reference.
type ResourceGroupCfg struct {
	generated.ResourceConfigBase
}

// NewResourceGroupCfg builds a resource-group config with the given Terraform state
// label. The resource type is read from the ResourceGroup descriptor, so the caller
// never restates it.
func NewResourceGroupCfg(label string) ResourceGroupCfg {
	return ResourceGroupCfg{generated.NewResourceConfigBase(ResourceGroup.Name, label)}
}

// Basic is a minimal resource group named acctest-rg-<n>.
func (r ResourceGroupCfg) Basic() string {
	return r.named("acctest-rg-{{.RandomInteger}}")
}

// Named builds a resource group with an explicit name, for validation scenarios
// (e.g. asserting a name that violates the schema regex is rejected at plan time).
func (r ResourceGroupCfg) Named(name string) string {
	return r.named(name)
}

func (r ResourceGroupCfg) named(name string) string {
	return fmt.Sprintf(`
resource %q %q {
  name            = %q
  subscription_id = "/subscriptions/{{.SubscriptionID}}"
  location        = "{{.Location}}"
}
`, r.ResourceType(), r.ResourceLabel(), name)
}
