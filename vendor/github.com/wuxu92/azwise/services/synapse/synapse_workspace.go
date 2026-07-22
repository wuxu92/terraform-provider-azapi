package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseWorkspace provides resource knowledge for Microsoft.Synapse/workspaces.
//
// Mirrors azurerm_synapse_workspace.
//
// Sources:
//   - internal/services/synapse/synapse_workspace_resource.go
//     (schema 60-273: name ForceNew (WorkspaceName validator);
//     storage_data_lake_gen2_filesystem_id/sql_administrator_login/compute_subnet_id/
//     data_exfiltration_protection_enabled/managed_virtual_network_enabled ForceNew;
//     public_network_access_enabled Default true; azuread_authentication_only Default false;
//     sql_administrator_login_password Sensitive; timeouts Create/Update/Delete 30m Read 5m;
//     create 314-361 expand → WorkspaceProperties).
//   - internal/services/synapse/validate/workspace_name.go
//     (regex ^[a-z0-9]([a-z0-9-]{0,48}[a-z0-9])?$, 1-50 chars, no "-ondemand").
//   - internal/services/synapse/validate/sql_administrator_login_name.go.
//   - go-azure-sdk resource-manager/synapse/2021-06-01/workspaces:
//     model_workspaceproperties.go (json tags), constants.go
//     (WorkspacePublicNetworkAccess Enabled/Disabled), id_workspace.go
//     (type segment casing "workspaces", provider "Microsoft.Synapse").
//
// Notes:
//   - customer_managed_key expands to properties.encryption (one-to-many:
//     keyName/keyVaultUrl/userAssignedIdentity), not properties.customerManagedKey.
//   - azure_devops_repo / github_repo both expand into the single ARM object
//     properties.workspaceRepositoryConfiguration (discriminated by hostName), so the
//     AzureRM ConflictsWith between them is not expressible as an ARM-path relational
//     rule (both sides resolve to the same path) — omitted.
//   - sql_identity_control_enabled is applied via the separate
//     managedIdentitySqlControlSettings sub-API, not the workspace body — omitted.
type SynapseWorkspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseWorkspace)(nil)

func NewSynapseWorkspace() *SynapseWorkspace {
	return &SynapseWorkspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.defaultDataLakeStorage"},
				{PropertyPath: "properties.sqlAdministratorLogin"},
				{PropertyPath: "properties.virtualNetworkProfile.computeSubnetId"},
				{PropertyPath: "properties.managedVirtualNetworkSettings.preventDataExfiltration"},
				{PropertyPath: "properties.managedVirtualNetwork"},
			},
			StringRules: []azwise.StringRule{
				{
					// validate.WorkspaceName
					Regex:     `^[a-z0-9]([a-z0-9-]{0,48}[a-z0-9])?$`,
					MinLength: 1,
					MaxLength: 50,
					Message:   "must be 1-50 chars, start/end with a lowercase letter or number, contain only lowercase letters, numbers or hyphens, and not contain '-ondemand'",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
			},
			SensitiveFields: []string{
				"properties.sqlAdministratorLoginPassword",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.azureADOnlyAuthentication", Value: false},
			},
			RequiredFields: []string{
				"properties.defaultDataLakeStorage",
			},
			// Server-populated, read-only ARM properties (response-only, absent from
			// the create body semantics).
			ComputedFields: []string{
				"properties.connectivityEndpoints",
				"properties.provisioningState",
				"properties.adlaResourceId",
				"properties.workspaceUID",
				"properties.extraProperties",
				"properties.privateEndpointConnections",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseWorkspace()) }
