package managedidentity

import (
	"time"

	"github.com/wuxu92/azwise"
)

// UserAssignedIdentity provides resource knowledge for
// Microsoft.ManagedIdentity/userAssignedIdentities.
//
// Mirrors azurerm_user_assigned_identity. The identity name and resource group
// live on the operational envelope (Required + RequiresReplace by construction),
// so their ForceNew is not repeated here; only the body-path knowledge is.
//
// isolation_scope is deliberately left to the mechanical bicep enum
// (None | Regional): AzureRM restricts the exposed set to Regional and maps None
// back to empty on read, but azapin keeps the raw ARM surface (Optional+Computed
// with OneOf("None","Regional")), so no azwise rule narrows it.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedidentity/user_assigned_identity_resource.go
//   - Arguments(): location = commonschema.Location (ForceNew), name/resource_group_name ForceNew (envelope)
//   - Attributes(): client_id, principal_id, tenant_id are Computed (read-only)
//   - Create()/Read()/Update()/Delete() timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/managedidentity/2024-11-30/identities model:
//     UserAssignedIdentityProperties.{ClientId,PrincipalId,TenantId} are absent from
//     the Create/Update model (read-only), so they are stripped from the PUT body.
type UserAssignedIdentity struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*UserAssignedIdentity)(nil)

// NewUserAssignedIdentity returns knowledge for the userAssignedIdentities resource.
func NewUserAssignedIdentity() *UserAssignedIdentity {
	return &UserAssignedIdentity{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ManagedIdentity/userAssignedIdentities",
			ApiVersions:  []string{"2024-11-30"},
			// Changing the geo-location replaces the identity (commonschema.Location
			// is ForceNew). name/resource_group are envelope-owned and already
			// Required+RequiresReplace by construction.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Server-populated identity coordinates: present in the GET response but
			// absent from the ARM Create/Update model. Stripped from the PUT body so a
			// value carried forward in state never rides into an update request.
			ComputedFields: []string{
				"properties.clientId",
				"properties.principalId",
				"properties.tenantId",
			},
		},
	}
}

func init() { azwise.Register(NewUserAssignedIdentity()) }
