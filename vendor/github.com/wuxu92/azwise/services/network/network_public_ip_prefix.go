package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PublicIPPrefix provides resource knowledge for Microsoft.Network/publicIPPrefixes.
//
// Mirrors azurerm_public_ip_prefix. sku / sku_tier expand into the top-level
// sku.name / sku.tier; prefix_length and ip_version live under properties.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/public_ip_prefix_resource.go
//     (resourcePublicIpPrefix schema lines 33-127, expand lines 149-166)
//   - go-azure-sdk resource-manager/network/2025-01-01/publicipprefixes:
//     id_publicipprefix.go (ARM type segment "publicIPPrefixes"),
//     model_publicipprefixpropertiesformat.go (prefixLength, publicIPAddressVersion,
//     customIPPrefix json tags), constants.go (PublicIPPrefixSkuName,
//     PublicIPPrefixSkuTier, IPVersion)
//
// Not encoded (deliberate):
//   - ip_prefix is Computed read-only (properties.ipPrefix); already bicep ReadOnly.
//   - the CustomizeDiff "sku must be Standard when sku_tier is Global" rule is a
//     cross-field semantic check with no single ARM body field.
type PublicIPPrefix struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PublicIPPrefix)(nil)

// NewPublicIPPrefix returns knowledge for the publicIPPrefixes resource.
func NewPublicIPPrefix() *PublicIPPrefix {
	return &PublicIPPrefix{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/publicIPPrefixes",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.customIPPrefix.id"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "sku.tier"},
				{PropertyPath: "properties.prefixLength"},
				{PropertyPath: "properties.publicIPAddressVersion"},
				{PropertyPath: "zones"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Standard", "StandardV2"},
					Message:       "must be Standard or StandardV2",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Global", "Regional"},
					Message:       "must be Global or Regional",
				},
				{
					PropertyPath:  "properties.publicIPAddressVersion",
					AllowedValues: []string{"IPv4", "IPv6"},
					Message:       "must be IPv4 or IPv6",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.prefixLength",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(127)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "Standard"},
				{PropertyPath: "sku.tier", Value: "Regional"},
				{PropertyPath: "properties.prefixLength", Value: float64(28)},
				{PropertyPath: "properties.publicIPAddressVersion", Value: "IPv4"},
			},
		},
	}
}

func init() { azwise.Register(NewPublicIPPrefix()) }
