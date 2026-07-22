package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseWorkspaceKey provides resource knowledge for
// Microsoft.Synapse/workspaces/keys.
//
// Mirrors azurerm_synapse_workspace_key.
//
// Sources:
//   - internal/services/synapse/synapse_workspace_key_resource.go
//     (schema 45-67: customer_managed_key_name + active Required;
//     customer_managed_key_versionless_id Optional (KeyVault nested key ID validator);
//     timeouts Create/Update/Delete 30m Read 5m; create 96-103 → KeyProperties).
//   - go-azure-sdk resource-manager/synapse/2021-06-01/workspaces:
//     model_workspacekeydetails.go / KeyProperties (isActiveCMK/keyVaultUrl json tags),
//     resourceids.go WorkspaceKeys (segment casing "keys").
//
// Notes:
//   - customer_managed_key_name maps to the resource name segment (keyName), not a body
//     property.
//   - customer_managed_key_versionless_id maps to properties.keyVaultUrl.
//   - active maps to properties.isActiveCMK.
type SynapseWorkspaceKey struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseWorkspaceKey)(nil)

func NewSynapseWorkspaceKey() *SynapseWorkspaceKey {
	return &SynapseWorkspaceKey{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/keys",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.isActiveCMK",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseWorkspaceKey()) }
