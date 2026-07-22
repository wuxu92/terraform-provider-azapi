package authorization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RoleDefinition provides resource knowledge for
// Microsoft.Authorization/roleDefinitions.
//
// Mirrors azurerm_role_definition. The role definition is a scope-based extension
// resource: its ARM resource name is the role definition GUID (azurerm's
// role_definition_id) and its parent is an arbitrary scope (azurerm's scope). Both
// live on the operational envelope (name + parent_id), which is Required +
// RequiresReplace by construction, so azurerm's ForceNew on role_definition_id and
// scope is not repeated here.
//
// The display name and the not-empty / UUID / scope-shape validators are handled at
// the schema layer instead of here:
//   - properties.roleName (azurerm `name`) is promoted to Required and given a
//     not-empty length rule by the customizer.
//   - the envelope name gets a UUID validator (azurerm's role_definition_id IsUUID)
//     in the customizer, because ApplyAzwise skips empty-path (name) StringRules.
//   - assignable_scopes' per-element commonids.ValidateScopeID is deliberately
//     dropped: it accepts the full spread of ARM scope IDs (tenant "/", subscription,
//     resource group, management group, and resource scopes) that no single native
//     shared validator matches, and it targets an array element (unsupported by the
//     overlay). AzAPI keeps the raw ARM string surface for the list.
//
// assignable_scopes defaults to the role's own scope when omitted; the ARM API
// requires at least one assignable scope, and azurerm supplies [scope] via
// expandRoleDefinitionAssignableScopes. That default is dynamic (it is the scope,
// not a static value), so it is applied by a runtime Before-Create/Update hook (see
// role_definition_hooks.go), not a static DefaultValue here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/authorization/role_definition_resource.go
//     schema + Create/Update/Delete (type hardcoded to "CustomRole"; assignable
//     scopes default to [scope]; timeouts 30m/5m/60m/30m)
//   - go-azure-sdk resource-manager/authorization/2022-04-01/roledefinitions
//     RoleDefinitionProperties: roleName/description/type/permissions/assignableScopes
//     are settable; createdBy/createdOn/updatedBy/updatedOn are read-only (bicep
//     ReadOnly, so no ComputedFields entry is needed to strip them from the PUT body).
type RoleDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RoleDefinition)(nil)

// NewRoleDefinition returns knowledge for the roleDefinitions resource.
func NewRoleDefinition() *RoleDefinition {
	return &RoleDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/roleDefinitions",
			ApiVersions:  []string{"2022-04-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// AzureRM hardcodes the role type to "CustomRole" (the only valid value
			// for a user-managed role definition; built-in roles are not creatable).
			// It is Optional+Computed with this default so an omitted config still
			// PUTs "CustomRole", matching AzureRM.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.type", Value: "CustomRole"},
			},
			// AzureRM validates both the display name and the description with
			// StringIsNotEmpty. roleName is additionally promoted to Required by the
			// customizer.
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.roleName", MinLength: 1, Message: "role name must not be empty"},
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewRoleDefinition()) }
