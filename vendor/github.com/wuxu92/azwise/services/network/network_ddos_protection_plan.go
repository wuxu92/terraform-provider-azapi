package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkDdosProtectionPlan provides resource knowledge for
// Microsoft.Network/ddosProtectionPlans.
//
// Mirrors azurerm_network_ddos_protection_plan. Create sends only Location +
// Tags; all body properties are server-managed.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_ddos_protection_plan_resource.go
//     (schema, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/ddosprotectionplans:
//     model_ddosprotectionplanpropertiesformat.go, id_ddosprotectionplan.go
//     (segment casing "ddosProtectionPlans")
//
// Not encoded (deliberate):
//   - properties.virtualNetworks, properties.publicIPAddresses,
//     properties.resourceGuid and properties.provisioningState are all read-only
//     in the ARM model (bicep ReadOnly, flags=2); the generator already marks them
//     Computed and strips them from the PUT body, so no azwise ComputedFields entry
//     is required.
type NetworkDdosProtectionPlan struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkDdosProtectionPlan)(nil)

// NewNetworkDdosProtectionPlan returns knowledge for the ddosProtectionPlans resource.
func NewNetworkDdosProtectionPlan() *NetworkDdosProtectionPlan {
	return &NetworkDdosProtectionPlan{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/ddosProtectionPlans",
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

func init() { azwise.Register(NewNetworkDdosProtectionPlan()) }
