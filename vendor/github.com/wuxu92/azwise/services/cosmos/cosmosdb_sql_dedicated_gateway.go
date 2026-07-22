package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlDedicatedGateway provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/services (azurerm_cosmosdb_sql_dedicated_gateway).
//
// The service name is always "SqlDedicatedGateway" and properties.serviceType is
// hardcoded to the same value by AzureRM.
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_dedicated_gateway_resource.go
//     :46-72 (schema: cosmosdb_account_id, instance_size ServiceSize enum ForceNew,
//     instance_count IntBetween(1,5)), :78-124 (Create, timeout 30m; serviceType
//     hardcoded to SqlDedicatedGateway)
//   - go-azure-sdk resource-manager/cosmosdb/2022-05-15/sqldedicatedgateway:
//     model_serviceresourcecreateupdateparameters.go,
//     model_serviceresourcecreateupdateproperties.go,
//     constants.go PossibleValuesForServiceSize
//
// Not encoded (deliberate): cosmosdb_account_id is the parent reference (URL segment).
// Read/Update timeouts are not declared in the typed SDK resource; only Create/Delete
// use 30m, so Read/Update fall back to provider defaults.
type CosmosDbSqlDedicatedGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlDedicatedGateway)(nil)

func NewCosmosDbSqlDedicatedGateway() *CosmosDbSqlDedicatedGateway {
	return &CosmosDbSqlDedicatedGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/services",
			ApiVersions:  []string{"2022-05-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.instanceSize"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.instanceSize", AllowedValues: []string{"Cosmos.D4s", "Cosmos.D8s", "Cosmos.D16s"}},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.instanceCount", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(5))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.serviceType", Value: "SqlDedicatedGateway"},
			},
			RequiredFields: []string{
				"properties.serviceType",
				"properties.instanceSize",
				"properties.instanceCount",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlDedicatedGateway()) }
