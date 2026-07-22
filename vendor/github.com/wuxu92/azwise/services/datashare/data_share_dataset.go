package datashare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataShareDataSet provides resource knowledge for
// Microsoft.DataShare/accounts/shares/dataSets.
//
// This ARM type is a discriminated union keyed by the top-level `kind` field.
// AzureRM splits it into four typed Terraform resources, which together cover
// eight of the SDK `kind` variants — all merged into this single file:
//
//   - azurerm_data_share_dataset_blob_storage        → kind Blob / BlobFolder / Container
//   - azurerm_data_share_dataset_data_lake_gen2      → kind AdlsGen2File / AdlsGen2Folder / AdlsGen2FileSystem
//   - azurerm_data_share_dataset_kusto_cluster       → kind KustoCluster
//   - azurerm_data_share_dataset_kusto_database      → kind KustoDatabase
//
// Only universally-true knowledge is unioned. `name` and `kind` are ForceNew for
// every variant; `kind` is required for every variant; `properties.dataSetId`
// is read-only for every variant. The per-kind body properties are disjoint
// across the discriminated types (e.g. properties.filePath only exists on Blob /
// AdlsGen2File bodies, properties.kustoClusterResourceId only on KustoCluster),
// so their ForceNew rules fire only when that property is present and never
// corrupt validation for another kind. Datasets have no Update operation at all —
// every body property is immutable.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datashare
//     data_share_dataset_blob_storage_resource.go (schema 41-110; Create 137-167
//     Blob/BlobFolder/Container variants; Read 202-228)
//     data_share_dataset_data_lake_gen2_resource.go (schema 40-89; Create 121-152
//     AdlsGen2File/Folder/FileSystem variants)
//     data_share_dataset_kusto_cluster_resource.go (schema 42-73; Create 100-104)
//     data_share_dataset_kusto_database_resource.go (schema 42-73; Create 100-104)
//     timeouts 30m/5m/-/30m (Create/Read/Delete; no Update).
//   - internal/services/datashare/validate/data_set_name.go (name regex, 2-90 chars).
//   - go-azure-sdk resource-manager/datashare/2019-11-01/dataset
//     BaseDataSetImpl (Kind json:"kind" required discriminator), per-kind *Properties
//     models (Blob/BlobFolder/BlobContainer/ADLSGen2File/Folder/FileSystem/
//     KustoCluster/KustoDatabase), constants.go PossibleValuesForDataSetKind.
//
// NOTE (validators not expressible declaratively here):
//   - storage_account_id / kusto_cluster_id / kusto_database_id use resource-ID
//     validators (commonids.Validate*ID). storage_account_id is parsed and split
//     into properties.{storageAccountName,resourceGroup,subscriptionId} (one-to-many,
//     no single body path); kusto_*_id maps to properties.kusto*ResourceId as a
//     full ARM ID — a semantic AzureResourceID rule belongs in an azapin customizer,
//     not a declarative StringRule, so it is left to the generator layer.
//   - container_name uses storageValidate.StorageContainerName (3-63, no consecutive
//     hyphens, plus $root/$web/$logs special names); too complex for a safe regex —
//     left to the storage service validator rather than risk rejecting valid names.
//   - file_path/folder_path ConflictsWith is a Terraform-schema construct: in ARM the
//     two map to mutually-exclusive discriminated kinds (Blob vs BlobFolder, etc.),
//     never coexisting in one body, so no RelationalRule is emitted.
type DataShareDataSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataShareDataSet)(nil)

// NewDataShareDataSet returns knowledge for the accounts/shares/dataSets resource.
func NewDataShareDataSet() *DataShareDataSet {
	return &DataShareDataSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataShare/accounts/shares/dataSets",
			ApiVersions:  []string{"2019-11-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// Universal across every kind.
				{PropertyPath: "name"},
				{PropertyPath: "kind"},
				// Per-kind, disjoint body properties (all immutable — no Update op).
				// Blob / BlobFolder / Container:
				{PropertyPath: "properties.containerName"},
				{PropertyPath: "properties.filePath"},   // Blob, AdlsGen2File
				{PropertyPath: "properties.prefix"},      // BlobFolder
				{PropertyPath: "properties.folderPath"},  // AdlsGen2Folder
				// Blob + ADLS Gen2 storage-account addressing:
				{PropertyPath: "properties.storageAccountName"},
				{PropertyPath: "properties.resourceGroup"},
				{PropertyPath: "properties.subscriptionId"},
				{PropertyPath: "properties.fileSystem"}, // ADLS Gen2 kinds
				// Kusto kinds:
				{PropertyPath: "properties.kustoClusterResourceId"},
				{PropertyPath: "properties.kustoDatabaseResourceId"},
			},
			// `kind` is the required discriminator for every dataSet body
			// (BaseDataSetImpl.Kind, json:"kind", no omitempty). Per-kind required
			// properties are variant-specific and intentionally NOT unioned here.
			RequiredFields: []string{
				"kind",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM DataSetName():
					// numbers, letters, - and _, 2-90 characters.
					Regex:     `^[\w-]{2,90}$`,
					MinLength: 2,
					MaxLength: 90,
					Message:   "Dataset name can only contain number, letters, - and _, and must be between 2 and 90 characters long.",
				},
				{
					// Discriminator enum — full ARM SDK set (PossibleValuesForDataSetKind).
					PropertyPath: "kind",
					AllowedValues: []string{
						"AdlsGen1File", "AdlsGen1Folder",
						"AdlsGen2File", "AdlsGen2FileSystem", "AdlsGen2Folder",
						"Blob", "BlobFolder", "Container",
						"KustoCluster", "KustoDatabase",
						"SqlDBTable", "SqlDWTable",
					},
				},
				{
					// storage_account.name → storageValidate.StorageAccountName:
					// 3-24 lowercase alphanumeric. Present on Blob/ADLS Gen2 kinds.
					PropertyPath: "properties.storageAccountName",
					Regex:        `^[a-z0-9]{3,24}$`,
					MinLength:    3,
					MaxLength:    24,
					Message:      "storage account name must be 3-24 lowercase alphanumeric characters",
				},
				{
					// storage_account.subscription_id → validation.IsUUID, transferred
					// declaratively as a UUID regex. Present on Blob/ADLS Gen2 kinds.
					PropertyPath: "properties.subscriptionId",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "must be a valid UUID",
				},
			},
			// Server-populated, read-only ARM properties. dataSetId is present on
			// every kind; location/provisioningState are populated on the Kusto kinds.
			ComputedFields: []string{
				"properties.dataSetId",
				"properties.location",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewDataShareDataSet()) }
