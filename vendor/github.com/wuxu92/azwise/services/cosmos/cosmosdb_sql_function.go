package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlFunction provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/userDefinedFunctions
// (azurerm_cosmosdb_sql_function).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_function_resource.go
//     :28-33 (timeouts 30/5/30/30), :40-59 (schema: name, container_id,
//     body StringIsNotEmpty), :86-94 (expand under properties.resource)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_sqluserdefinedfunctioncreateupdateproperties.go,
//     model_sqluserdefinedfunctionresource.go
//
// Not encoded (deliberate): name / container_id are envelope-owned (URL segments).
// The Create/Update operation replaces the body in place, so no ForceNew body
// properties exist. The name has no ValidateFunc in AzureRM.
type CosmosDbSqlFunction struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlFunction)(nil)

func NewCosmosDbSqlFunction() *CosmosDbSqlFunction {
	return &CosmosDbSqlFunction{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/userDefinedFunctions",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.resource.body",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlFunction()) }
