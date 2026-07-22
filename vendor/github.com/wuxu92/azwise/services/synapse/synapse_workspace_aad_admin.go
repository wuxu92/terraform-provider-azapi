package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseWorkspaceAADAdmin provides resource knowledge for
// Microsoft.Synapse/workspaces/administrators.
//
// Mirrors azurerm_synapse_workspace_aad_admin (the workspace Active Directory
// administrator, ID .../administrators/activeDirectory).
//
// Sources:
//   - internal/services/synapse/synapse_workspace_aad_admin_resource.go
//     (schema 40-64: login/object_id/tenant_id Required; object_id + tenant_id IsUUID;
//     timeouts Create/Update/Delete 30m Read 5m; create 80-87 → AadAdminProperties with
//     administratorType hardcoded "ActiveDirectory").
//   - go-azure-sdk resource-manager/synapse/2021-06-01/workspaces:
//     model_aadadminproperties.go (tenantId/login/administratorType/sid json tags),
//     resourceids.go WorkspaceAADAdmin (segment casing "administrators").
//
// Notes:
//   - object_id maps to properties.sid (not properties.objectId).
//   - IsUUID checks on object_id/tenant_id are semantic validators; represented here
//     as-is only via required-field presence — the UUID format check belongs in an
//     azapin customizer (schema/validators UUID) if attached.
type SynapseWorkspaceAADAdmin struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseWorkspaceAADAdmin)(nil)

func NewSynapseWorkspaceAADAdmin() *SynapseWorkspaceAADAdmin {
	return &SynapseWorkspaceAADAdmin{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/administrators",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.login",
				"properties.sid",
				"properties.tenantId",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.administratorType", Value: "ActiveDirectory"},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseWorkspaceAADAdmin()) }
