package authorization

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/hashicorp/go-uuid"
)

// RoleDefinitionCfg carries the Terraform address metadata and the fixed role
// definition GUID for azapi_authorization_role_definition acceptance scenarios.
// Construct it with NewRoleDefinitionCfg, then wrap it in a scenario type
// (RoleDefinitionCfg_Basic, _Complete, _Complete_update) when applying. Every
// scenario scopes the role at the subscription (parent_id =
// "/subscriptions/{{.SubscriptionID}}"), the azapi analogue of azurerm's
// data.azurerm_subscription scope, so no base resource is needed.
//
// The GUID (the ARM resource name, azurerm's role_definition_id) is generated once
// per Cfg and held so every scenario renders the SAME name: the name is ForceNew, so a
// stable GUID keeps Basic -> Complete -> Complete_update an in-place update chain on a
// single role. A fresh GUID per run avoids colliding with a role leaked by an earlier
// run at the same subscription scope.
type RoleDefinitionCfg struct {
	config.ResourceConfigBase
	name string
}

// NewRoleDefinitionCfg builds a role-definition config scoped at the subscription. The
// label is optional — omit it for the single-instance default ("test"), or pass an
// explicit label when a scope holds more than one. The resource type is read from the
// AuthorizationRoleDefinition descriptor.
func NewRoleDefinitionCfg(label ...string) RoleDefinitionCfg {
	guid, err := uuid.GenerateUUID()
	if err != nil {
		panic(fmt.Sprintf("authorization: generating role definition GUID: %v", err))
	}
	return RoleDefinitionCfg{
		ResourceConfigBase: config.NewResourceConfigBase(AuthorizationRoleDefinition.Name, label...),
		name:               guid,
	}
}

// RoleDefinitionCfg_Basic is a minimal custom role: a display name and a single
// wildcard permission. It sets no type (defaulting to "CustomRole" via the azwise
// overlay) and no assignable_scopes (the BeforeCreate hook defaults it to the role's
// own scope, the subscription), so it exercises both defaults round-tripping without
// drift. data_actions / not_data_actions are left to compute.
type RoleDefinitionCfg_Basic RoleDefinitionCfg

func (r RoleDefinitionCfg_Basic) Config() string {
	return RoleDefinitionCfg(r).config(`
  properties = {
    role_name = "accazapirole-{{.RandomInteger}}"
    permissions = [
      {
        actions     = ["*"]
        not_actions = []
      }
    ]
  }`)
}

// RoleDefinitionCfg_Complete provisions the full custom-role surface: a description,
// all four permission action lists spelled explicitly, and an explicit
// assignable_scopes. Applied as an in-place update on the role created by Basic (same
// GUID), it proves the added description / widened permissions / explicit scopes
// round-trip (Update -> Read -> empty plan).
type RoleDefinitionCfg_Complete RoleDefinitionCfg

func (r RoleDefinitionCfg_Complete) Config() string {
	return RoleDefinitionCfg(r).config(`
  properties = {
    role_name   = "accazapirole-{{.RandomInteger}}"
    description = "azapi native acceptance role"
    permissions = [
      {
        actions          = ["*"]
        data_actions     = ["Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read"]
        not_actions      = ["Microsoft.Authorization/*/read"]
        not_data_actions = []
      }
    ]
    assignable_scopes = ["/subscriptions/{{.SubscriptionID}}"]
  }`)
}

// RoleDefinitionCfg_Complete_update flips the description and narrows the permission
// set from Complete — clearing data_actions and widening not_actions, each spelled
// explicitly so no Optional+Computed list is left ambiguous — so applying Complete then
// Complete_update proves an in-place update of the description and permissions round-
// trips (Update -> Read -> empty plan).
type RoleDefinitionCfg_Complete_update RoleDefinitionCfg

func (r RoleDefinitionCfg_Complete_update) Config() string {
	return RoleDefinitionCfg(r).config(`
  properties = {
    role_name   = "accazapirole-{{.RandomInteger}}"
    description = "azapi native acceptance role updated"
    permissions = [
      {
        actions          = ["*"]
        data_actions     = []
        not_actions      = ["Microsoft.Authorization/*/read", "Microsoft.Authorization/*/write"]
        not_data_actions = []
      }
    ]
    assignable_scopes = ["/subscriptions/{{.SubscriptionID}}"]
  }`)
}

func (r RoleDefinitionCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		Name:       r.name,
		ParentAttr: "parent_id",
		ParentRef:  `"/subscriptions/{{.SubscriptionID}}"`,
		Body:       body,
	})
}
