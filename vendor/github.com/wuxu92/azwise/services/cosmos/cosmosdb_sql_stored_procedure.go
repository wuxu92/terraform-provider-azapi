package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlStoredProcedure provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/storedProcedures
// (azurerm_cosmosdb_sql_stored_procedure).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_stored_procedure_resource.go
//     :35-40 (timeouts 30/5/30/30), :42-78 (schema: name StringIsNotEmpty,
//     body StringIsNotEmpty; account_name/container_name/database_name envelope),
//     :100-108 (expand under properties.resource)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_sqlstoredprocedurecreateupdateproperties.go, model_sqlstoredprocedureresource.go
//
// Not encoded (deliberate): name / account_name / container_name / database_name are
// envelope-owned (URL segments). The Create/Update operation replaces the body in
// place, so no ForceNew body properties exist.
type CosmosDbSqlStoredProcedure struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlStoredProcedure)(nil)

func NewCosmosDbSqlStoredProcedure() *CosmosDbSqlStoredProcedure {
	return &CosmosDbSqlStoredProcedure{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/storedProcedures",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{MinLength: 1, Message: "name must not be empty"},
			},
			RequiredFields: []string{
				"properties.resource.body",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlStoredProcedure()) }
