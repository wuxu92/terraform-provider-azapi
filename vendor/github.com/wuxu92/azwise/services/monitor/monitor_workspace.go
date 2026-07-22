package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Workspace provides resource knowledge for Microsoft.Monitor/accounts
// (the Azure Monitor Workspace).
//
// Mirrors azurerm_monitor_workspace.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_workspace_resource.go
//     Arguments (48-69): name Required+ForceNew (StringIsNotEmpty); resource_group_name +
//     location ForceNew (commonschema); public_network_access_enabled Optional Default true →
//     properties.publicNetworkAccess (Enabled/Disabled); tags. Attributes (71-86):
//     query_endpoint, default_data_collection_endpoint_id, default_data_collection_rule_id
//     Computed. Create/Update/Delete 30m, Read 5m.
//   - go-azure-sdk resource-manager/monitor/2023-04-03/azuremonitorworkspaces:
//     AzureMonitorWorkspace (model_azuremonitorworkspace.go) fields accountId/
//     defaultIngestionSettings/metrics/privateEndpointConnections/provisioningState
//     read-only; PublicNetworkAccess{Disabled,Enabled}; id_account.go type
//     Microsoft.Monitor/accounts.
type Workspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Workspace)(nil)

// NewWorkspace returns knowledge for the Monitor accounts resource.
func NewWorkspace() *Workspace {
	return &Workspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Monitor/accounts",
			ApiVersions:  []string{"2023-04-03"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "name must not be empty",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "publicNetworkAccess must be one of Disabled, Enabled",
				},
			},
			// public_network_access_enabled defaults true → publicNetworkAccess "Enabled".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.accountId",
				"properties.defaultIngestionSettings",
				"properties.metrics",
				"properties.privateEndpointConnections",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewWorkspace()) }
