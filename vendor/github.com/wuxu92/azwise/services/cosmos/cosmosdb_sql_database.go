package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlDatabase provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlDatabases (azurerm_cosmosdb_sql_database).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_database_resource.go
//     :41-46 (timeouts 30/5/30/30), :48-73 (schema: name CosmosEntityName,
//     throughput CosmosThroughput, autoscale_settings), :95-112 (expand:
//     throughput/autoscaleSettings under properties.options)
//   - AzureRM internal/services/cosmos/validate/cosmos.go
//     :22-32 CosmosEntityName (1-255), :34-50 CosmosThroughput (>=400, /100),
//     :52-72 CosmosMaxThroughput (>=1000, /1000)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_sqldatabasecreateupdateproperties.go (options/resource),
//     model_createupdateoptions.go (throughput/autoscaleSettings),
//     model_autoscalesettings.go (maxThroughput)
//
// Not encoded (deliberate):
//   - name / resource_group_name / account_name are envelope-owned (URL segments),
//     not ARM body properties; the naming rule is a StringRule with empty path.
//   - CosmosThroughput "increments of 100" and CosmosMaxThroughput "increments of
//     1000" are modulo constraints not expressible as a min/max IntRule; only the
//     minimum bound is encoded.
type CosmosDbSqlDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlDatabase)(nil)

func NewCosmosDbSqlDatabase() *CosmosDbSqlDatabase {
	return &CosmosDbSqlDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlDatabases",
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
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.options.throughput", MinValue: azwise.Ptr(int64(400))},
				{PropertyPath: "properties.options.autoscaleSettings.maxThroughput", MinValue: azwise.Ptr(int64(1000))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.options.throughput"}, // optional+computed; server decides
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlDatabase()) }
