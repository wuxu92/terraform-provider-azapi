package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageSyncGroup provides resource knowledge for
// Microsoft.StorageSync/storageSyncServices/syncGroups.
//
// In AzureRM this is the azurerm_storage_sync_group resource. The sync group has
// no settable body properties — it is created with an empty
// SyncGroupCreateParameters{} — so only name/type/timeout knowledge applies.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_sync_group_resource.go:21-53
//     (schema), :80 (create with empty SyncGroupCreateParameters{})
//   - .../internal/services/storage/validate/storage_sync_name.go (name regex)
//   - go-azure-sdk resource-manager/storagesync/2020-03-01/syncgroupresource:
//     id_syncgroup.go (type-segment casing)
type StorageSyncGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageSyncGroup)(nil)

// NewStorageSyncGroup returns knowledge for the syncGroups sub-resource.
func NewStorageSyncGroup() *StorageSyncGroup {
	return &StorageSyncGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageSync/storageSyncServices/syncGroups",
			ApiVersions:  []string{"2020-03-01"},
			// name and storage_sync_id are both ForceNew (schema L39-51) but are
			// envelope/parent references, not body properties. The resource has no
			// mutable body fields (no Update func).
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: letters, numbers, spaces and .-_, not ending in '. '
					// (validate.StorageSyncName)
					Regex:   `^[0-9a-zA-Z-_. ]*[0-9a-zA-Z-_]$`,
					Message: "may contain letters, numbers, spaces and '.-_', and must not end with '.' or a space",
				},
			},
		},
	}
}

func init() { azwise.Register(NewStorageSyncGroup()) }
