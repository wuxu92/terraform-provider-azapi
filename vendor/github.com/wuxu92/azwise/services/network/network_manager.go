package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManager provides resource knowledge for Microsoft.Network/networkManagers.
//
// Mirrors azurerm_network_manager. name and resource_group_name are envelope-owned
// (Required + RequiresReplace by construction), so their ForceNew is not repeated
// here; only body/envelope-path knowledge is encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_resource.go
//     (schema Arguments, expandNetworkManagerScope, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/networkmanagers:
//     model_networkmanagerproperties.go,
//     model_networkmanagerpropertiesnetworkmanagerscopes.go, constants.go
//
// Not encoded (deliberate):
//   - scope_accesses is an array-of-enum body field (properties.networkManagerScopeAccesses,
//     []ConfigurationType). azwise StringRule.AllowedValues lowers to a scalar OneOf
//     validator and cannot target an array field, so the per-element enum
//     (Connectivity/Routing/SecurityAdmin/SecurityUser) is documented, not emitted.
//   - cross_tenant_scopes is Computed in AzureRM but lives on the settable
//     networkManagerScopes.crossTenantScopes field (present in the Create model), so it
//     is NOT a ComputedField — listing it would strip a user-settable path.
//   - provisioningState / resourceGuid are bicep ReadOnly; the generator strips them
//     mechanically, so no ComputedFields entry is needed.
type NetworkManager struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManager)(nil)

func NewNetworkManager() *NetworkManager {
	return &NetworkManager{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers",
			ApiVersions:  []string{"2025-01-01"},
			// location (commonschema.Location) replaces the manager on change.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// networkManagerScopes is Required (MinItems 1, MaxItems 1) in AzureRM.
			RequiredFields: []string{"properties.networkManagerScopes"},
			// AzureRM AtLeastOneOf on scope.management_group_ids / scope.subscription_ids.
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.networkManagerScopes.managementGroups",
						"properties.networkManagerScopes.subscriptions",
					},
					Message: "at least one of `management_group_ids` or `subscription_ids` must be set on `scope`",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManager()) }
