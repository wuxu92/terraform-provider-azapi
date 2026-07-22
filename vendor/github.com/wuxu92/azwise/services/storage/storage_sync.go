package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageSyncService provides resource knowledge for
// Microsoft.StorageSync/storageSyncServices.
//
// In AzureRM this is the azurerm_storage_sync resource (Azure File Sync). Note
// this is a distinct ARM provider (Microsoft.StorageSync) from Microsoft.Storage,
// even though AzureRM keeps it in the storage service package.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_sync_resource.go:33-85
//     (schema), :108-114 (create payload → ARM body)
//   - .../internal/services/storage/validate/storage_sync_name.go (name regex)
//   - go-azure-sdk resource-manager/storagesync/2020-03-01/storagesyncservicesresource:
//     model_storagesyncservicecreateparametersproperties.go (ARM json path),
//     constants.go (IncomingTrafficPolicy values), id_storagesyncservice.go (type casing)
type StorageSyncService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageSyncService)(nil)

// NewStorageSyncService returns knowledge for the storageSyncServices resource.
func NewStorageSyncService() *StorageSyncService {
	return &StorageSyncService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageSync/storageSyncServices",
			ApiVersions:  []string{"2020-03-01"},
			// name is ForceNew (schema L54-59) but is an envelope field, not a body
			// property. incoming_traffic_policy is updatable in place.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: letters, numbers, spaces and .-_, not ending in '. '
					// (validate.StorageSyncName)
					Regex:   `^[0-9a-zA-Z-_. ]*[0-9a-zA-Z-_]$`,
					Message: "may contain letters, numbers, spaces and '.-_', and must not end with '.' or a space",
				},
				{
					PropertyPath: "properties.incomingTrafficPolicy",
					// storagesyncservicesresource.PossibleValuesForIncomingTrafficPolicy()
					AllowedValues: []string{"AllowAllTraffic", "AllowVirtualNetworksOnly"},
					Message:       "must be AllowAllTraffic or AllowVirtualNetworksOnly",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// incoming_traffic_policy defaults to AllowAllTraffic (schema L68).
				{PropertyPath: "properties.incomingTrafficPolicy", Value: "AllowAllTraffic"},
			},
		},
	}
}

func init() { azwise.Register(NewStorageSyncService()) }
