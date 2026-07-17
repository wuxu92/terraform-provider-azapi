package resources

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
)

// ResourceGroupCfg carries the Terraform address metadata for azapi_resource_group
// acceptance-test scenarios. Construct it with NewResourceGroupCfg, then wrap it in a
// scenario type (ResourceGroupCfg_Basic, ResourceGroupCfg_Named) when applying. The
// returned HCL is a template — {{.RandomInteger}}, {{.Location}} and
// {{.SubscriptionID}} are filled by the acceptance framework's renderer.
type ResourceGroupCfg struct {
	config.ResourceConfigBase
}

// NewResourceGroupCfg builds a resource-group config. The label is optional — omit it
// for the single-instance default ("test"), or pass an explicit label when a scope
// holds more than one. The resource type is read from the ResourceGroup descriptor.
func NewResourceGroupCfg(label ...string) ResourceGroupCfg {
	return ResourceGroupCfg{config.NewResourceConfigBase(ResourceGroup.Name, label...)}
}

// ResourceGroupCfg_Basic is a minimal resource group named acctest-rg-<n>.
type ResourceGroupCfg_Basic ResourceGroupCfg

func (r ResourceGroupCfg_Basic) Config() string {
	return ResourceGroupCfg(r).config("accazapi-rg-{{.RandomInteger}}", "")
}

// ResourceGroupCfg_Complete sets the resource group's whole caller-settable surface:
// a tag map plus managed_by. name and location are ForceNew and properties/system_data
// are Computed-only, so tags and managed_by are the only in-place-writable attributes —
// there is no properties block to set. managed_by carries the AzureRM-proven literal
// "test" (azurerm_resource_group's withManagedBy scenario).
type ResourceGroupCfg_Complete ResourceGroupCfg

func (r ResourceGroupCfg_Complete) Config() string {
	return ResourceGroupCfg(r).config("accazapi-rg-{{.RandomInteger}}", `
  managed_by = "test"
  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }`)
}

// ResourceGroupCfg_Complete_update flips the in-place-updatable surface set by
// ResourceGroupCfg_Complete to prove it survives Update -> Read -> empty plan: the tag
// map is rewritten (environment changed, cost_center dropped), mirroring azurerm's
// withTags -> withTagsUpdated chain. managed_by is held at "test": it is optional and
// non-ForceNew, but no AzureRM scenario chains an in-place managed_by change, and
// reassigning a resource group's management owner mid-test is not a proven round-trip.
type ResourceGroupCfg_Complete_update ResourceGroupCfg

func (r ResourceGroupCfg_Complete_update) Config() string {
	return ResourceGroupCfg(r).config("accazapi-rg-{{.RandomInteger}}", `
  managed_by = "test" // held: no proven in-place managed_by update chain; reassigning management ownership mid-test is not a safe round-trip
  tags = {
    environment = "staging"
  }`)
}

// ResourceGroupCfg_Named renders a resource group with an explicit name, for
// validation scenarios (e.g. asserting a name that violates the schema regex is
// rejected at plan time).
type ResourceGroupCfg_Named struct {
	ResourceGroupCfg
	Name string
}

func (r ResourceGroupCfg_Named) Config() string {
	return r.config(r.Name, "")
}

func (r ResourceGroupCfg) config(name, body string) string {
	return r.RenderConfig(config.ConfigEnvelope{Name: name,
		ParentAttr: "subscription_id",
		ParentRef:  `"/subscriptions/{{.SubscriptionID}}"`,
		Location:   true,
		Body:       body})
}

// ResourceGroupDataCfg is the azapi_resource_group data source config for
// acceptance. It reads back the resource group created by the ResourceGroupCfg
// resource (same label), referencing that resource's name and subscription_id so
// Terraform reads the data source after the resource is created. Construct it with
// NewResourceGroupDataCfg.
type ResourceGroupDataCfg struct {
	config.DataSourceConfigBase
}

// NewResourceGroupDataCfg builds a resource-group data source config. The label
// defaults to "test" to match the single-instance ResourceGroupCfg it reads back.
func NewResourceGroupDataCfg(label ...string) ResourceGroupDataCfg {
	return ResourceGroupDataCfg{config.NewDataSourceConfigBase(ResourceGroup.Name, label...)}
}

// Config renders a data block that reads the resource group by the managed
// resource's own name and subscription_id. depends_on forces the read to apply time
// (after the create) — the inputs are known at plan, so without it Terraform would
// read the data source during plan, before the resource group exists.
func (r ResourceGroupDataCfg) Config() string {
	res := "azapi_resource_group." + r.Label()
	return `data "azapi_resource_group" "` + r.Label() + `" {
  name            = ` + res + `.name
  subscription_id = ` + res + `.subscription_id
  depends_on      = [` + res + `]
}`
}
