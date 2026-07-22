package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseSqlPoolExtendedAuditingPolicy provides resource knowledge for
// Microsoft.Synapse/workspaces/sqlPools/extendedAuditingSettings.
//
// Mirrors azurerm_synapse_sql_pool_extended_auditing_policy (ID
// .../sqlPools/{pool}/extendedAuditingSettings/default).
//
// Sources:
//   - internal/services/synapse/synapse_sql_pool_extended_auditing_policy_resource.go
//     (schema 42-81: storage_endpoint (IsURLWithHTTPS); storage_account_access_key
//     Sensitive; storage_account_access_key_is_secondary Default false;
//     retention_in_days IntBetween(0,3285) Default 0; log_monitoring_enabled Default true;
//     timeouts Create/Update/Delete 30m Read 5m).
//   - go-azure-sdk resource-manager (track1) synapse model
//     ExtendedServerBlobAuditingPolicyProperties (state/storageEndpoint/
//     storageAccountAccessKey/isStorageSecondaryKeyInUse/retentionDays/
//     isAzureMonitorTargetEnabled json tags), resourceids.go
//     SqlPoolExtendedAuditingPolicy (segment casing "extendedAuditingSettings").
type SynapseSqlPoolExtendedAuditingPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseSqlPoolExtendedAuditingPolicy)(nil)

func NewSynapseSqlPoolExtendedAuditingPolicy() *SynapseSqlPoolExtendedAuditingPolicy {
	return &SynapseSqlPoolExtendedAuditingPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/sqlPools/extendedAuditingSettings",
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
func init() { azwise.Register(NewSynapseSqlPoolExtendedAuditingPolicy()) }
