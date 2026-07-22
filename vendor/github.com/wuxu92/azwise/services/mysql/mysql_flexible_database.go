package mysql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MySQLFlexibleDatabase provides resource knowledge for Microsoft.DBforMySQL/flexibleServers/databases.
//
// Contributing Terraform resource: azurerm_mysql_flexible_database.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mysql/mysql_flexible_database_resource.go
//     (schema L45-73, Create body L97-102, timeouts L35-39)
//   - go-azure-sdk resource-manager/mysql/2023-12-30/databases:
//     model_database.go, model_databaseproperties.go, id_database.go
//
// Notes:
//   - name/server_name/resource_group_name are envelope / parent-reference fields; not emitted as body rules.
//   - charset (properties.charset) and collation (properties.collation) are Required + ForceNew.
type MySQLFlexibleDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MySQLFlexibleDatabase)(nil)

// NewMySQLFlexibleDatabase returns knowledge for the flexibleServers/databases resource.
func NewMySQLFlexibleDatabase() *MySQLFlexibleDatabase {
	return &MySQLFlexibleDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforMySQL/flexibleServers/databases",
			ApiVersions:  []string{"2023-12-30"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.charset"},
				{PropertyPath: "properties.collation"},
			},
			RequiredFields: []string{
				"properties.charset",
				"properties.collation",
			},
		},
	}
}

func init() { azwise.Register(NewMySQLFlexibleDatabase()) }
