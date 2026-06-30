package resource

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
)

func init() {
	RegisterHooks("azapi_web_site", &Hooks{
		AfterCreate: webSiteReadConfiguration,
		AfterUpdate: webSiteReadConfiguration,
		AfterRead:   webSiteReadConfiguration,
	})
}

// webSiteReadConfiguration mirrors AzureRM's explicit configuration read: the
// Microsoft.Web/sites GET response documents siteConfig but does not reliably
// return it. Bicep models the separate child resource Microsoft.Web/sites/config
// with discriminator name "web"; its ID is /sites/{name}/config/web and its
// properties type is the same SiteConfig object used by properties.siteConfig.
func webSiteReadConfiguration(c *CrudCtx) {
	if c == nil || c.Client == nil || c.Client.ResourceClient == nil {
		return
	}
	config, err := c.Client.ResourceClient.Get(c.Ctx, webSiteConfigurationID(c.ID.AzureResourceId), c.ID.ApiVersion, clients.DefaultRequestOptions())
	if err != nil {
		c.Diags.AddError("Failed to retrieve web site configuration", fmt.Errorf("reading configuration for %s: %w", c.ID.AzureResourceId, err).Error())
		return
	}
	mergeWebSiteConfiguration(c.Response, asMap(config))
}

func webSiteConfigurationID(siteID string) string { return siteID + "/config/web" }

func mergeWebSiteConfiguration(site, config map[string]interface{}) {
	configProps := asMap(config["properties"])
	if len(configProps) == 0 {
		return
	}
	siteProps := asMap(site["properties"])
	if len(siteProps) == 0 {
		siteProps = map[string]interface{}{}
		site["properties"] = siteProps
	}
	siteProps["siteConfig"] = configProps
}
