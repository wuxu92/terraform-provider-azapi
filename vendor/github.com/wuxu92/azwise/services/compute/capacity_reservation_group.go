package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CapacityReservationGroup provides resource knowledge for
// Microsoft.Compute/capacityReservationGroups.
//
// Contributing Terraform resource: azurerm_capacity_reservation_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/capacity_reservation_group_resource.go
//     (schema L48-63, Create body L87-95)
//   - terraform-provider-azurerm internal/services/compute/validate/capacity_reservation_group_name.go
//   - go-azure-sdk resource-manager/compute/2022-03-01/capacityreservationgroups:
//     model_capacityreservationgroup.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not body fields.
//   - `zones` (multiple) is a top-level ARM array and is ForceNew.
type CapacityReservationGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CapacityReservationGroup)(nil)

func NewCapacityReservationGroup() *CapacityReservationGroup {
	return &CapacityReservationGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/capacityReservationGroups",
			ApiVersions:  []string{"2022-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "zones"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.CapacityReservationGroupName (1-64 chars)
				{
					Regex:     `^[^_\W]([\w-._]{0,62}[\w_])?$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 chars, cannot contain special characters or whitespace, or begin with '_' or end with '.' or '-'",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCapacityReservationGroup()) }
