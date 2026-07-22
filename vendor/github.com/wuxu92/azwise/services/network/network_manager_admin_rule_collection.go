package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerAdminRuleCollection provides resource knowledge for
// Microsoft.Network/networkManagers/securityAdminConfigurations/ruleCollections.
//
// Mirrors azurerm_network_manager_admin_rule_collection. name and the parent
// security_admin_configuration_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_admin_rule_collection_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/adminrulecollections:
//     model_adminrulecollectionpropertiesformat.go
//
// Not encoded (deliberate):
//   - network_group_ids elements validate with a NetworkGroup resource-ID validator — a
//     semantic rule inside an array; belongs on an azapin customizer, noted only.
type NetworkManagerAdminRuleCollection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerAdminRuleCollection)(nil)

func NewNetworkManagerAdminRuleCollection() *NetworkManagerAdminRuleCollection {
	return &NetworkManagerAdminRuleCollection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/securityAdminConfigurations/ruleCollections",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// network_group_ids is Required (properties.appliesToGroups).
			RequiredFields: []string{"properties.appliesToGroups"},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerAdminRuleCollection()) }
