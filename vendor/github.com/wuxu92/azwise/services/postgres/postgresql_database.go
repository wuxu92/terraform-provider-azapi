package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlDatabase provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers/databases (azurerm_postgresql_database).
//
// Deprecated Single Server sub-resource; DISTINCT from
// Microsoft.DBforPostgreSQL/flexibleServers/databases.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_database_resource.go
//     :25-78   (schema: name ForceNew, server_name ForceNew, charset Required ForceNew,
//     collation Required ForceNew)
//     :35-39   (timeouts: create/delete 60m, read 5m)
//     :102-107 (create mapping to databases.DatabaseProperties)
//   - go-azure-sdk resource-manager/postgresql/2017-12-01/databases:
//     model_databaseproperties.go (properties.charset / properties.collation),
//     id_database.go:123-125 (staticServers/staticDatabases)
type PostgresqlDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlDatabase)(nil)

// NewPostgresqlDatabase returns knowledge for the servers/databases resource.
func NewPostgresqlDatabase() *PostgresqlDatabase {
	return &PostgresqlDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers/databases",
			ApiVersions:  []string{"2017-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.charset"},
				{PropertyPath: "properties.collation"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 60 * time.Minute,
			},
			RequiredFields: []string{
				"properties.charset",
				"properties.collation",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlDatabase()) }
