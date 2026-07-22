package databasemigration

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DatabaseMigrationProject provides resource knowledge for Microsoft.DataMigration/services/projects.
//
// Contributing Terraform resource:
//   - azurerm_database_migration_project
//
// Sources:
//   - terraform-provider-azurerm internal/services/databasemigration/database_migration_project_resource.go
//     (schema L47-94, timeouts L40-45, Create body L124-131)
//   - internal/services/databasemigration/validate/project_name.go (name regex)
//   - go-azure-sdk resource-manager/datamigration/2021-06-30/projectresource:
//     model_project.go, model_projectproperties.go, constants.go
//     (ProjectSourcePlatform L105-111, ProjectTargetPlatform L155-162).
//
// Notes:
//   - service_name is the parent DMS resource name (envelope/ID), not an ARM body property.
//   - source_platform/target_platform map to properties.sourcePlatform / properties.targetPlatform
//     (non-pointer enums, required by the ARM API).
//   - name/service_name/source_platform/target_platform are all ForceNew.
type DatabaseMigrationProject struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DatabaseMigrationProject)(nil)

// NewDatabaseMigrationProject returns knowledge for the Microsoft.DataMigration/services/projects resource.
func NewDatabaseMigrationProject() *DatabaseMigrationProject {
	return &DatabaseMigrationProject{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataMigration/services/projects",
			ApiVersions:  []string{"2021-06-30"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.sourcePlatform"},
				{PropertyPath: "properties.targetPlatform"},
			},
			StringRules: []azwise.StringRule{
				// name (resource name; empty PropertyPath).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9\-_.]+$`,
					Message:      "name must start with a letter or number and can contain letters, numbers, underscores, dashes and periods",
				},
				{
					PropertyPath:  "properties.sourcePlatform",
					AllowedValues: []string{"MongoDb", "MySQL", "PostgreSql", "SQL", "Unknown"},
					Message:       "must be one of MongoDb, MySQL, PostgreSql, SQL, or Unknown",
				},
				{
					PropertyPath:  "properties.targetPlatform",
					AllowedValues: []string{"AzureDbForMySql", "AzureDbForPostgreSql", "MongoDb", "SQLDB", "SQLMI", "Unknown"},
					Message:       "must be one of AzureDbForMySql, AzureDbForPostgreSql, MongoDb, SQLDB, SQLMI, or Unknown",
				},
			},
			// Server-populated, read-only ARM properties (absent from the create body).
			ComputedFields: []string{
				"properties.creationTime",
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.sourcePlatform",
				"properties.targetPlatform",
			},
		},
	}
}

// Self-registers into the azwise registry.
func init() { azwise.Register(NewDatabaseMigrationProject()) }
