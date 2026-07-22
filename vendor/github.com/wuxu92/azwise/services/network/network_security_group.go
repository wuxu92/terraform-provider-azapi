package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkSecurityGroup provides resource knowledge for
// Microsoft.Network/networkSecurityGroups.
//
// Mirrors azurerm_network_security_group. name and resource_group are
// envelope-owned (Required + RequiresReplace by construction); location is the
// only body/envelope path that forces replacement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_security_group_resource.go
//     (schema, expandSecurityRules, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/networksecuritygroups:
//     model_networksecuritygrouppropertiesformat.go, id_networksecuritygroup.go
//     (segment casing "networkSecurityGroups"), constants.go
//
// Not encoded (deliberate):
//   - azurerm_network_security_rule folds into this resource as the
//     properties.securityRules[*] array (see network_security_group_resource.go
//     expandSecurityRules). Its per-rule enums (protocol Any/Tcp/Udp/Icmp/Ah/Esp,
//     access Allow/Deny, direction Inbound/Outbound), priority IntBetween(100,4096)
//     and description StringLenBetween(0,140) all live under array elements, which
//     azwise/azapin cannot lower or resolve through "[*]". Skipped, not emitted.
type NetworkSecurityGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkSecurityGroup)(nil)

// NewNetworkSecurityGroup returns knowledge for the networkSecurityGroups resource.
func NewNetworkSecurityGroup() *NetworkSecurityGroup {
	return &NetworkSecurityGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkSecurityGroups",
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

func init() { azwise.Register(NewNetworkSecurityGroup()) }
