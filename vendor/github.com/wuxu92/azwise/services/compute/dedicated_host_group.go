package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DedicatedHostGroup provides resource knowledge for Microsoft.Compute/hostGroups.
//
// Contributing Terraform resource: azurerm_dedicated_host_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/dedicated_host_group_resource.go
//     (schema L49-77, Create body L104-120)
//   - terraform-provider-azurerm internal/services/compute/validate/dedicated_host_group_name.go
//   - go-azure-sdk resource-manager/compute/2024-03-01/dedicatedhostgroups:
//     model_dedicatedhostgroup.go, model_dedicatedhostgroupproperties.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not body fields.
//   - `zone` (single) expands into the top-level ARM `zones` array; it is ForceNew.
type DedicatedHostGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DedicatedHostGroup)(nil)

func NewDedicatedHostGroup() *DedicatedHostGroup {
	return &DedicatedHostGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/hostGroups",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.platformFaultDomainCount"},
				{PropertyPath: "properties.supportAutomaticPlacement"}, // automatic_placement_enabled
				{PropertyPath: "zones"},                                // zone
			},
			RequiredFields: []string{
				"properties.platformFaultDomainCount",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.DedicatedHostGroupName
				{
					Regex:   `^[^_\W][\w-.]{0,78}[\w]$`,
					Message: "must be 2-80 chars, contain only word characters, hyphens and periods, and not start with an underscore",
				},
			},
			IntRules: []azwise.IntRule{
				// platform_fault_domain_count ── IntBetween(1, 3)
				{PropertyPath: "properties.platformFaultDomainCount", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(3)), Message: "must be between 1 and 3"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.supportAutomaticPlacement", Value: false},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDedicatedHostGroup()) }
