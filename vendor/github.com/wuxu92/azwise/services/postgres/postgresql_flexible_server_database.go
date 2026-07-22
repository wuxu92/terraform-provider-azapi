package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServerDatabase provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers/databases
// (azurerm_postgresql_flexible_server_database).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_database_resource.go
//     :24-72   (schema: name FlexibleServerDatabaseName len 1..63 regex ForceNew,
//     charset ForceNew default UTF8, collation ForceNew default en_US.utf8)
//     :34-38   (timeouts: create/delete 30m, read 5m)
//     :103-108 (create mapping to databases.DatabaseProperties)
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/databases:
//     model_databaseproperties.go (properties.charset / properties.collation),
//     id_database.go:125 (staticDatabases)
type PostgresqlFlexibleServerDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServerDatabase)(nil)

// NewPostgresqlFlexibleServerDatabase returns knowledge for the
// flexibleServers/databases resource.
func NewPostgresqlFlexibleServerDatabase() *PostgresqlFlexibleServerDatabase {
	return &PostgresqlFlexibleServerDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers/databases",
			ApiVersions:  []string{"2025-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.charset"},
				{PropertyPath: "properties.collation"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    63,
					Regex:        `^[a-zA-Z-_][a-zA-Z0-9-_]*$`,
					Message:      "database name must begin with a letter and contain only letters, numbers, '-' and '_' (1-63 characters)",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.charset", Value: "UTF8"},
				{PropertyPath: "properties.collation", Value: "en_US.utf8"},
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServerDatabase()) }
