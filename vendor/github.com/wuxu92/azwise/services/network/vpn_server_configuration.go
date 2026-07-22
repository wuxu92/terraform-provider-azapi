package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VpnServerConfiguration provides resource knowledge for
// Microsoft.Network/vpnServerConfigurations.
//
// Mirrors azurerm_vpn_server_configuration. name and resource_group_name are
// envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/vpn_server_configuration_resource.go
//     (schema 43-267, expand 335-367, timeouts 36-41)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_vpnserverconfiguration.go:104 (segment casing),
//     model_vpnserverconfigurationproperties.go, model_radiusserver.go
//
// Not encoded (deliberate):
//   - vpn_authentication_types (Required) and vpn_protocols expand to enum arrays
//     (properties.vpnAuthenticationTypes[*] / properties.vpnProtocols[*]); their
//     per-element enums (AAD/Certificate/Radius, DhGroup/IkeEncryption/etc. in
//     ipsec_policy) live under array elements and cannot be lowered by azwise/azapin.
//   - AzureRM's "azure_active_directory_authentication required when auth type AAD"
//     is a cross-field semantic rule over an array; not representable declaratively.
type VpnServerConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VpnServerConfiguration)(nil)

// NewVpnServerConfiguration returns knowledge for the vpnServerConfigurations resource.
func NewVpnServerConfiguration() *VpnServerConfiguration {
	return &VpnServerConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/vpnServerConfigurations",
			ApiVersions:  []string{"2025-01-01"},
			// location replaces the resource; all other fields are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			SensitiveFields: []string{
				"properties.radiusServerSecret",
			},
			RequiredFields: []string{
				"properties.vpnAuthenticationTypes",
			},
		},
	}
}

func init() { azwise.Register(NewVpnServerConfiguration()) }
