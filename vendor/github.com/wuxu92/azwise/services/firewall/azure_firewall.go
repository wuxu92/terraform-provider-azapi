package firewall

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AzureFirewall provides resource knowledge for Microsoft.Network/azureFirewalls.
//
// Mirrors azurerm_firewall. name, location and resource_group_name live on the
// operational envelope, so their ForceNew is owned by the envelope and not
// repeated here; only body-path knowledge is encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/firewall/firewall_resource.go
//     (resourceFirewall schema, resourceFirewallCreateUpdate, CRUD timeouts lines 46-234)
//   - terraform-provider-azurerm internal/services/firewall/validate/firewall_name.go
//     (FirewallName regex, line 16)
//   - go-azure-sdk resource-manager/network/2025-01-01/azurefirewalls:
//     model_azurefirewall.go, model_azurefirewallpropertiesformat.go,
//     model_azurefirewallsku.go, constants.go (enum values)
//
// Not encoded (deliberate):
//   - ip_configuration.subnet_id and management_ip_configuration.subnet_id are
//     ForceNew but live under array/nested-element paths
//     (properties.ipConfigurations[*].subnet, properties.managementIpConfiguration.subnet);
//     the whole management_ip_configuration block IS ForceNew and is emitted.
//   - dns_servers, dns_proxy_enabled and private_ip_ranges all expand into the
//     properties.additionalProperties string map (map-key writes), which has no
//     single settable ARM body path — skipped.
//   - virtual_hub / management_ip_configuration are MaxItems:1 TF blocks that
//     collapse to single ARM objects (virtualHub SubResource,
//     managementIpConfiguration object), so no ArrayRule MaxItems applies.
type AzureFirewall struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AzureFirewall)(nil)

// NewAzureFirewall returns knowledge for the azureFirewalls resource.
func NewAzureFirewall() *AzureFirewall {
	return &AzureFirewall{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/azureFirewalls",
			ApiVersions:  []string{"2025-01-01"},
			// sku_name (properties.sku.name), the management IP configuration block,
			// and availability zones all replace the firewall on change. sku_tier is
			// updatable (not ForceNew).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sku.name"},
				{PropertyPath: "properties.managementIpConfiguration"},
				{PropertyPath: "zones"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// resource name (empty PropertyPath = name attribute)
					Regex:   `^[0-9a-zA-Z]([0-9a-zA-Z._-]{0,}[0-9a-zA-Z_])?$`,
					Message: "must begin with a letter or number, end with a letter, number or underscore, and may contain only letters, numbers, underscores, periods, or hyphens",
				},
				{
					PropertyPath:  "properties.sku.name",
					AllowedValues: []string{"AZFW_Hub", "AZFW_VNet"},
					Message:       "must be AZFW_Hub or AZFW_VNet",
				},
				{
					PropertyPath:  "properties.sku.tier",
					AllowedValues: []string{"Basic", "Premium", "Standard"},
					Message:       "must be Basic, Premium or Standard",
				},
				{
					PropertyPath:  "properties.threatIntelMode",
					AllowedValues: []string{"Alert", "Deny", "Off"},
					Message:       "must be Alert, Deny or Off",
				},
			},
			// provisioningState and ipGroups are populated by Azure in the GET
			// response and never set by the user.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.ipGroups",
			},
			// sku_name and sku_tier are both Required in AzureRM.
			RequiredFields: []string{
				"properties.sku.name",
				"properties.sku.tier",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewAzureFirewall()) }
