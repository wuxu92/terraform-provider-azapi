package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationGateway provides resource knowledge for
// Microsoft.Network/applicationGateways.
//
// Mirrors azurerm_application_gateway. This is a large resource whose settable
// surface is almost entirely nested array blocks (backend pools, listeners,
// routing rules, ...) that live under properties.<collection>[*]; azwise cannot
// resolve rules through array elements, so those are not encoded. The single
// (MaxItems: 1) sku and autoscale_configuration objects, plus the top-level
// scalar toggles, are captured here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/application_gateway_resource.go
//     (schema lines 132-1801: sku 1051-1090, autoscale 1032-1050; expand
//     lines 1837-1996)
//   - go-azure-sdk resource-manager/network/2025-01-01/applicationgateways:
//     id_applicationgateway.go (ARM type segment "applicationGateways"),
//     model_applicationgatewaypropertiesformat.go (sku, autoscaleConfiguration,
//     enableFips, forceFirewallPolicyAssociation, firewallPolicy,
//     backendHttpSettingsCollection/backendSettingsCollection, httpListeners/
//     listeners, requestRoutingRules/routingRules json tags),
//     model_applicationgatewaysku.go (sku.name/tier/capacity),
//     model_applicationgatewayautoscaleconfiguration.go (minCapacity/maxCapacity),
//     constants.go (ApplicationGatewaySkuName, ApplicationGatewayTier)
//
// Not encoded (deliberate):
//   - All array blocks (backend_address_pool, backend_http_settings, backend,
//     frontend_ip_configuration, frontend_port, gateway_ip_configuration,
//     http_listener, listener, probe, request_routing_rule, routing_rule,
//     redirect_configuration, rewrite_rule_set, ssl_certificate, ssl_profile,
//     url_path_map, private_link_configuration, custom_error_configuration,
//     trusted_root_certificate, trusted_client_certificate,
//     authentication_certificate) expand under properties.<collection>[*]; their
//     element-level enums (cookie_based_affinity, protocol, ...) are skipped.
//   - waf_configuration / ssl_policy / global are optional sub-objects whose
//     internal enums are validated by AzureRM CustomizeDiff cross-field checks.
type ApplicationGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationGateway)(nil)

// NewApplicationGateway returns knowledge for the applicationGateways resource.
func NewApplicationGateway() *ApplicationGateway {
	return &ApplicationGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/applicationGateways",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "zones"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.sku.name",
					AllowedValues: []string{
						"Basic", "Standard_Small", "Standard_Medium", "Standard_Large",
						"Standard_v2", "WAF_Large", "WAF_Medium", "WAF_v2",
					},
					Message: "must be a valid Application Gateway SKU name",
				},
				{
					PropertyPath: "properties.sku.tier",
					AllowedValues: []string{
						"Basic", "Standard", "Standard_v2", "WAF", "WAF_v2",
					},
					Message: "must be Basic, Standard, Standard_v2, WAF or WAF_v2",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.autoscaleConfiguration.minCapacity",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(100)),
				},
				{
					PropertyPath: "properties.autoscaleConfiguration.maxCapacity",
					MinValue:     azwise.Ptr(int64(2)),
					MaxValue:     azwise.Ptr(int64(125)),
				},
			},
			RequiredFields: []string{
				"properties.sku",
				"properties.backendAddressPools",
				"properties.frontendIPConfigurations",
				"properties.frontendPorts",
				"properties.gatewayIPConfigurations",
			},
			// AzureRM AtLeastOneOf pairs, each mapping a legacy/renamed block to its
			// replacement. Both members are top-level property collections.
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.backendHttpSettingsCollection", "properties.backendSettingsCollection"},
					Message: "at least one of backend_http_settings or backend must be set",
				},
				{
					Paths:   []string{"properties.httpListeners", "properties.listeners"},
					Message: "at least one of http_listener or listener must be set",
				},
				{
					Paths:   []string{"properties.requestRoutingRules", "properties.routingRules"},
					Message: "at least one of request_routing_rule or routing_rule must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApplicationGateway()) }
