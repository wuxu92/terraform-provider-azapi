package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnEndpoint provides resource knowledge for Microsoft.Cdn/profiles/endpoints
// (classic CDN endpoint).
//
// Contributing Terraform resource: azurerm_cdn_endpoint.
//
// Sources:
//   - AzureRM internal/services/cdn/cdn_endpoint_resource.go:27-220 (schema: ForceNew,
//     defaults, enum ValidateFuncs, Required origin, Computed fqdn, 30m timeouts)
//   - AzureRM internal/services/cdn/cdn_endpoint_resource.go:222-331 (create expand: ARM body mapping)
//   - AzureRM internal/services/cdn/cdn_endpoint_resource.go:466-535 (read flatten: d.Set ARM paths)
//   - azure-sdk-for-go services/cdn/mgmt/2020-09-01/cdn/models.go:5747 (EndpointProperties json tags)
//   - azure-sdk-for-go services/cdn/mgmt/2020-09-01/cdn/enums.go (QueryStringCachingBehavior,
//     OptimizationType enum values)
//
// Notes on constraints that cannot be expressed declaratively:
//   - The `origin` block is a Required, ForceNew TypeSet mapping to properties.origins.
//     Its sub-fields (name, host_name, http_port default 80, https_port default 443) are
//     array-element paths (properties.origins[*].*) which azwise/the generator cannot
//     resolve, so their per-element ForceNew and defaults are documented here only.
//   - The `geo_filter.action` enum (Allow/Block) is likewise an array-element path
//     (properties.geoFilters[*].action) and is not emitted as a StringRule.
type CdnEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnEndpoint)(nil)

// NewCdnEndpoint returns a CdnEndpoint knowledge instance.
func NewCdnEndpoint() *CdnEndpoint {
	return &CdnEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/endpoints",
			ApiVersions:  []string{"2020-09-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				// `origin` TypeSet is ForceNew as a whole; the array itself replaces on change.
				{PropertyPath: "properties.origins"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// querystring_caching_behaviour — validation.StringInSlice;
					// full ARM QueryStringCachingBehavior enum.
					PropertyPath:  "properties.queryStringCachingBehavior",
					AllowedValues: []string{"BypassCaching", "IgnoreQueryString", "NotSet", "UseQueryString"},
					Message:       "must be one of BypassCaching, IgnoreQueryString, NotSet, UseQueryString",
				},
				{
					// optimization_type — validation.StringInSlice; full ARM OptimizationType enum.
					PropertyPath: "properties.optimizationType",
					AllowedValues: []string{
						"DynamicSiteAcceleration", "GeneralMediaStreaming", "GeneralWebDelivery",
						"LargeFileDownload", "VideoOnDemandMediaStreaming",
					},
					Message: "must be a valid CDN endpoint optimization type",
				},
			},
			// fqdn (properties.hostName) is READ-ONLY in EndpointProperties (models.go:5748).
			ComputedFields: []string{"properties.hostName"},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isHttpAllowed", Value: true},
				{PropertyPath: "properties.isHttpsAllowed", Value: true},
				{PropertyPath: "properties.queryStringCachingBehavior", Value: "IgnoreQueryString"},
			},
			// `origin` (Required TypeSet) maps to properties.origins.
			RequiredFields: []string{"properties.origins"},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCdnEndpoint()) }
