package storagemover

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageMoverProject provides resource knowledge for
// Microsoft.StorageMover/storageMovers/projects.
//
// Mirrors azurerm_storage_mover_project.
//
// Sources:
//   - internal/services/storagemover/storage_mover_project_resource.go
//     (Arguments 53-78: name Required ForceNew StringMatch ^[0-9a-zA-Z][-_0-9a-zA-Z]{0,63}$;
//     storage_mover_id ForceNew parent ref; description Optional; create 112-118 →
//     Project{Properties.Description}; timeouts Create 30m).
//   - go-azure-sdk resource-manager/storagemover/2025-07-01/projects:
//     model_projectproperties.go (description; provisioningState read-only),
//     id_project.go (type segment casing "storageMovers/projects").
type StorageMoverProject struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageMoverProject)(nil)

// NewStorageMoverProject returns knowledge for the storageMovers/projects resource.
func NewStorageMoverProject() *StorageMoverProject {
	return &StorageMoverProject{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageMover/storageMovers/projects",
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
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringMatch.
					Regex:     `^[0-9a-zA-Z][-_0-9a-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, begin with a letter or number, and contain only letters, numbers, dashes and underscores",
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
func init() { azwise.Register(NewStorageMoverProject()) }
