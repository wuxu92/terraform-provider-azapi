package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerScopeConnection provides resource knowledge for
// Microsoft.Network/networkManagers/scopeConnections.
//
// Mirrors azurerm_network_manager_scope_connection. name and the parent
// network_manager_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_scope_connection_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/scopeconnections:
//     model_scopeconnectionproperties.go
//
// Not encoded (deliberate):
//   - connection_state is Computed in AzureRM but lives on the settable
//     ScopeConnectionProperties.ConnectionState field (present in the Create model), so
//     it is NOT a ComputedField.
//   - tenant_id validates with validation.IsUUID — a generic semantic rule that belongs
//     on an azapin customizer (validators.UUID against properties.tenantId), not
//     expressible as a declarative StringRule here.
type NetworkManagerScopeConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerScopeConnection)(nil)

func NewNetworkManagerScopeConnection() *NetworkManagerScopeConnection {
	return &NetworkManagerScopeConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/scopeConnections",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// target_scope_id (properties.resourceId) and tenant_id (properties.tenantId)
			// are both Required in AzureRM.
			RequiredFields: []string{
				"properties.resourceId",
				"properties.tenantId",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerScopeConnection()) }
