package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlServerExtendedAuditingPolicy provides resource knowledge for
// Microsoft.Sql/servers/extendedAuditingSettings
// (azurerm_mssql_server_extended_auditing_policy).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_server_extended_auditing_policy_resource.go
//     :37-42  (timeouts: create/update/delete 30m, read 5m)
//     :44-56  (server_id ForceNew; enabled default true → properties.state)
//     :64-95  (storage_account_access_key Sensitive; retention_in_days IntBetween(0,3285);
//     log_monitoring_enabled default true; storage_account_subscription_id Sensitive IsUUID)
//     :165-200 (create mapping: properties.storageEndpoint / isStorageSecondaryKeyInUse /
//     retentionDays / isAzureMonitorTargetEnabled / state / storageAccountSubscriptionId /
//     storageAccountAccessKey / predicateExpression / auditActionsAndGroups)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/blobauditing:
//     model_extendedserverblobauditingpolicyproperties.go, constants.go
//     (BlobAuditingPolicyState)
//
// NOTE: this ARM type is distinct from Microsoft.Sql/servers/devOpsAuditingSettings
// (azurerm_mssql_server_microsoft_support_auditing_policy) — verified via the SDK id
// parsers (extendedAuditingSettings vs devOpsAuditingSettings) — so the two are NOT
// merged despite the shared "auditing" naming.
type MsSqlServerExtendedAuditingPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlServerExtendedAuditingPolicy)(nil)

// NewMsSqlServerExtendedAuditingPolicy returns knowledge for the
// servers/extendedAuditingSettings resource.
func NewMsSqlServerExtendedAuditingPolicy() *MsSqlServerExtendedAuditingPolicy {
	return &MsSqlServerExtendedAuditingPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/extendedAuditingSettings",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.storageEndpoint",
					Regex:        `^https://`,
					Message:      "blob_storage_endpoint must be a valid HTTPS URL",
				},
				{
					PropertyPath: "properties.storageAccountSubscriptionId",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "storage_account_subscription_id must be a valid UUID",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.retentionDays",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(3285)),
					Message:      "retention_in_days must be between 0 and 3285",
				},
			},
			SensitiveFields: []string{
				"properties.storageAccountAccessKey",
				"properties.storageAccountSubscriptionId",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.state", Value: "Enabled"},
				{PropertyPath: "properties.retentionDays", Value: float64(0)},
				{PropertyPath: "properties.isAzureMonitorTargetEnabled", Value: true},
				{PropertyPath: "properties.isStorageSecondaryKeyInUse", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlServerExtendedAuditingPolicy()) }
