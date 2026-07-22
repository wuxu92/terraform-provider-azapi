package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineScaleSetStandbyPool provides resource knowledge for
// Microsoft.StandbyPool/standbyVirtualMachinePools.
//
// Contributing Terraform resource: azurerm_virtual_machine_scale_set_standby_pool.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/virtual_machine_scale_set_standby_pool_resource.go
//     (schema L57-108, CustomizeDiff L114-130, Create body L157-165)
//   - go-azure-sdk resource-manager/standbypool/2025-03-01/standbyvirtualmachinepools:
//     model_standbyvirtualmachinepoolresourceproperties.go,
//     model_standbyvirtualmachinepoolelasticityprofile.go, constants.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not body fields.
//   - The CustomizeDiff constraint min_ready_capacity <= max_ready_capacity is a
//     cross-field comparison over two numeric values that azwise cannot express
//     declaratively (RelationalRule only checks presence), so it is not encoded.
type VirtualMachineScaleSetStandbyPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineScaleSetStandbyPool)(nil)

func NewVirtualMachineScaleSetStandbyPool() *VirtualMachineScaleSetStandbyPool {
	return &VirtualMachineScaleSetStandbyPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StandbyPool/standbyVirtualMachinePools",
			ApiVersions:  []string{"2025-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.attachedVirtualMachineScaleSetId",
				"properties.elasticityProfile",
				"properties.virtualMachineState",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringMatch
				{
					Regex:     `^[a-zA-Z0-9-]{3,24}$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "must be 3-24 characters and may contain only letters, numbers and hyphens",
				},
				// virtual_machine_state → properties.virtualMachineState ── full ARM SDK enum set
				{
					PropertyPath:  "properties.virtualMachineState",
					AllowedValues: []string{"Deallocated", "Hibernated", "Running"},
					Message:       "must be one of Deallocated, Hibernated, or Running",
				},
			},
			IntRules: []azwise.IntRule{
				// elasticity_profile.max_ready_capacity ── IntBetween(0, 2000)
				{PropertyPath: "properties.elasticityProfile.maxReadyCapacity", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(2000)), Message: "must be between 0 and 2000"},
				// elasticity_profile.min_ready_capacity ── IntBetween(0, 2000)
				{PropertyPath: "properties.elasticityProfile.minReadyCapacity", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(2000)), Message: "must be between 0 and 2000"},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVirtualMachineScaleSetStandbyPool()) }
