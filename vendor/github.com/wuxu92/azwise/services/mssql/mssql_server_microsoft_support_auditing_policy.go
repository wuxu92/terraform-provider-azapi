package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlServerMicrosoftSupportAuditingPolicy provides resource knowledge for
// Microsoft.Sql/servers/devOpsAuditingSettings
// (azurerm_mssql_server_microsoft_support_auditing_policy).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_server_microsoft_support_auditing_policy_resource.go
//     :35-40  (timeouts: create/update/delete 30m, read 5m)
//     :43-80  (server_id ForceNew; enabled default true → properties.state;
//     storage_account_access_key Sensitive; log_monitoring_enabled default true;
//     storage_account_subscription_id Sensitive IsUUID)
//     :112-134 (create mapping: properties.isAzureMonitorTargetEnabled / storageEndpoint /
//     state / storageAccountSubscriptionId / storageAccountAccessKey)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/serverdevopsaudit:
//     model_serverdevopsauditsettingsproperties.go, constants.go (BlobAuditingPolicyState)
//
// NOTE: the ARM id parser resolves this to servers/devOpsAuditingSettings, a DIFFERENT
// ARM type from servers/extendedAuditingSettings (azurerm_mssql_server_extended_auditing_policy).
// The assignment's assumption that both map to extendedAuditingSettings is contradicted
// by the SDK, so they are emitted as separate files rather than merged.
type MsSqlServerMicrosoftSupportAuditingPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlServerMicrosoftSupportAuditingPolicy)(nil)

// NewMsSqlServerMicrosoftSupportAuditingPolicy returns knowledge for the
// servers/devOpsAuditingSettings resource.
func NewMsSqlServerMicrosoftSupportAuditingPolicy() *MsSqlServerMicrosoftSupportAuditingPolicy {
	return &MsSqlServerMicrosoftSupportAuditingPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/devOpsAuditingSettings",
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
			SensitiveFields: []string{
				"properties.storageAccountAccessKey",
				"properties.storageAccountSubscriptionId",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.state", Value: "Enabled"},
				{PropertyPath: "properties.isAzureMonitorTargetEnabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlServerMicrosoftSupportAuditingPolicy()) }
