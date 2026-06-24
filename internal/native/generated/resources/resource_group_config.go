package resources

import "fmt"

// ResourceGroup builds azapi_resource_group acceptance-test configurations.
//
// Label is the Terraform state label; the acceptance Resource handle that applies a
// returned config must use the same label. The returned HCL is a template —
// {{.RandomInteger}}, {{.Location}} and {{.SubscriptionID}} are filled by the
// acceptance framework's renderer. A dependent resource (e.g. a storage account)
// reuses a resource group as its base by calling Basic here and injecting the group's
// Terraform address (the acceptance Resource's IDRef) as its parent reference.
type ResourceGroup struct {
	Label string
}

// Basic is a minimal resource group named acctest-rg-<n>.
func (r ResourceGroup) Basic() string {
	return r.named("acctest-rg-{{.RandomInteger}}")
}

// Named builds a resource group with an explicit name, for validation scenarios
// (e.g. asserting a name that violates the schema regex is rejected at plan time).
func (r ResourceGroup) Named(name string) string {
	return r.named(name)
}

func (r ResourceGroup) named(name string) string {
	return fmt.Sprintf(`
resource "azapi_resource_group" %q {
  name            = %q
  subscription_id = "/subscriptions/{{.SubscriptionID}}"
  location        = "{{.Location}}"
}
`, r.Label, name)
}

// ResourceType is the Terraform type this builder configures; with ResourceLabel it
// satisfies acceptance.ResourceConfig, so a scope can vend the matching handle
// directly — acct.ResourceFor(cfg) — without restating the literal type and label.
func (r ResourceGroup) ResourceType() string { return "azapi_resource_group" }

// ResourceLabel is the Terraform state label (see ResourceType).
func (r ResourceGroup) ResourceLabel() string { return r.Label }
