package databasemigration

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DatabaseMigrationService provides resource knowledge for Microsoft.DataMigration/services.
//
// Contributing Terraform resource:
//   - azurerm_database_migration_service
//
// Sources:
//   - terraform-provider-azurerm internal/services/databasemigration/database_migration_service_resource.go
//     (schema L49-83, timeouts L42-47, Create body L106-118 with Kind hardcoded "Cloud")
//   - internal/services/databasemigration/validate/service_name.go (name regex)
//   - go-azure-sdk resource-manager/datamigration/2021-06-30/serviceresource:
//     model_datamigrationservice.go, model_datamigrationserviceproperties.go, model_servicesku.go.
//
// Notes:
//   - subnet_id maps to properties.virtualSubnetId (a non-pointer string, required by ARM).
//   - sku_name maps to sku.name. AzureRM restricts it to the four cloud SKUs (no SDK enum
//     exists; the list is derived from the resourceSkus/listSkus endpoint).
//   - kind is hardcoded to "Cloud" by AzureRM; the ARM API accepts it, so it is required.
//   - name/subnet_id/sku_name are all ForceNew.
type DatabaseMigrationService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DatabaseMigrationService)(nil)

// NewDatabaseMigrationService returns knowledge for the Microsoft.DataMigration/services resource.
func NewDatabaseMigrationService() *DatabaseMigrationService {
	return &DatabaseMigrationService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataMigration/services",
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
				{PropertyPath: "properties.virtualSubnetId"},
				{PropertyPath: "sku.name"},
			},
			StringRules: []azwise.StringRule{
				// name (resource name; empty PropertyPath).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9\-_.]+$`,
					Message:      "name must start with a letter or number and can contain letters, numbers, underscores, dashes and periods",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Premium_4vCores", "Standard_1vCores", "Standard_2vCores", "Standard_4vCores"},
					Message:       "must be one of Premium_4vCores, Standard_1vCores, Standard_2vCores, or Standard_4vCores",
				},
			},
			// Server-populated, read-only ARM properties (absent from the create body).
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.publicKey",
				"properties.virtualNicId",
			},
			RequiredFields: []string{
				"properties.virtualSubnetId",
				"sku.name",
				"kind", // AzureRM hardcodes "Cloud"
			},
		},
	}
}

// Self-registers into the azwise registry.
func init() { azwise.Register(NewDatabaseMigrationService()) }
