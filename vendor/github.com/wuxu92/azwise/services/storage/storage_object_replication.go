package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageObjectReplication provides resource knowledge for
// Microsoft.Storage/storageAccounts/objectReplicationPolicies.
//
// In AzureRM this is the azurerm_storage_object_replication resource. It creates
// a policy in both the source and destination storage accounts, but each is a
// single objectReplicationPolicies/{id} ARM sub-resource with the same body,
// hence one knowledge file.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_object_replication_resource.go:47-119
//     (schema: source/destination account, rules set, metrics_enabled)
//   - .../storage_object_replication_resource.go:164-173 (create payload → ARM body)
//   - go-azure-sdk resource-manager/storage/2025-08-01/objectreplicationpolicyoperationgroup:
//     model_objectreplicationpolicyproperties.go, model_objectreplicationpolicypropertiesmetrics.go,
//     id_objectreplicationpolicy.go (type-segment casing)
type StorageObjectReplication struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageObjectReplication)(nil)

// NewStorageObjectReplication returns knowledge for the objectReplicationPolicies sub-resource.
func NewStorageObjectReplication() *StorageObjectReplication {
	return &StorageObjectReplication{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts/objectReplicationPolicies",
			ApiVersions:  []string{"2025-08-01"},
			// source_storage_account_id / destination_storage_account_id are ForceNew
			// (schema L48-60) and map to the body fields properties.sourceAccount /
			// properties.destinationAccount (create payload L166-167).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sourceAccount"},
				{PropertyPath: "properties.destinationAccount"},
			},
			// sourceAccount, destinationAccount and rules are all Required in the ARM
			// body (rules is Required in schema L62-64).
			RequiredFields: []string{
				"properties.sourceAccount",
				"properties.destinationAccount",
				"properties.rules",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				// metrics_enabled defaults to false (schema L104-108).
				{PropertyPath: "properties.metrics.enabled", Value: false},
			},
			// Per-rule config (source_container_name, destination_container_name,
			// copy_blobs_created_after, filter_out_blobs_with_prefix) lives in
			// properties.rules[*] array elements — azwise does not represent
			// array-element paths, so those StringInSlice/container-name validators
			// are not encoded here.
		},
	}
}

func init() { azwise.Register(NewStorageObjectReplication()) }
