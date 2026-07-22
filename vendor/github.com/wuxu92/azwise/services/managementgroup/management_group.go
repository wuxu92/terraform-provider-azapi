package managementgroup

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagementGroup provides resource knowledge for
// Microsoft.Management/managementGroups.
//
// Mirrors azurerm_management_group. The group name is the URL segment (envelope);
// when omitted AzureRM generates a UUID. The create body carries only
// properties.displayName, properties.details.parent.id and properties.tenantId
// (the latter injected from the caller's tenant). subscription_ids are not part of
// this body — membership is managed through managementGroups/subscriptions child
// PUTs (see ManagementGroupSubscriptionAssociation).
//
// Sources:
//   - terraform-provider-azurerm internal/services/managementgroup/management_group_resource.go
//   - Schema(): name (Optional+Computed+ForceNew, validate.ManagementGroupName),
//     display_name (Optional+Computed -> properties.displayName),
//     parent_management_group_id (Optional+Computed -> properties.details.parent.id)
//   - resourceManagementGroupCreateUpdate(): properties.tenantId injected from account tenant
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/management/2020-05-01/managementgroups
//     CreateManagementGroupProperties.{DisplayName, Details.Parent.Id, TenantId}
type ManagementGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagementGroup)(nil)

// NewManagementGroup returns knowledge for the managementGroups resource.
func NewManagementGroup() *ManagementGroup {
	return &ManagementGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Management/managementGroups",
			ApiVersions:  []string{"2020-05-01"},
			// name is ForceNew but is the URL segment (envelope), so no body ForceNew rule.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name: validate.ManagementGroupName — ASCII letters, digits and
				// - _ ( ) . up to length 90.
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    90,
					Regex:        `^[a-zA-Z0-9_().-]{1,90}$`,
					Message:      "name must be ASCII letters, digits, or - _ ( ) . and at most 90 characters",
				},
			},
		},
	}
}

func init() { azwise.Register(NewManagementGroup()) }
