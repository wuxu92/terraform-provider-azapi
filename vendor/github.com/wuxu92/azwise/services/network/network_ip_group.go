package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IPGroup provides resource knowledge for Microsoft.Network/ipGroups.
//
// Mirrors azurerm_ip_group. The CIDR list (cidrs) expands into
// properties.ipAddresses; firewall_ids / firewall_policy_ids are read-only
// (computed) back-references and are not settable.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/ip_group_resource.go
//     (resourceIpGroup schema + timeouts lines 34-94, expand lines 138-147)
//   - go-azure-sdk resource-manager/network/2025-01-01/ipgroups:
//     id_ipgroup.go (ARM type segment "ipGroups"),
//     model_ipgrouppropertiesformat.go (ipAddresses json tag)
//
// Not encoded (deliberate):
//   - azurerm_ip_group_cidr manages individual CIDR entries of the SAME ipGroups
//     resource (it folds into properties.ipAddresses on the parent body); it is
//     not a distinct ARM resource type, so no separate knowledge file is emitted.
//   - firewall_ids / firewall_policy_ids are Computed-only back-references
//     (properties.firewalls / properties.firewallPolicies) populated by Azure when
//     a firewall attaches; already bicep ReadOnly, so no ComputedFields entry.
type IPGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IPGroup)(nil)

// NewIPGroup returns knowledge for the ipGroups resource.
func NewIPGroup() *IPGroup {
	return &IPGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/ipGroups",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewIPGroup()) }
