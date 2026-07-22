package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageSyncCloudEndpoint provides resource knowledge for
// Microsoft.StorageSync/storageSyncServices/syncGroups/cloudEndpoints.
//
// In AzureRM this is the azurerm_storage_sync_cloud_endpoint resource. Every
// schema field is ForceNew (the endpoint has no Update func).
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_sync_cloud_endpoint_resource.go:28-83
//     (schema), :110-121 (create payload → ARM body)
//   - go-azure-sdk resource-manager/storagesync/2020-03-01/cloudendpointresource:
//     model_cloudendpointcreateparametersproperties.go (ARM json paths),
//     id_cloudendpoint.go (type-segment casing)
type StorageSyncCloudEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageSyncCloudEndpoint)(nil)

// NewStorageSyncCloudEndpoint returns knowledge for the cloudEndpoints sub-resource.
func NewStorageSyncCloudEndpoint() *StorageSyncCloudEndpoint {
	return &StorageSyncCloudEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageSync/storageSyncServices/syncGroups/cloudEndpoints",
			ApiVersions:  []string{"2020-03-01"},
			// file_share_name, storage_account_id and storage_account_tenant_id are
			// all ForceNew (schema L61-81) and map to body fields. name and
			// storage_sync_group_id are ForceNew envelope/parent references.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.azureFileShareName"},
				{PropertyPath: "properties.storageAccountResourceId"},
				{PropertyPath: "properties.storageAccountTenantId"},
			},
			// file_share_name and storage_account_id are Required (schema L61-73).
			RequiredFields: []string{
				"properties.azureFileShareName",
				"properties.storageAccountResourceId",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 45 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 45 * time.Minute,
			},
			// storage_account_tenant_id is validated as a UUID in AzureRM
			// (validation.IsUUID). That is a generic semantic validator; in azapin it
			// belongs on properties.storageAccountTenantId via
			// typegraph.Validator(validators.UUID) in the resource customizer, not as a
			// declarative StringRule. Noted here for the customizer author.
		},
	}
}

func init() { azwise.Register(NewStorageSyncCloudEndpoint()) }
