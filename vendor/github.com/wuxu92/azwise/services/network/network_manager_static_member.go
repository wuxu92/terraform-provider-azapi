package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerStaticMember provides resource knowledge for
// Microsoft.Network/networkManagers/networkGroups/staticMembers.
//
// Mirrors azurerm_network_manager_static_member. name and the parent
// network_group_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_static_member_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/staticmembers:
//     model_staticmemberproperties.go
//
// Not encoded (deliberate):
//   - region is Computed in AzureRM but lives on the settable StaticMemberProperties.Region
//     field (present in the Create model), so it is NOT a ComputedField.
//   - target_virtual_network_id validates with a resource-ID validator (VirtualNetwork or
//     Subnet); a semantic rule that belongs on an azapin customizer, not expressible as a
//     declarative StringRule here.
type NetworkManagerStaticMember struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerStaticMember)(nil)

func NewNetworkManagerStaticMember() *NetworkManagerStaticMember {
	return &NetworkManagerStaticMember{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/networkGroups/staticMembers",
			ApiVersions:  []string{"2025-01-01"},
			// target_virtual_network_id (properties.resourceId) is ForceNew in AzureRM.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resourceId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{"properties.resourceId"},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerStaticMember()) }
