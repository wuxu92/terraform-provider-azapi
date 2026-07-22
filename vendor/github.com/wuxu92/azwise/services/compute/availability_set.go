package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AvailabilitySet provides resource knowledge for Microsoft.Compute/availabilitySets.
//
// Contributing Terraform resource: azurerm_availability_set.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/availability_set_resource.go
//     (schema L50-101, Create body L132-152)
//   - go-azure-sdk resource-manager/compute/2024-03-01/availabilitysets:
//     model_availabilityset.go, model_availabilitysetproperties.go, model_sku.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not body fields.
//   - The `managed` bool maps one-to-one onto sku.name: managed=true → "Aligned"
//     (Managed AV set), managed=false omits the Sku (Classic). It is ForceNew and
//     defaults to true, hence the sku.name default/ForceNew below.
type AvailabilitySet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AvailabilitySet)(nil)

func NewAvailabilitySet() *AvailabilitySet {
	return &AvailabilitySet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/availabilitySets",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.platformUpdateDomainCount"},
				{PropertyPath: "properties.platformFaultDomainCount"},
				{PropertyPath: "properties.proximityPlacementGroup"},
				{PropertyPath: "sku.name"}, // managed
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringMatch (up to 80 chars)
				{
					Regex:     `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,78}[a-zA-Z0-9_])?$`,
					MinLength: 1,
					MaxLength: 80,
					Message:   "may contain only letters, numbers, periods, hyphens and underscores, up to 80 characters, and must begin with a letter or number and end with a letter, number or underscore",
				},
			},
			IntRules: []azwise.IntRule{
				// platform_update_domain_count ── IntBetween(1, 20)
				{PropertyPath: "properties.platformUpdateDomainCount", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(20)), Message: "must be between 1 and 20"},
				// platform_fault_domain_count ── IntBetween(1, 3)
				{PropertyPath: "properties.platformFaultDomainCount", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(3)), Message: "must be between 1 and 3"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.platformUpdateDomainCount", Value: float64(5)},
				{PropertyPath: "properties.platformFaultDomainCount", Value: float64(3)},
				{PropertyPath: "sku.name", Value: "Aligned"}, // managed default true
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewAvailabilitySet()) }
