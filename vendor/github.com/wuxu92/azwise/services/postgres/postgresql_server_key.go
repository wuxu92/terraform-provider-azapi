package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlServerKey provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers/keys (azurerm_postgresql_server_key).
//
// Deprecated Single Server sub-resource (Transparent Data Encryption key). It is a
// singleton whose ARM name AzureRM derives from the Key Vault key URI.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_server_key_resource.go
//     :26-61   (schema: server_id ForceNew, key_vault_key_id Required — a versioned
//     Key Vault key URI)
//     :38-43   (timeouts: create/update/delete 60m, read 5m)
//     :133-138 (create mapping; serverKeyType hardcoded AzureKeyVault, uri<-key_vault_key_id)
//   - go-azure-sdk resource-manager/postgresql/2020-01-01/serverkeys:
//     model_serverkeyproperties.go (properties.serverKeyType / uri / creationDate),
//     constants.go:14-16 (ServerKeyType), id_key.go:123-125 (staticServers/staticKeys)
//
// Not encoded (deliberate):
//   - key_vault_key_id uses keyvault.ValidateNestedItemID (a versioned Key Vault key URI
//     check), a semantic validator that belongs in an azapin customizer, not declarative.
type PostgresqlServerKey struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlServerKey)(nil)

// NewPostgresqlServerKey returns knowledge for the servers/keys resource.
func NewPostgresqlServerKey() *PostgresqlServerKey {
	return &PostgresqlServerKey{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers/keys",
			ApiVersions:  []string{"2020-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ComputedFields: []string{
				"properties.creationDate",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.serverKeyType",
					AllowedValues: []string{"AzureKeyVault"},
					Message:       "serverKeyType must be AzureKeyVault",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.serverKeyType", Value: "AzureKeyVault"},
			},
			RequiredFields: []string{
				"properties.serverKeyType",
				"properties.uri",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlServerKey()) }
