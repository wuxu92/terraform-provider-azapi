package web

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// WebServerFarmCfg carries the Terraform address metadata and resource-group
// dependency for azapi_web_server_farm acceptance-test scenarios. Construct it
// with NewWebServerFarmCfg, then wrap it in a scenario type when applying.
type WebServerFarmCfg struct {
	services.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewWebServerFarmCfg builds an App Service plan config depending on the parent
// resource group. The label is optional; omit it for the default single plan.
func NewWebServerFarmCfg(resourceGroup resources.ResourceGroupCfg, label ...string) WebServerFarmCfg {
	return WebServerFarmCfg{
		ResourceConfigBase: services.NewResourceConfigBase(WebServerFarm.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// WebServerFarmCfg_Basic is a free Windows App Service plan. Free SKU keeps the
// live acceptance dependency small while still creating the required server farm.
type WebServerFarmCfg_Basic WebServerFarmCfg

func (r WebServerFarmCfg_Basic) Config() string {
	return WebServerFarmCfg(r).config("B1", 1, `
  properties = {
    reserved       = false
    hyper_v        = false
    zone_redundant = false
  }`)
}

// WebServerFarmCfg_Complete exercises common in-place server farm settings that
// are valid on the free plan.
type WebServerFarmCfg_Complete WebServerFarmCfg

func (r WebServerFarmCfg_Complete) Config() string {
	return WebServerFarmCfg(r).config("B1", 1, `
  properties = {
    reserved              = false
    hyper_v               = false
    per_site_scaling      = false
    elastic_scale_enabled = false
    zone_redundant        = false
  }`)
}

func (r WebServerFarmCfg) config(skuName string, capacity int, extra string) string {
	return fmt.Sprintf(`
resource %q %q {
  name              = "acctest-asp-{{.RandomString}}"
  resource_group_id = %s
  location          = "{{.Location}}"
  kind              = "app"
  sku = {
    name     = %q
    capacity = %d
  }%s
}
`, r.ResourceType(), r.ResourceLabel(), r.resourceGroup.IDRef(), skuName, capacity, extra)
}
