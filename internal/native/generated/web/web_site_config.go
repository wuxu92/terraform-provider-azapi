package web

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
)

// WebSiteCfg carries the Terraform address metadata and dependencies for
// azapi_web_site acceptance-test scenarios. Construct it with NewWebSiteCfg,
// then wrap it in a scenario type when applying.
type WebSiteCfg struct {
	generated.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
	serverFarm    WebServerFarmCfg
}

// NewWebSiteCfg builds a Web site config depending on the parent resource group
// and the native App Service plan config. The label is optional.
func NewWebSiteCfg(resourceGroup resources.ResourceGroupCfg, serverFarm WebServerFarmCfg, label ...string) WebSiteCfg {
	return WebSiteCfg{
		ResourceConfigBase: generated.NewResourceConfigBase(WebSite.Name, label...),
		resourceGroup:      resourceGroup,
		serverFarm:         serverFarm,
	}
}

// WebSiteCfg_Basic is a minimal Windows App Service site bound to the native
// server farm dependency.
type WebSiteCfg_Basic WebSiteCfg

func (r WebSiteCfg_Basic) Config() string {
	return WebSiteCfg(r).config(`
  properties = {
    server_farm_id = %s
  }`)
}

// WebSiteCfg_Complete sets common in-place site properties and site_config
// values shared across AzureRM web-app variants.
type WebSiteCfg_Complete WebSiteCfg

func (r WebSiteCfg_Complete) Config() string {
	return WebSiteCfg(r).config(`
  properties = {
    server_farm_id          = %s
    client_affinity_enabled = false
    client_cert_enabled     = false
    client_cert_mode        = "Required"
    enabled                 = true
    https_only              = true
    public_network_access   = "Enabled"
    site_config = {
      ftps_state          = "Disabled"
      http20_enabled      = false
      min_tls_version     = "1.2"
      scm_min_tls_version = "1.2"
    }
  }`)
}

func (r WebSiteCfg) config(propertiesFmt string) string {
	return fmt.Sprintf(`
resource %q %q {
  name              = "acctest-web-{{.RandomString}}"
  resource_group_id = %s
  location          = "{{.Location}}"
  kind              = "app"%s
}
`, r.ResourceType(), r.ResourceLabel(), r.resourceGroup.IDRef(), fmt.Sprintf(propertiesFmt, r.serverFarm.IDRef()))
}
