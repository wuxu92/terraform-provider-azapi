package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageSyncServerEndpoint provides resource knowledge for
// Microsoft.StorageSync/storageSyncServices/syncGroups/serverEndpoints.
//
// In AzureRM this is the azurerm_storage_sync_server_endpoint resource.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_sync_server_endpoint_resource.go:61-122
//     (schema), :160-178 (create payload → ARM body)
//   - go-azure-sdk resource-manager/storagesync/2020-03-01/serverendpointresource:
//     model_serverendpointcreateparametersproperties.go (ARM json paths),
//     constants.go (InitialDownloadPolicy / LocalCacheMode values),
//     id_serverendpoint.go (type-segment casing)
type StorageSyncServerEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageSyncServerEndpoint)(nil)

// NewStorageSyncServerEndpoint returns knowledge for the serverEndpoints sub-resource.
func NewStorageSyncServerEndpoint() *StorageSyncServerEndpoint {
	return &StorageSyncServerEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageSync/storageSyncServices/syncGroups/serverEndpoints",
			ApiVersions:  []string{"2020-03-01"},
			// server_local_path, registered_server_id and initial_download_policy are
			// ForceNew (schema L82-114) and map to body fields. name and
			// storage_sync_group_id are ForceNew envelope/parent references.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.serverLocalPath"},
				{PropertyPath: "properties.serverResourceId"},
				{PropertyPath: "properties.initialDownloadPolicy"},
			},
			// server_local_path and registered_server_id are Required (schema L82-87).
			RequiredFields: []string{
				"properties.serverLocalPath",
				"properties.serverResourceId",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.initialDownloadPolicy",
					// serverendpointresource.PossibleValuesForInitialDownloadPolicy()
					AllowedValues: []string{
						"AvoidTieredFiles", "NamespaceOnly", "NamespaceThenModifiedFiles",
					},
					Message: "must be AvoidTieredFiles, NamespaceOnly, or NamespaceThenModifiedFiles",
				},
				{
					PropertyPath: "properties.localCacheMode",
					// serverendpointresource.PossibleValuesForLocalCacheMode()
					AllowedValues: []string{
						"DownloadNewAndModifiedFiles", "UpdateLocallyCachedFiles",
					},
					Message: "must be DownloadNewAndModifiedFiles or UpdateLocallyCachedFiles",
				},
			},
			IntRules: []azwise.IntRule{
				{
					// volume_free_space_percent: IntBetween(1, 100)
					PropertyPath: "properties.volumeFreeSpacePercent",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](100),
				},
				{
					// tier_files_older_than_days: IntBetween(1, math.MaxInt32)
					PropertyPath: "properties.tierFilesOlderThanDays",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](2147483647),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// cloud_tiering_enabled defaults to false → cloudTiering "off" (schema L89-93).
				{PropertyPath: "properties.cloudTiering", Value: "off"},
				// volume_free_space_percent defaults to 20 (schema L95-100).
				{PropertyPath: "properties.volumeFreeSpacePercent", Value: float64(20)},
				// initial_download_policy defaults to NamespaceThenModifiedFiles (schema L108-114).
				{PropertyPath: "properties.initialDownloadPolicy", Value: "NamespaceThenModifiedFiles"},
				// local_cache_mode defaults to UpdateLocallyCachedFiles (schema L116-121).
				{PropertyPath: "properties.localCacheMode", Value: "UpdateLocallyCachedFiles"},
			},
		},
	}
}

func init() { azwise.Register(NewStorageSyncServerEndpoint()) }
