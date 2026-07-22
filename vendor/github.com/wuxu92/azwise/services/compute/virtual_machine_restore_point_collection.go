package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineRestorePointCollection provides resource knowledge for
// Microsoft.Compute/restorePointCollections.
//
// Contributing TF resource:
//   - azurerm_virtual_machine_restore_point_collection (internal/services/compute/virtual_machine_restore_point_collection_resource.go)
//
// Sources:
//   - virtual_machine_restore_point_collection_resource.go:46-217 (schema, timeouts, ForceNew)
//   - go-azure-sdk compute/2024-03-01/restorepointcollections/model_restorepointcollectionproperties.go
type VirtualMachineRestorePointCollection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineRestorePointCollection)(nil)

func NewVirtualMachineRestorePointCollection() *VirtualMachineRestorePointCollection {
	return &VirtualMachineRestorePointCollection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/restorePointCollections",
			ApiVersions:  []string{"2024-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.source.id"}, // source_virtual_machine_id
			},
			RequiredFields: []string{
				"properties.source.id", // source_virtual_machine_id
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineRestorePointCollection()) }
