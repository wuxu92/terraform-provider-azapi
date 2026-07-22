package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryDataset provides resource knowledge for
// Microsoft.DataFactory/factories/datasets.
//
// The dataset body is a discriminated union keyed by the required
// `properties.type` field (SDK: datafactory.BasicDataset interface /
// Dataset.Type `json:"type"`; DatasetResource.Properties BasicDataset
// `json:"properties"`). AzureRM splits this single ARM type into 13 typed
// Terraform resources, one per discriminator value. Because each kind has a
// completely different valid shape (its own `typeProperties`), only UNIVERSAL
// knowledge is encoded here (skill "MUST NOT union"): the name validation,
// timeouts, and the two body fields common to EVERY dataset kind —
// `properties.type` and `properties.linkedServiceName`. Kind-specific
// RequiredFields / ForceNew / DefaultValues (e.g. AzureBlob path/filename,
// CosmosDb collection_name, Snowflake table/schema) are intentionally NOT
// unioned, since they would corrupt validation for the other 12 kinds.
//
// Contributing Terraform resources and their `properties.type` discriminator:
//   - data_factory_custom_dataset          -> <user-supplied>  (type set verbatim; any dataset type)
//   - data_factory_dataset_azure_blob      -> AzureBlob
//   - data_factory_dataset_azure_sql_table -> AzureSqlTable
//   - data_factory_dataset_binary          -> Binary
//   - data_factory_dataset_cosmosdb_sqlapi -> CosmosDbSqlApiCollection
//   - data_factory_dataset_delimited_text  -> DelimitedText
//   - data_factory_dataset_http            -> HttpFile
//   - data_factory_dataset_json            -> Json
//   - data_factory_dataset_mysql           -> RelationalTable
//   - data_factory_dataset_parquet         -> Parquet
//   - data_factory_dataset_postgresql      -> RelationalTable
//   - data_factory_dataset_snowflake       -> SnowflakeTable
//   - data_factory_dataset_sql_server_table -> SqlServerTable
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_dataset_binary_resource.go:34-59
//     (representative timeouts 30/5/30/30; name ForceNew + LinkedServiceDatasetName; linked_service_name Required)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_custom_dataset_resource.go:45-90,166-169
//     (custom dataset: name ForceNew; `type` Required+ForceNew written verbatim to properties.type;
//     `linked_service` Required block -> properties.linkedServiceName)
//   - terraform-provider-azurerm internal/services/datafactory/validate/linked_service_dataset_name.go:11-16
//     (LinkedServiceDatasetName — see name-regex note below)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_dataset_azure_blob_resource.go:236,
//     data_factory_dataset_azure_sql_table_resource.go:245, data_factory_dataset_binary_resource.go:312,
//     data_factory_dataset_cosmosdb_sqlapi_resource.go (CosmosDbSQLAPICollectionDataset),
//     data_factory_dataset_delimited_text_resource.go:445, data_factory_dataset_http_resource.go:235,
//     data_factory_dataset_json_resource.go:304, data_factory_dataset_mysql_resource.go (RelationalTableDataset),
//     data_factory_dataset_parquet_resource.go:364, data_factory_dataset_postgresql_resource.go (RelationalTableDataset),
//     data_factory_dataset_snowflake_resource.go:246, data_factory_dataset_sql_server_table_resource.go:218
//     (Properties struct per kind -> properties.type discriminator)
//   - terraform-provider-azurerm vendor/github.com/jackofallops/kermit/sdk/datafactory/2018-06-01/datafactory/models.go:88967-88986
//     (Dataset base: LinkedServiceName `json:"linkedServiceName,omitempty"`, Type `json:"type,omitempty"`)
//   - terraform-provider-azurerm vendor/github.com/jackofallops/kermit/sdk/datafactory/2018-06-01/datafactory/enums.go:2524,2544,2550,2560,2570,2598,2608,2644,2658,2664,2698
//     (TypeBasicDataset discriminator string literals above)
//
// Intentionally NOT encoded:
//   - Kind-specific typeProperties (path, filename, collection_name, table_name,
//     schema, compression, encoding, ...): each valid only for one discriminator
//     value, so unioning them would reject every other kind's body.
//   - data_factory_id / name: `name` is ForceNew (encoded); the parent factory is
//     an ARM ID segment, not a dataset body property, so it carries no rule.
type DataFactoryDataset struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryDataset)(nil)

// NewDataFactoryDataset returns knowledge for the
// Microsoft.DataFactory/factories/datasets resource.
func NewDataFactoryDataset() *DataFactoryDataset {
	return &DataFactoryDataset{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/datasets",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			// Universal across every dataset kind: the child resource name is
			// ForceNew in all 13 AzureRM resources.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// validate.LinkedServiceDatasetName is a *negative* whole-string check:
			// it rejects a name composed ENTIRELY of the characters -.+?/<>*%&:\ .
			// The exact complement (valid iff the name contains at least one character
			// outside that set) is expressible as an unanchored negated character
			// class — azwise StringRule.Regex uses regexp.MatchString, so no negative
			// lookahead is required. AzureRM imposes no length bound.
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // empty = validates the resource name attribute
					Regex:        `[^-.+?/<>*%&:\\]`,
					Message:      `dataset name may not consist solely of the characters - . + ? / < > * % & : \`,
				},
			},
			// Both fields are present on the Dataset base struct and required on every
			// dataset body: `properties.type` is the discriminator (Dataset.Type,
			// written by each kind's MarshalJSON) and `properties.linkedServiceName`
			// is Required in all 13 AzureRM resources (linked_service_name / the
			// linked_service block).
			RequiredFields: []string{
				"properties.type",
				"properties.linkedServiceName",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryDataset()) }
