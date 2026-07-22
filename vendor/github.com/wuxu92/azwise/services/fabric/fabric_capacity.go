package fabric

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FabricCapacity provides resource knowledge for Microsoft.Fabric/capacities
// (Microsoft Fabric capacity).
//
// Contributing Terraform resource: azurerm_fabric_capacity.
//
// Sources:
//   - terraform-provider-azurerm internal/services/fabric/fabric_capacity_resource.go
//     (schema L52-113, create L119-170 requiring administration_members, expand/flatten L278-294)
//   - go-azure-sdk resource-manager/fabric/2023-11-01/fabriccapacities:
//     model_fabriccapacity.go (top-level Sku), model_fabriccapacityproperties.go,
//     model_capacityadministration.go, model_rpsku.go, constants.go
//     (RpSkuTier / ProvisioningState / ResourceState).
//
// Notes:
//   - name, location & resource_group_name are envelope-owned (all ForceNew); none map
//     to a body property, so ForceNew is empty here.
//   - sku is a top-level ARM object (sku.name / sku.tier), not under properties.
//   - administration_members is Optional in the schema but the create func rejects an
//     empty list, so properties.administration.members is treated as required.
type FabricCapacity struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FabricCapacity)(nil)

func NewFabricCapacity() *FabricCapacity {
	return &FabricCapacity{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Fabric/capacities",
			ApiVersions:  []string{"2023-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringMatch
				{
					Regex:     `^[a-z]([a-z\d]{2,62})$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 characters, contain only lowercase letters and numbers, and start with a lowercase letter",
				},
				// sku { name } → sku.name
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"F2", "F4", "F8", "F16", "F32", "F64", "F128", "F256", "F512", "F1024", "F2048"},
				},
				// sku { tier } → sku.tier
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Fabric"},
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.state",
			},
			// sku and administration_members are required for creation.
			RequiredFields: []string{
				"sku.name",
				"sku.tier",
				"properties.administration.members",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFabricCapacity()) }
