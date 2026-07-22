package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VpnSite provides resource knowledge for Microsoft.Network/vpnSites.
//
// Mirrors azurerm_vpn_site. name and resource_group_name are envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/vpn_site_resource.go
//     (schema 46-183, expand 209-220, timeouts 39-44)
//   - internal/services/network/validate/vpn_site_name.go:13 (name regex)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_vpnsite.go:104 (segment casing), model_vpnsiteproperties.go
//
// Not encoded (deliberate):
//   - link (properties.vpnSiteLinks[*]) and o365_policy nested booleans are array/
//     nested shapes; their per-element rules cannot be lowered by azwise/azapin.
type VpnSite struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VpnSite)(nil)

// NewVpnSite returns knowledge for the vpnSites resource.
func NewVpnSite() *VpnSite {
	return &VpnSite{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/vpnSites",
			ApiVersions:  []string{"2025-01-01"},
			// location and virtual_wan_id replace the site.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.virtualWan.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: must not contain '<>%&:?/+
					Regex:   `^[^'<>%&:?/+]+$`,
					Message: "name must not contain the characters '<>%&:?/+",
				},
			},
			RequiredFields: []string{
				"properties.virtualWan.id",
			},
		},
	}
}

func init() { azwise.Register(NewVpnSite()) }
