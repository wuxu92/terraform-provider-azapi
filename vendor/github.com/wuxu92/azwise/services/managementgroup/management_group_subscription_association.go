package managementgroup

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagementGroupSubscriptionAssociation provides resource knowledge for
// Microsoft.Management/managementGroups/subscriptions.
//
// Mirrors azurerm_management_group_subscription_association. Both arguments
// (management_group_id and subscription_id) are path segments — the association is
// a bodyless PUT (managementgroups.SubscriptionsCreate with empty options), so
// there are no body properties to constrain, default, or force-replace. Only the
// timeouts carry over.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managementgroup/management_group_subscription_association_resource.go
//   - Schema(): management_group_id (Required, ForceNew), subscription_id (Required, ForceNew) — both path
//   - resourceManagementGroupSubscriptionAssociationCreate(): SubscriptionsCreate has an empty body
//   - Create/Read/Delete timeouts: 5m / 5m / 5m
type ManagementGroupSubscriptionAssociation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagementGroupSubscriptionAssociation)(nil)

// NewManagementGroupSubscriptionAssociation returns knowledge for the
// managementGroups/subscriptions association resource.
func NewManagementGroupSubscriptionAssociation() *ManagementGroupSubscriptionAssociation {
	return &ManagementGroupSubscriptionAssociation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Management/managementGroups/subscriptions",
			ApiVersions:  []string{"2020-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 5 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewManagementGroupSubscriptionAssociation()) }
