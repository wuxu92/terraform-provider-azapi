package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryLinkedService provides resource knowledge for
// Microsoft.DataFactory/factories/linkedservices.
//
// The linked-service body is a discriminated union keyed by the required
// `properties.type` field (SDK: linkedservices.LinkedService interface /
// BaseLinkedServiceImpl.Type). AzureRM splits this single ARM type into 23 typed
// Terraform resources, one per discriminator value. Because each kind has a
// completely different valid shape, only UNIVERSAL knowledge is encoded here
// (skill "MUST NOT union"): the name validation, timeouts, and the single body
// field common to every linked-service kind — `properties.type`. Kind-specific
// RequiredFields / ForceNew / DefaultValues are intentionally NOT unioned, since
// they would corrupt validation for the other 22 kinds.
//
// Contributing Terraform resources and their `properties.type` discriminator:
//   - data_factory_linked_custom_service        -> <user-supplied>  (any linked-service type; `type` set verbatim)
//   - data_factory_linked_service_web           -> Web
//   - data_factory_linked_service_key_vault     -> AzureKeyVault
//   - data_factory_linked_service_kusto         -> AzureDataExplorer
//   - data_factory_linked_service_mysql         -> MySql
//   - data_factory_linked_service_odata         -> OData
//   - data_factory_linked_service_odbc          -> Odbc
//   - data_factory_linked_service_postgresql    -> PostgreSql
//   - data_factory_linked_service_sftp          -> Sftp
//   - data_factory_linked_service_snowflake     -> Snowflake
//   - data_factory_linked_service_sql_managed_instance -> AzureSqlMI
//   - data_factory_linked_service_sql_server    -> SqlServer
//   - data_factory_linked_service_synapse       -> AzureSqlDW
//   - data_factory_linked_service_azure_blob_storage    -> AzureBlobStorage
//   - data_factory_linked_service_azure_databricks      -> AzureDatabricks
//   - data_factory_linked_service_azure_file_storage    -> AzureFileStorage
//   - data_factory_linked_service_azure_function        -> AzureFunction
//   - data_factory_linked_service_azure_search          -> AzureSearch
//   - data_factory_linked_service_azure_sql_database    -> AzureSqlDatabase
//   - data_factory_linked_service_azure_table_storage   -> AzureTableStorage
//   - data_factory_linked_service_cosmosdb              -> CosmosDb
//   - data_factory_linked_service_cosmosdb_mongoapi     -> CosmosDbMongoDbApi
//   - data_factory_linked_service_data_lake_storage_gen2 -> AzureBlobFS
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_linked_service_web_resource.go:35-54
//     (representative timeouts 30/5/30/30; name ForceNew + LinkedServiceDatasetName; data_factory_id ForceNew)
//   - terraform-provider-azurerm internal/services/datafactory/validate/linked_service_dataset_name.go:11-18
//     (LinkedServiceDatasetName — negative "not allowed characters" check; see note below)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_linked_custom_service_resource.go:58-61,153-154
//     (custom service: `type` Required+ForceNew, written verbatim to properties.type)
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/linkedservices/model_linkedservice.go:12-25
//     (BaseLinkedServiceImpl: Type string `json:"type"` is required on every kind)
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/linkedservices/model_linkedservice.go:50-1044
//     (UnmarshalLinkedServiceImplementation: `properties.type` discriminator values above)
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/linkedservices/model_linkedserviceresource.go:13-20
//     (LinkedServiceResource envelope; Properties LinkedService `json:"properties"`)
//
// Intentionally NOT encoded:
//   - Name validation regex: validate.LinkedServiceDatasetName is a *negative* check —
//     it errors only when the name consists ENTIRELY of `-.+?/<>*%&:\` characters.
//     RE2 (Go regexp, used by azwise StringRule.Regex) has no negative lookahead, so
//     this "must NOT match" rule cannot be expressed as a positive regex. Omitted
//     rather than emit a wrong constraint; there is no length limit in AzureRM either.
//   - Kind-specific ForceNew (e.g. data_factory_id parent reference is an ID segment,
//     not a body property) and every per-kind RequiredField / DefaultValue: these vary
//     across the 23 discriminated bodies and would corrupt validation for other kinds.
type DataFactoryLinkedService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryLinkedService)(nil)

// NewDataFactoryLinkedService returns knowledge for the
// Microsoft.DataFactory/factories/linkedservices resource.
func NewDataFactoryLinkedService() *DataFactoryLinkedService {
	return &DataFactoryLinkedService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/linkedservices",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			// Universal across every linked-service kind: the child resource name is
			// ForceNew in all 23 AzureRM resources.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// `properties.type` is the discriminator and is required on every linked
			// service body (BaseLinkedServiceImpl.Type `json:"type"`, no omitempty).
			RequiredFields: []string{
				"properties.type",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryLinkedService()) }
