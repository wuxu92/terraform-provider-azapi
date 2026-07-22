package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ProximityPlacementGroup provides resource knowledge for
// Microsoft.Compute/proximityPlacementGroups.
//
// Contributing Terraform resource: azurerm_proximity_placement_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/proximity_placement_group_resource.go
//     (schema L49-87, Create body L113-129)
//   - go-azure-sdk resource-manager/compute/2022-03-01/proximityplacementgroups:
//     model_proximityplacementgroup.go, model_proximityplacementgroupproperties.go,
//     model_proximityplacementgrouppropertiesintent.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not body fields.
//   - `zone` (single) expands into the top-level ARM `zones` array; it is ForceNew.
//   - allowed_vm_sizes is conditionally ForceNew (only when transitioning from a
//     non-empty set to empty), so it is left out of the declarative ForceNew rules.
type ProximityPlacementGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ProximityPlacementGroup)(nil)

func NewProximityPlacementGroup() *ProximityPlacementGroup {
	return &ProximityPlacementGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/proximityPlacementGroups",
			ApiVersions:  []string{"2022-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "zones"}, // zone
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringIsNotEmpty
				{MinLength: 1, Message: "must not be empty"},
			},
			// zone (→ zones) requires allowed_vm_sizes (→ properties.intent.vmSizes).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"zones", "properties.intent.vmSizes"}, Message: "zone requires allowed_vm_sizes to be set"},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewProximityPlacementGroup()) }
