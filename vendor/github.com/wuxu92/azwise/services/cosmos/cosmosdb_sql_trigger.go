package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlTrigger provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/triggers
// (azurerm_cosmosdb_sql_trigger).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_trigger_resource.go
//     :29-34 (timeouts 30/5/30/30), :41-82 (schema: name CosmosEntityName,
//     body StringIsNotEmpty, operation TriggerOperation enum, type TriggerType enum),
//     :111-121 (expand under properties.resource)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_sqltriggercreateupdateproperties.go, model_sqltriggerresource.go,
//     constants.go PossibleValuesForTriggerOperation/TriggerType
//
// Not encoded (deliberate): name / container_id are envelope-owned (URL segments).
// The Create/Update operation replaces the trigger body in place, so no ForceNew
// body properties exist.
type CosmosDbSqlTrigger struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlTrigger)(nil)

func NewCosmosDbSqlTrigger() *CosmosDbSqlTrigger {
	return &CosmosDbSqlTrigger{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/triggers",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{MinLength: 1, MaxLength: 255, Message: "Cosmos DB entity name must be between 1 and 255 characters"},
				{PropertyPath: "properties.resource.triggerOperation", AllowedValues: []string{"All", "Create", "Delete", "Replace", "Update"}},
				{PropertyPath: "properties.resource.triggerType", AllowedValues: []string{"Post", "Pre"}},
			},
			RequiredFields: []string{
				"properties.resource.body",
				"properties.resource.triggerOperation",
				"properties.resource.triggerType",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlTrigger()) }
