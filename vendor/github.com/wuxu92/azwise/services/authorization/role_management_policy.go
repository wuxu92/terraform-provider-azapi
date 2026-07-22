package authorization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RoleManagementPolicy provides resource knowledge for
// Microsoft.Authorization/roleManagementPolicies.
//
// Mirrors azurerm_role_management_policy. Every target scope already has a default
// role management policy, so this resource is effectively PATCH-only: AzureRM's
// Create retrieves the existing policy and Updates it, and Delete is a no-op
// (policies cannot be deleted). Consequently there are no ARM-body Required fields
// and no create-time defaults to strip.
//
// The two identity arguments role_definition_id and scope are ForceNew in AzureRM,
// but they are envelope/identity inputs used to locate the policy (the resource ID
// itself changes on every update), not ARM request-body properties, so they are not
// represented as body ForceNew rules here.
//
// TODO: The ARM request body's configurable surface is `properties.rules`, a
// polymorphic array whose elements are discriminated by `ruleType`
// (RoleManagementPolicyApprovalRule / AuthenticationContextRule / EnablementRule /
// ExpirationRule / NotificationRule) and addressed by a stable string `id` such as
// "Expiration_Admin_Eligibility", "Enablement_Admin_Assignment",
// "Approval_EndUser_Assignment", "Notification_Admin_Admin_Assignment", etc. All of
// AzureRM's declarative validators live inside these array elements:
//   - expire_after / maximum_duration StringInSlice ISO-8601 durations
//     (P15D/P30D/P90D/P180D/P365D and PT30M..P1D) on ExpirationRule.maximumDuration
//   - approver `type` StringInSlice {User, Group} and object_id IsUUID on
//     ApprovalRule.setting.approvalStages[*].primaryApprovers[*]
//   - notification_level StringInSlice on NotificationRule.notificationLevel
// azwise's rule model targets stable dot-paths and cannot address array elements
// keyed by a discriminator `id`, so these constraints are intentionally not encoded
// declaratively. They remain enforced by the underlying ARM API.
//
// Sources:
//   - terraform-provider-azurerm internal/services/authorization/role_management_policy_resource.go
//     schema + Create/Update/Read/Delete (Create/Update timeout 30m; Read/Delete 5m;
//     Delete is a no-op; rules mutated in place via buildRoleManagementPolicyForUpdate)
//   - go-azure-sdk resource-manager/authorization/2020-10-01/rolemanagementpolicies
//     RoleManagementPolicyProperties.rules is *[]RoleManagementPolicyRule (polymorphic,
//     ruleType-discriminated); description/displayName/effectiveRules/
//     isOrganizationDefault/lastModified*/policyProperties/scope are read-only.
type RoleManagementPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RoleManagementPolicy)(nil)

// NewRoleManagementPolicy returns knowledge for the roleManagementPolicies resource.
func NewRoleManagementPolicy() *RoleManagementPolicy {
	return &RoleManagementPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/roleManagementPolicies",
			ApiVersions:  []string{"2020-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 5 * time.Minute,
			},
			// Read-only policy properties returned by GET but not settable.
			ComputedFields: []string{
				"properties.effectiveRules",
				"properties.isOrganizationDefault",
				"properties.lastModifiedBy",
				"properties.lastModifiedDateTime",
				"properties.policyProperties",
				"properties.scope",
			},
		},
	}
}

func init() { azwise.Register(NewRoleManagementPolicy()) }
