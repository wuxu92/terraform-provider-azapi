package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorRoute provides resource knowledge for
// Microsoft.Cdn/profiles/afdEndpoints/routes (Azure Front Door route).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_route.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_route_resource.go
//     (schema L43-201, Create body L280-290, cache expand L595-620)
//   - internal/services/cdn/validate/front_door_route_name.go (name regex)
//   - go-azure-sdk cdn/mgmt/2021-06-01/cdn models.go (RouteProperties,
//     AfdRouteCacheConfiguration, CompressionSettings) — the ARM body shape is stable
//     at the 2025-12-01 profiles/afdEndpoints/routes API.
//
// Notes:
//   - name & cdn_frontdoor_endpoint_id are envelope/parent-owned (ForceNew).
//   - cdn_frontdoor_origin_ids is provisioning-order-only and is NOT sent to the API.
//   - cdn_frontdoor_origin_group_id -> properties.originGroup.id,
//     cdn_frontdoor_custom_domain_ids -> properties.customDomains[*].id, and
//     cdn_frontdoor_rule_set_ids -> properties.ruleSets[*].id are ResourceReference
//     IDs validated by semantic validators at the azapin layer.
//   - enabled / https_redirect_enabled / link_to_default_domain are bools mapped to
//     Enabled/Disabled ARM enums.
type CdnFrontDoorRoute struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorRoute)(nil)

func NewCdnFrontDoorRoute() *CdnFrontDoorRoute {
	return &CdnFrontDoorRoute{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/afdEndpoints/routes",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorRouteName
				{
					Regex:     `^[\da-zA-Z][-\da-zA-Z]{0,88}[\da-zA-Z]$`,
					MinLength: 2,
					MaxLength: 90,
					Message:   "must be 2-90 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
				// forwarding_protocol → properties.forwardingProtocol
				{
					PropertyPath:  "properties.forwardingProtocol",
					AllowedValues: []string{"HttpOnly", "HttpsOnly", "MatchRequest"},
				},
				// cache.query_string_caching_behavior → properties.cacheConfiguration.queryStringCachingBehavior
				{
					PropertyPath:  "properties.cacheConfiguration.queryStringCachingBehavior",
					AllowedValues: []string{"IgnoreQueryString", "IgnoreSpecifiedQueryStrings", "IncludeSpecifiedQueryStrings", "UseQueryString"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enabledState", Value: "Enabled"},
				{PropertyPath: "properties.httpsRedirect", Value: "Enabled"},
				{PropertyPath: "properties.linkToDefaultDomain", Value: "Enabled"},
				{PropertyPath: "properties.forwardingProtocol", Value: "MatchRequest"},
				{PropertyPath: "properties.cacheConfiguration.queryStringCachingBehavior", Value: "IgnoreQueryString"},
				{PropertyPath: "properties.cacheConfiguration.compressionSettings.isCompressionEnabled", Value: false},
			},
			ComputedFields: []string{
				"properties.endpointName",
				"properties.provisioningState",
				"properties.deploymentStatus",
			},
			// patterns_to_match, supported_protocols and origin_group are Required.
			RequiredFields: []string{
				"properties.patternsToMatch",
				"properties.supportedProtocols",
				"properties.originGroup",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorRoute()) }
