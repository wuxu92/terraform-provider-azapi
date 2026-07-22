package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PublicIP provides resource knowledge for Microsoft.Network/publicIPAddresses.
//
// Mirrors azurerm_public_ip. sku / sku_tier expand into the top-level
// sku.name / sku.tier; the remaining settable fields live under properties.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/public_ip_resource.go
//     (resourcePublicIp schema lines 36-205, expand lines 209-330)
//   - go-azure-sdk resource-manager/network/2025-01-01/publicipaddresses:
//     id via commonids PublicIPAddressId (segment "publicIPAddresses"),
//     model_publicipaddresspropertiesformat.go / model_publicipaddressdnssettings.go
//     (json tags), constants.go (IPAllocationMethod, DdosSettingsProtectionMode,
//     IPVersion, PublicIPAddressSkuName, PublicIPAddressSkuTier,
//     PublicIPAddressDnsSettingsDomainNameLabelScope)
//
// Not encoded (deliberate):
//   - fqdn / ip_address are Computed read-only (properties.dnsSettings.fqdn,
//     properties.ipAddress); already bicep ReadOnly.
//   - the CustomizeDiff Basic-SKU deprecation guard and the "sku must be Standard
//     when sku_tier is Global" rule are cross-field semantic checks with no single
//     ARM body field; they are not expressible as declarative StringRules.
type PublicIP struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PublicIP)(nil)

// NewPublicIP returns knowledge for the publicIPAddresses resource.
func NewPublicIP() *PublicIP {
	return &PublicIP{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/publicIPAddresses",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation"}, // edge_zone
				{PropertyPath: "properties.publicIPAddressVersion"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "sku.tier"},
				{PropertyPath: "properties.publicIPPrefix.id"},
				{PropertyPath: "properties.ipTags"},
				{PropertyPath: "zones"},
				// domain_name_label_scope is conditionally ForceNew (ForceNewIfChange:
				// replaces when previously set or being cleared).
				{PropertyPath: "properties.dnsSettings.domainNameLabelScope"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.publicIPAllocationMethod",
					AllowedValues: []string{"Static", "Dynamic"},
					Message:       "must be Static or Dynamic",
				},
				{
					PropertyPath:  "properties.ddosSettings.protectionMode",
					AllowedValues: []string{"Disabled", "Enabled", "VirtualNetworkInherited"},
					Message:       "must be Disabled, Enabled or VirtualNetworkInherited",
				},
				{
					PropertyPath:  "properties.publicIPAddressVersion",
					AllowedValues: []string{"IPv4", "IPv6"},
					Message:       "must be IPv4 or IPv6",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Standard", "StandardV2"},
					Message:       "must be Basic, Standard or StandardV2",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Global", "Regional"},
					Message:       "must be Global or Regional",
				},
				{
					PropertyPath:  "properties.dnsSettings.domainNameLabelScope",
					AllowedValues: []string{"NoReuse", "ResourceGroupReuse", "SubscriptionReuse", "TenantReuse"},
					Message:       "must be NoReuse, ResourceGroupReuse, SubscriptionReuse or TenantReuse",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.idleTimeoutInMinutes",
					MinValue:     azwise.Ptr(int64(4)),
					MaxValue:     azwise.Ptr(int64(30)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.ddosSettings.protectionMode", Value: "VirtualNetworkInherited"},
				{PropertyPath: "properties.publicIPAddressVersion", Value: "IPv4"},
				{PropertyPath: "sku.name", Value: "Standard"},
				{PropertyPath: "sku.tier", Value: "Regional"},
				{PropertyPath: "properties.idleTimeoutInMinutes", Value: float64(4)},
			},
			RequiredFields: []string{
				"properties.publicIPAllocationMethod",
			},
		},
	}
}

func init() { azwise.Register(NewPublicIP()) }
