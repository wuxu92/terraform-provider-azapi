package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateLinkScopedService provides resource knowledge for
// Microsoft.Insights/privateLinkScopes/scopedResources.
//
// Mirrors azurerm_monitor_private_link_scoped_service.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_private_link_scoped_service_resource.go
//     Schema (45-73): name Required+ForceNew (StringIsNotEmpty); resource_group_name;
//     scope_name Required+ForceNew (parent PrivateLinkScopeName); linked_resource_id
//     Required+ForceNew (App Insights component / LA workspace / DCE id) →
//     properties.linkedResourceId. Create-only (no Update). Timeouts Create/Delete 30m,
//     Read 5m.
//   - go-azure-sdk resource-manager/insights/2019-10-17-preview/privatelinkscopedresources:
//     ScopedResourceProperties.linkedResourceId (settable), provisioningState read-only;
//     id_scopedresource.go type Microsoft.Insights/privateLinkScopes/scopedResources.
type PrivateLinkScopedService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateLinkScopedService)(nil)

// NewPrivateLinkScopedService returns knowledge for the scopedResources resource.
func NewPrivateLinkScopedService() *PrivateLinkScopedService {
	return &PrivateLinkScopedService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/privateLinkScopes/scopedResources",
			ApiVersions:  []string{"2019-10-17-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.linkedResourceId"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "name must not be empty",
				},
			},
			// linked_resource_id is Required.
			RequiredFields: []string{
				"properties.linkedResourceId",
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewPrivateLinkScopedService()) }
