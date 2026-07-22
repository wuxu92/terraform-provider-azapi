package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CapacityReservation provides resource knowledge for
// Microsoft.Compute/capacityReservationGroups/capacityReservations.
//
// Contributing Terraform resource: azurerm_capacity_reservation.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/capacity_reservation_resource.go
//     (schema L51-91, Create body L129-138, expandCapacityReservationSku L238-244)
//   - terraform-provider-azurerm internal/services/compute/validate/capacity_reservation_name.go
//   - go-azure-sdk resource-manager/compute/2022-03-01/capacityreservations:
//     model_capacityreservation.go, model_sku.go.
//
// Notes:
//   - name is envelope-owned; capacity_reservation_group_id is the parent
//     reference (envelope); location is inherited from the parent group.
//   - sku maps to the top-level ARM `sku` object (sku.name, sku.capacity), not
//     a properties.* field.
//   - `zone` (single) expands into the top-level ARM `zones` array; it is ForceNew.
type CapacityReservation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CapacityReservation)(nil)

func NewCapacityReservation() *CapacityReservation {
	return &CapacityReservation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/capacityReservationGroups/capacityReservations",
			ApiVersions:  []string{"2022-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"}, // sku.name
				{PropertyPath: "zones"},    // zone
			},
			RequiredFields: []string{
				"sku.name",
				"sku.capacity",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.CapacityReservationName (1-80 chars)
				{
					Regex:     `^[^_\W]([\w-._]{0,78}[\w_])?$`,
					MinLength: 1,
					MaxLength: 80,
					Message:   "must be 1-80 chars, cannot contain special characters or whitespace, or begin with '_' or end with '.' or '-'",
				},
				// sku.name ── StringIsNotEmpty
				{PropertyPath: "sku.name", MinLength: 1, Message: "must not be empty"},
			},
			IntRules: []azwise.IntRule{
				// sku.capacity ── IntAtLeast(0)
				{PropertyPath: "sku.capacity", MinValue: azwise.Ptr(int64(0)), Message: "must be at least 0"},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCapacityReservation()) }
