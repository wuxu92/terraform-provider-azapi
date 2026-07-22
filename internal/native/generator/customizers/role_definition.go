package customizers

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/schema/validators"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// customizeRoleDefinition applies role-definition-specific schema rules that neither
// the bicep type graph nor the azwise overlay can express:
//   - The ARM resource name is the role definition GUID (azurerm's role_definition_id,
//     validated with IsUUID). It is not part of the body type graph — it maps to the
//     ARM ID — so its UUID constraint is attached to the envelope name attribute. The
//     azwise layer cannot carry it: ApplyAzwise skips empty-path (name) StringRules.
//   - properties.roleName (the display name, azurerm's Required `name`) is bicep
//     Optional; a role definition must have a name, so it is promoted to Required.
//     The azwise not-empty length rule and the description not-empty rule ride the
//     overlay automatically; only the Required flag needs promoting here (azwise
//     RequiredFields is planner metadata, not a schema flag).
//
// The role type default ("CustomRole"), the not-empty rules, and the dynamic
// assignable_scopes default (a runtime hook) are handled by the azwise overlay and
// role_definition_hooks.go respectively; see github.com/wuxu92/azwise/role_definition.go.
func customizeRoleDefinition(def *typegraph.ResourceDefinition) {
	def.SetNameValidators(typegraph.Validator(validators.UUID))
	def.SetParent("scope_id", "The scope ID where this role definition is defined.")

	def.Required("properties.roleName")
}
