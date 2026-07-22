package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoClusterDatabaseDataConnection provides resource knowledge for
// Microsoft.Kusto/clusters/databases/dataConnections.
//
// The ARM body is a DISCRIMINATED ROOT keyed by `kind` (CosmosDb / EventGrid /
// EventHub / IotHub). Four AzureRM TF resources map onto this single ARM type:
//   - azurerm_kusto_eventgrid_data_connection  (kind = EventGrid)
//   - azurerm_kusto_eventhub_data_connection   (kind = EventHub)
//   - azurerm_kusto_iothub_data_connection     (kind = IotHub)
//   - azurerm_kusto_cosmosdb_data_connection   (kind = CosmosDb)
//
// Per the azwise merge policy, only knowledge UNIVERSAL to every kind's body is
// encoded here. Kind-specific ForceNew (consumerGroup, eventHubResourceId,
// cosmosDbAccountResourceId, ...), kind-specific RequiredFields, and kind-specific
// DefaultValues (databaseRouting=Single, compression=None, blobStorageEventType=
// Microsoft.Storage.BlobCreated) are intentionally NOT unioned — injecting them
// would corrupt validation for the other kinds. Value constraints (enums) on a
// sub-object that only one kind sets ARE included, because they only fire when that
// path is present in the submitted body.
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_eventgrid_data_connection_resource.go:31-185
//     (EventGrid schema: name/location ForceNew, blob_storage_event_type default+enum,
//     data_format enum, database_routing_type default+enum+ForceNew, timeouts)
//   - terraform-provider-azurerm internal/services/kusto/kusto_eventhub_data_connection_resource.go:29-147
//     (EventHub schema: compression default+enum+ForceNew, data_format enum,
//     database_routing_type default+enum+ForceNew)
//   - terraform-provider-azurerm internal/services/kusto/kusto_iothub_data_connection_resource.go:29-135
//     (IotHub schema: data_format enum+ForceNew, database_routing_type default+enum+ForceNew)
//   - terraform-provider-azurerm internal/services/kusto/kusto_cosmosdb_data_connection_resource.go:41-87
//     (CosmosDb schema: name/location ForceNew, no databaseRouting/dataFormat/compression)
//   - terraform-provider-azurerm internal/services/kusto/validate/name.go:11-27
//     (DataConnectionName: charclass regex + max length 40; used by all four `name` fields)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/dataconnections:
//     model_dataconnection.go:18-24 (base: kind, location, name),
//     model_eventgridconnectionproperties.go:6-20, model_eventhubconnectionproperties.go:12-25,
//     model_iothubconnectionproperties.go:12-23, model_cosmosdbdataconnectionproperties.go:12-22
//     (per-kind properties.* body paths),
//     id_dataconnection.go:115-135 (ARM path casing: clusters/databases/dataConnections),
//     constants.go:14-17 (BlobStorageEventType), 55-58 (Compression: GZip/None),
//     96-101 (DataConnectionKind), 181-184 (DatabaseRouting: Multi/Single),
//     222-239/305-322/388-405 (EventGrid/EventHub/IotHub DataFormat — identical value sets)
//
// Not encoded (deliberate):
//   - consumer_group, storage_account_id, eventhub_id, iothub_id,
//     shared_access_policy_name, cosmosdb_container_id, managed_identity_id,
//     table_name, retrieval_start_date, event_system_properties: kind-specific body
//     fields (Required/ForceNew varies by kind), so not universal. The semantic ID
//     validators (ValidateStorageAccountID, ValidateEventhubID, IotHubID,
//     ValidateContainerID, azure.ValidateResourceID) belong in an azapin customizer,
//     not as declarative rules.
//   - database_routing_type / compression / blob_storage_event_type DEFAULTS: each
//     applies to only one (or three) kind(s); a body-level DefaultValue would inject
//     the field into a kind that has no such property, so defaults are omitted.
//   - kind-specific ForceNew and RequiredFields: omitted for the same reason.
//   - provisioningState is server-computed read-only in every kind (see ComputedFields).
type KustoClusterDatabaseDataConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoClusterDatabaseDataConnection)(nil)

// NewKustoClusterDatabaseDataConnection returns knowledge for the
// clusters/databases/dataConnections resource (all four kinds).
func NewKustoClusterDatabaseDataConnection() *KustoClusterDatabaseDataConnection {
	dataFormats := []string{
		"APACHEAVRO", "AVRO", "CSV", "JSON", "MULTIJSON", "ORC", "PARQUET", "PSV",
		"RAW", "SCSV", "SINGLEJSON", "SOHSV", "TSV", "TSVE", "TXT", "W3CLOGFILE",
	}
	return &KustoClusterDatabaseDataConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/databases/dataConnections",
			ApiVersions:  []string{"2025-02-14"},
			// location (commonschema.Location()) is Required + ForceNew for every kind
			// — the only universal ForceNew across the discriminated union.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name (PropertyPath == ""): DataConnectionName validator
				// (universal — all four kinds use it).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9\s.-]+$`,
					MaxLength:    40,
					Message:      "must only contain letters, digits, whitespaces, dashes and dots, and be at most 40 characters",
				},
				// databaseRouting: EventGrid/EventHub/IotHub only. Fires only when set.
				{
					PropertyPath:  "properties.databaseRouting",
					AllowedValues: []string{"Multi", "Single"},
					Message:       "must be Multi or Single",
				},
				// dataFormat: EventGrid/EventHub/IotHub — identical value sets. Fires only when set.
				{
					PropertyPath:  "properties.dataFormat",
					AllowedValues: dataFormats,
					Message:       "must be a valid Kusto data connection data format",
				},
				// compression: EventHub only. Fires only when set.
				{
					PropertyPath:  "properties.compression",
					AllowedValues: []string{"GZip", "None"},
					Message:       "must be GZip or None",
				},
				// blobStorageEventType: EventGrid only. Fires only when set.
				{
					PropertyPath:  "properties.blobStorageEventType",
					AllowedValues: []string{"Microsoft.Storage.BlobCreated", "Microsoft.Storage.BlobRenamed"},
					Message:       "must be Microsoft.Storage.BlobCreated or Microsoft.Storage.BlobRenamed",
				},
			},
			// provisioningState is read-only in every kind's properties.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewKustoClusterDatabaseDataConnection()) }
