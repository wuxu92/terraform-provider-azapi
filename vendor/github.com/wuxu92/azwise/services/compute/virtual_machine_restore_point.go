package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineRestorePoint provides resource knowledge for
// Microsoft.Compute/restorePointCollections/restorePoints.
//
// Contributing TF resource:
//   - azurerm_virtual_machine_restore_point (internal/services/compute/virtual_machine_restore_point_resource.go)
//
// Sources:
//   - virtual_machine_restore_point_resource.go:44-139 (schema, timeouts, ForceNew)
//   - go-azure-sdk compute/2024-03-01/restorepoints/model_restorepointproperties.go + constants.go
type VirtualMachineRestorePoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineRestorePoint)(nil)

func NewVirtualMachineRestorePoint() *VirtualMachineRestorePoint {
	return &VirtualMachineRestorePoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/restorePointCollections/restorePoints",
			ApiVersions:  []string{"2024-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.consistencyMode"}, // crash_consistency_mode_enabled
				{PropertyPath: "properties.excludeDisks"},    // excluded_disks
			},
			StringRules: []azwise.StringRule{
				// consistency mode (ConsistencyModeTypes). AzureRM only wires the bool
				// crash_consistency_mode_enabled to CrashConsistent, but AzAPI users send
				// the raw enum, so the full SDK set is allowed.
				{
					PropertyPath:  "properties.consistencyMode",
					AllowedValues: []string{"ApplicationConsistent", "CrashConsistent", "FileSystemConsistent"},
				},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineRestorePoint()) }
