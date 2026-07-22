package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DatabaseExtendedAuditingPolicy provides resource knowledge for
// Microsoft.Sql/servers/databases/extendedAuditingSettings.
//
// Contributing Terraform resource: azurerm_mssql_database_extended_auditing_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_database_extended_auditing_policy_resource.go
//     (schema L42-107, Create body params L138-161)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/blobauditing:
//     model_extendeddatabaseblobauditingpolicyproperties.go, constants.go (BlobAuditingPolicyState)
//
// Notes:
//   - database_id is the parent reference (envelope); not emitted as a body rule.
//   - The ARM resource is a singleton child named "default" under the database.
//   - blob_storage_endpoint (IsURLWithHTTPS) and storage_account_access_key
//     (StringIsNotEmpty) are semantic/non-enumerable validators; no declarative
//     StringRule is emitted for them.
//   - enabled maps to properties.state (Enabled/Disabled); AzureRM defaults it to
//     enabled=true, i.e. properties.state = "Enabled".
type DatabaseExtendedAuditingPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DatabaseExtendedAuditingPolicy)(nil)

// NewDatabaseExtendedAuditingPolicy returns knowledge for the extendedAuditingSettings resource.
func NewDatabaseExtendedAuditingPolicy() *DatabaseExtendedAuditingPolicy {
	return &DatabaseExtendedAuditingPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/databases/extendedAuditingSettings",
			ApiVersions:  []string{"2023-08-01-preview"},
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
					Message:       "state must be Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.retentionDays", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(3285))},
			},
			SensitiveFields: []string{
				"properties.storageAccountAccessKey",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.state", Value: "Enabled"},
				{PropertyPath: "properties.isStorageSecondaryKeyInUse", Value: false},
				{PropertyPath: "properties.retentionDays", Value: 0},
				{PropertyPath: "properties.isAzureMonitorTargetEnabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewDatabaseExtendedAuditingPolicy()) }
