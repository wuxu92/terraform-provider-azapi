package web

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/native/armjson"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
)

func init() {
	nativeresource.RegisterHooks(WebSite.Name, &nativeresource.Hooks{
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
func webSiteReadConfiguration(c *nativeresource.CrudCtx) {
	if c == nil || c.Client == nil || c.Client.ResourceClient == nil {
		return
	}
	config, err := c.Client.ResourceClient.Get(c.Ctx, webSiteConfigurationID(c.ID.AzureResourceId), c.ID.ApiVersion, clients.DefaultRequestOptions())
	if err != nil {
		c.Diags.AddError("Failed to retrieve web site configuration", fmt.Errorf("reading configuration for %s: %w", c.ID.AzureResourceId, err).Error())
		return
	}
	normalizeWebSiteResponse(c.Response, armjson.AsMap(config))
}

func webSiteConfigurationID(siteID string) string { return siteID + "/config/web" }

func normalizeWebSiteResponse(site, config map[string]interface{}) {
	normalizeWebSiteProperties(site)
	mergeWebSiteConfiguration(site, config)
}

func normalizeWebSiteProperties(site map[string]interface{}) {
	props, ok := site["properties"].(map[string]interface{})
	if !ok || props == nil {
		props = map[string]interface{}{}
		site["properties"] = props
	}
	if _, ok := props["clientAffinityPartitioningEnabled"]; !ok {
		props["clientAffinityPartitioningEnabled"] = false
	}
}

func mergeWebSiteConfiguration(site, config map[string]interface{}) {
	configProps := armjson.AsMap(config["properties"])
	if len(configProps) == 0 {
		return
	}
	normalizeWebSiteConfiguration(configProps)

	siteProps := armjson.AsMap(site["properties"])
	if len(siteProps) == 0 {
		siteProps = map[string]interface{}{}
		site["properties"] = siteProps
	}
	siteProps["siteConfig"] = configProps
}

func normalizeWebSiteConfiguration(siteConfig map[string]interface{}) {
	filterDefaultAccessRestriction(siteConfig, "ipSecurityRestrictions")
	filterDefaultAccessRestriction(siteConfig, "scmIpSecurityRestrictions")
}

func filterDefaultAccessRestriction(siteConfig map[string]interface{}, key string) {
	restrictions, ok := siteConfig[key].([]interface{})
	if !ok || len(restrictions) == 0 {
		return
	}
	filtered := make([]interface{}, 0, len(restrictions))
	for _, restriction := range restrictions {
		if isDefaultRestriction(armjson.AsMap(restriction)) {
			continue
		}
		filtered = append(filtered, restriction)
	}
	if len(filtered) == 0 {
		delete(siteConfig, key)
		return
	}
	siteConfig[key] = filtered
}

// webSiteDefaultRestrictionPriority is the reserved priority (int32 max) Azure assigns
// to the implicit default access-restriction rule.
const webSiteDefaultRestrictionPriority int64 = 2147483647

// isDefaultRestriction reports whether an access-restriction entry is Azure's implicit
// default rule. ARM tags that rule — the one governed by the *DefaultAction property —
// with the reserved priority above and returns it on every site GET even though it is
// not user-configurable, so the read hook strips it to keep it from drifting against
// the user's configuration. The schema forbids that priority for user rules (see
// customizers.customizeWebSite), so a match is always the default rule.
func isDefaultRestriction(restriction map[string]interface{}) bool {
	return armjson.Int64(restriction["priority"]) == webSiteDefaultRestrictionPriority
}
