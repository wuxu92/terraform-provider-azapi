package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageBlobInventoryPolicy provides resource knowledge for
// Microsoft.Storage/storageAccounts/inventoryPolicies.
//
// In AzureRM this is the azurerm_storage_blob_inventory_policy resource. It is an
// ARM singleton — always PUT to inventoryPolicies/default (see the SDK
// CreateOrUpdate path below) — modelled as a sub-resource of the storage account.
//
// The rule configuration (name, storage_container_name, format, schedule, scope,
// schema_fields, filter/blob_types) lives under properties.policy.rules[*], i.e.
// array-element paths, which azwise does not represent. The nested StringInSlice
// validators (FormatCsv/Parquet, ScheduleDaily/Weekly, ObjectTypeBlob/Container,
// blob_types) are therefore intentionally not encoded here. properties.policy.enabled
// and properties.policy.type are hard-coded by AzureRM (true / "Inventory").
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_blob_inventory_policy_resource.go:29-172
//     (schema), :209-217 (create → blobinventorypolicies.CreateOrUpdate)
//   - go-azure-sdk resource-manager/storage/2025-08-01/blobinventorypolicies:
//     method_createorupdate.go (path .../inventoryPolicies/default),
//     model_blobinventorypolicyproperties.go, model_blobinventorypolicyschema.go,
//     constants.go (InventoryRuleTypeInventory)
type StorageBlobInventoryPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageBlobInventoryPolicy)(nil)

// NewStorageBlobInventoryPolicy returns knowledge for the inventoryPolicies sub-resource.
func NewStorageBlobInventoryPolicy() *StorageBlobInventoryPolicy {
	return &StorageBlobInventoryPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts/inventoryPolicies",
			ApiVersions:  []string{"2025-08-01"},
			// storage_account_id is ForceNew (schema L55) but is the envelope parent
			// reference, not a body field. The rule set (properties.policy.rules) is
			// Required, but its content is entirely array-element paths (see doc).
			RequiredFields: []string{
				"properties.policy.rules",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewStorageBlobInventoryPolicy()) }
