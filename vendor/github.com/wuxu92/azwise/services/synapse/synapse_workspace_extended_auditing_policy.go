package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseWorkspaceExtendedAuditingPolicy provides resource knowledge for
// Microsoft.Synapse/workspaces/extendedAuditingSettings.
//
// Mirrors azurerm_synapse_workspace_extended_auditing_policy (ID
// .../extendedAuditingSettings/default).
//
// Sources:
//   - internal/services/synapse/synapse_workspace_extended_auditing_policy_resource.go
//     (schema 43-82: storage_endpoint (IsURLWithHTTPS); storage_account_access_key
//     Sensitive; storage_account_access_key_is_secondary Default false;
//     retention_in_days IntBetween(0,3285) Default 0; log_monitoring_enabled Default true;
//     timeouts Create/Update/Delete 30m Read 5m; create 116-123 →
//     ExtendedServerBlobAuditingPolicyProperties, state hardcoded Enabled).
//   - go-azure-sdk resource-manager (track1) synapse model
//     ExtendedServerBlobAuditingPolicyProperties (state/storageEndpoint/
//     storageAccountAccessKey/isStorageSecondaryKeyInUse/retentionDays/
//     isAzureMonitorTargetEnabled json tags), resourceids.go
//     WorkspaceExtendedAuditingPolicy (segment casing "extendedAuditingSettings").
type SynapseWorkspaceExtendedAuditingPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseWorkspaceExtendedAuditingPolicy)(nil)

func NewSynapseWorkspaceExtendedAuditingPolicy() *SynapseWorkspaceExtendedAuditingPolicy {
	return &SynapseWorkspaceExtendedAuditingPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/extendedAuditingSettings",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.retentionDays",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(3285)),
				},
			},
			SensitiveFields: []string{
				"properties.storageAccountAccessKey",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.state", Value: "Enabled"},
				{PropertyPath: "properties.isStorageSecondaryKeyInUse", Value: false},
				{PropertyPath: "properties.retentionDays", Value: int64(0)},
				{PropertyPath: "properties.isAzureMonitorTargetEnabled", Value: true},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseWorkspaceExtendedAuditingPolicy()) }
