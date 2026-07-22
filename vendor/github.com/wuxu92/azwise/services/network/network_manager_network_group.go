package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerNetworkGroup provides resource knowledge for
// Microsoft.Network/networkManagers/networkGroups.
//
// Mirrors azurerm_network_manager_network_group. name and the parent
// network_manager_id are envelope-owned (Required + RequiresReplace), so they are not
// repeated in ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_network_group_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/networkgroups:
//     model_networkgroupproperties.go, constants.go (GroupMemberType)
type NetworkManagerNetworkGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerNetworkGroup)(nil)

func NewNetworkManagerNetworkGroup() *NetworkManagerNetworkGroup {
	return &NetworkManagerNetworkGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/networkGroups",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.memberType",
					AllowedValues: []string{"Subnet", "VirtualNetwork"},
					Message:       "must be Subnet or VirtualNetwork",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.memberType", Value: "VirtualNetwork"},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerNetworkGroup()) }
