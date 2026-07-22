package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerConnection provides resource knowledge for
// Microsoft.Network/networkManagerConnections — the scoped connection type shared by
// two AzureRM resources:
//
//   - azurerm_network_manager_subscription_connection (subscription scope:
//     /subscriptions/{sub}/providers/Microsoft.Network/networkManagerConnections/{name})
//   - azurerm_network_manager_management_group_connection (management-group scope:
//     /providers/Microsoft.Management/managementGroups/{mg}/providers/Microsoft.Network/networkManagerConnections/{name})
//
// Both TF resources send the identical NetworkManagerConnectionProperties body, so they
// merge into this single ARM-type file (union of universal knowledge only).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_subscription_connection_resource.go
//   - terraform-provider-azurerm internal/services/network/network_manager_management_group_connection_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/networkmanagerconnections:
//     model_networkmanagerconnectionproperties.go
//
// Not encoded (deliberate):
//   - network_manager_id (properties.networkManagerId) is ForceNew for the
//     management-group connection but updatable for the subscription connection —
//     non-universal, so it is left out of the shared ForceNew.
//   - connection_state is Computed in AzureRM but lives on the settable
//     NetworkManagerConnectionProperties.ConnectionState field (present in the Create
//     model), so it is NOT a ComputedField.
type NetworkManagerConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerConnection)(nil)

func NewNetworkManagerConnection() *NetworkManagerConnection {
	return &NetworkManagerConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagerConnections",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// network_manager_id is Required in both contributing TF resources.
			RequiredFields: []string{"properties.networkManagerId"},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerConnection()) }
