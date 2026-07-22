package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineScaleSet provides resource knowledge for
// Microsoft.Compute/virtualMachineScaleSets.
//
// This ARM type is exposed by AzureRM as THREE typed Terraform resources, all
// merged here because azwise is keyed by ARM resource type. Orchestrated is a
// mode variant (orchestrationMode = Flexible) while the Windows/Linux resources
// use Uniform mode; only knowledge that is universally true for every one of the
// three bodies is encoded below. Mode/OS-specific ForceNew, RequiredFields, and
// DefaultValues are intentionally omitted so they cannot corrupt validation for
// the other contributors.
//
// Contributing Terraform resources:
//   - azurerm_windows_virtual_machine_scale_set
//     (internal/services/compute/windows_virtual_machine_scale_set_resource.go,
//     schema L1186-1476, timeouts L50-55)
//   - azurerm_linux_virtual_machine_scale_set
//     (internal/services/compute/linux_virtual_machine_scale_set_resource.go,
//     schema L1155-1434)
//   - azurerm_orchestrated_virtual_machine_scale_set
//     (internal/services/compute/orchestrated_virtual_machine_scale_set_resource.go,
//     schema L67-360, timeouts L54-59)
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/validate/virtual_machine_name.go
//   - go-azure-sdk resource-manager/compute/2025-04-01/virtualmachinescalesets:
//     model_virtualmachinescaleset.go, model_virtualmachinescalesetproperties.go,
//     model_virtualmachinescalesetvmprofile.go, model_upgradepolicy.go,
//     model_capacityreservationprofile.go, constants.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not body fields.
//   - `sku`/`sku_name` + `instances` map to the top-level ARM `sku` object; the
//     valid SKU name set differs per resource (Uniform vs Orchestrated) so no
//     universal enum is encoded for it.
//   - single_placement_group is ForceNew+defaulted (true) for Windows/Linux but
//     Computed for Orchestrated, so neither a shared ForceNew nor default is
//     emitted for it.
//   - platform_fault_domain_count is Required only for Orchestrated, so it is
//     ForceNew (universal) but not a universal RequiredField.
type VirtualMachineScaleSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineScaleSet)(nil)

func NewVirtualMachineScaleSet() *VirtualMachineScaleSet {
	return &VirtualMachineScaleSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/virtualMachineScaleSets",
			ApiVersions:  []string{"2022-03-01", "2025-04-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			// Universally-ForceNew body properties (identical mapping across all three resources).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.proximityPlacementGroup"},                                     // proximity_placement_group_id
				{PropertyPath: "properties.zoneBalance"},                                                 // zone_balance
				{PropertyPath: "properties.platformFaultDomainCount"},                                    // platform_fault_domain_count
				{PropertyPath: "properties.virtualMachineProfile.priority"},                              // priority
				{PropertyPath: "properties.virtualMachineProfile.evictionPolicy"},                        // eviction_policy
				{PropertyPath: "properties.virtualMachineProfile.capacityReservation.capacityReservationGroup"}, // capacity_reservation_group_id
				{PropertyPath: "properties.upgradePolicy.mode"},                                          // upgrade_mode / upgrade_policy_mode
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.VirtualMachineName (1-80 chars, cannot be all numbers)
				{
					Regex:     `^[a-zA-Z0-9]([a-zA-Z0-9._-]{0,78}[a-zA-Z0-9_])?$`,
					MinLength: 1,
					MaxLength: 80,
					Message:   "must be 1-80 characters, contain only alphanumerics, dots, dashes and underscores, begin with an alphanumeric character and not consist only of numbers",
				},
				// priority → properties.virtualMachineProfile.priority ── full ARM SDK enum set
				{
					PropertyPath:  "properties.virtualMachineProfile.priority",
					AllowedValues: []string{"Low", "Regular", "Spot"},
					Message:       "must be one of Low, Regular, or Spot",
				},
				// eviction_policy → properties.virtualMachineProfile.evictionPolicy ── full ARM SDK enum set
				{
					PropertyPath:  "properties.virtualMachineProfile.evictionPolicy",
					AllowedValues: []string{"Deallocate", "Delete"},
					Message:       "must be one of Deallocate or Delete",
				},
				// upgrade_mode → properties.upgradePolicy.mode ── full ARM SDK enum set
				{
					PropertyPath:  "properties.upgradePolicy.mode",
					AllowedValues: []string{"Automatic", "Manual", "Rolling"},
					Message:       "must be one of Automatic, Manual, or Rolling",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.virtualMachineProfile.priority", Value: "Regular"}, // priority default Regular
				{PropertyPath: "properties.upgradePolicy.mode", Value: "Manual"},              // upgrade_mode default Manual
				{PropertyPath: "properties.zoneBalance", Value: false},                        // zone_balance default false
			},
			// Read-only runtime/state properties returned by GET; stripping them avoids perpetual diffs.
			ComputedFields: []string{
				"properties.uniqueId",
				"properties.timeCreated",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVirtualMachineScaleSet()) }
