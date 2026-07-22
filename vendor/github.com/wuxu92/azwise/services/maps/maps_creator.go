package maps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MapsCreator provides resource knowledge for Microsoft.Maps/accounts/creators.
//
// Mirrors azurerm_maps_creator (deprecated in AzureRM, removed in v5.0, but the ARM
// resource type still exists and is exposed through AzAPI). The only body property
// is properties.storageUnits, an int in [1, 100].
//
// Sources:
//   - terraform-provider-azurerm internal/services/maps/maps_creator_resource.go
//   - Schema(): name (Required, ForceNew, envelope), maps_account_id (Required, ForceNew, parent),
//     location (ForceNew), storage_units (Required, IntBetween(1,100) -> properties.storageUnits)
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/maps/2023-06-01/creators model_creatorproperties.go:
//     CreatorProperties.StorageUnits int64 (Required)
type MapsCreator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MapsCreator)(nil)

// NewMapsCreator returns knowledge for the Maps creators resource.
func NewMapsCreator() *MapsCreator {
	return &MapsCreator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Maps/accounts/creators",
			ApiVersions:  []string{"2023-06-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name and maps_account_id are
				// envelope/parent path.
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.storageUnits",
			},
			IntRules: []azwise.IntRule{
				// storage_units: validation.IntBetween(1, 100).
				{
					PropertyPath: "properties.storageUnits",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "storageUnits must be between 1 and 100",
				},
			},
		},
	}
}

func init() { azwise.Register(NewMapsCreator()) }
