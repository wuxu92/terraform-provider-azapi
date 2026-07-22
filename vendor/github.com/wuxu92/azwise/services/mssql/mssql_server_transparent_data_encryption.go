package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlServerTransparentDataEncryption provides resource knowledge for
// Microsoft.Sql/servers/encryptionProtector
// (azurerm_mssql_server_transparent_data_encryption).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_server_transparent_data_encryption_resource.go
//     :41-46   (timeouts: create/update/delete 30m, read 5m)
//     :54-71   (server_id ForceNew; key_vault_key_id keyvault URI; auto_rotation_enabled default false)
//     :140-190 (payload: properties.autoRotationEnabled / serverKeyName / serverKeyType)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/encryptionprotectors:
//     model_encryptionprotectorproperties.go, constants.go (ServerKeyType)
//
// NOTE: the encryption protector name is always "current". key_vault_key_id is a
// composite Key Vault key URI that AzureRM splits into a separate
// Microsoft.Sql/servers/keys (serverKeys) resource plus properties.serverKeyName and
// properties.serverKeyType — it has no single ARM body field on this resource, so the
// keyvault URI validator is non-mappable and intentionally not represented as a rule.
type MsSqlServerTransparentDataEncryption struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlServerTransparentDataEncryption)(nil)

// NewMsSqlServerTransparentDataEncryption returns knowledge for the
// servers/encryptionProtector resource.
func NewMsSqlServerTransparentDataEncryption() *MsSqlServerTransparentDataEncryption {
	return &MsSqlServerTransparentDataEncryption{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/encryptionProtector",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.serverKeyType",
					AllowedValues: []string{"AzureKeyVault", "ServiceManaged"},
					Message:       "server_key_type must be one of AzureKeyVault or ServiceManaged",
				},
			},
			RequiredFields: []string{
				"properties.serverKeyType",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoRotationEnabled", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlServerTransparentDataEncryption()) }
