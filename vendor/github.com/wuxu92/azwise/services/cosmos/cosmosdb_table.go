package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbTable provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/tables (azurerm_cosmosdb_table).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_table_resource.go
//     :41-46 (timeouts 30/5/30/30), :48-73 (schema: name CosmosEntityName,
//     throughput CosmosThroughput, autoscale_settings), :95-110 (expand)
//   - AzureRM internal/services/cosmos/validate/cosmos.go CosmosEntityName/Throughput/MaxThroughput
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_tablecreateupdateproperties.go, model_tableresource.go,
//     model_createupdateoptions.go
//
// Not encoded (deliberate): name/resource_group_name/account_name are envelope-owned;
// throughput/maxThroughput modulo (100/1000) constraints are not expressible as IntRule.
type CosmosDbTable struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbTable)(nil)

func NewCosmosDbTable() *CosmosDbTable {
	return &CosmosDbTable{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/tables",
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

func init() { azwise.Register(NewCosmosDbTable()) }
