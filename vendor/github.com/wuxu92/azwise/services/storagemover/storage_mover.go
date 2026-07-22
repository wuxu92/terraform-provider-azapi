package storagemover

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageMover provides resource knowledge for Microsoft.StorageMover/storageMovers.
//
// Mirrors azurerm_storage_mover.
//
// Sources:
//   - internal/services/storagemover/storage_mover_resource.go
//     (Arguments 54-75: name Required ForceNew StringIsNotEmpty; location ForceNew;
//     description Optional; tags; create 106-118 → StorageMover{Location,
//     Properties.Description}; timeouts Create 30m Read 5m Delete 30m).
//   - go-azure-sdk resource-manager/storagemover/2025-07-01/storagemovers:
//     model_storagemoverproperties.go (description; provisioningState read-only),
//     id_storagemover.go (type segment casing "storageMovers").
type StorageMover struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageMover)(nil)

// NewStorageMover returns knowledge for the storageMovers resource.
func NewStorageMover() *StorageMover {
	return &StorageMover{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageMover/storageMovers",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewStorageMover()) }
