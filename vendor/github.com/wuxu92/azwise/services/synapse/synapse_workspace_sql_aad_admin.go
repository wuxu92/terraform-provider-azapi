package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseWorkspaceSqlAADAdmin provides resource knowledge for
// Microsoft.Synapse/workspaces/sqlAdministrators.
//
// Mirrors azurerm_synapse_workspace_sql_aad_admin (the workspace SQL Active
// Directory administrator, ID .../sqlAdministrators/activeDirectory).
//
// Sources:
//   - internal/services/synapse/synapse_workspace_sql_aad_admin_resource.go
//     (schema 40-64: login/object_id/tenant_id Required; object_id + tenant_id IsUUID;
//     timeouts Create/Update/Delete 30m Read 5m; create 82-89 → AadAdminProperties with
//     administratorType hardcoded "ActiveDirectory").
//   - go-azure-sdk resource-manager/synapse/2021-06-01/workspaces:
//     model_aadadminproperties.go, resourceids.go WorkspaceSqlAADAdmin
//     (segment casing "sqlAdministrators").
//
// Notes:
//   - object_id maps to properties.sid.
//   - Shares the AadAdminProperties body shape with workspaces/administrators but is a
//     distinct ARM resource type (sqlAdministrators).
type SynapseWorkspaceSqlAADAdmin struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseWorkspaceSqlAADAdmin)(nil)

func NewSynapseWorkspaceSqlAADAdmin() *SynapseWorkspaceSqlAADAdmin {
	return &SynapseWorkspaceSqlAADAdmin{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/sqlAdministrators",
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
func init() { azwise.Register(NewSynapseWorkspaceSqlAADAdmin()) }
