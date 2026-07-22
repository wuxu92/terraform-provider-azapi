package desktopvirtualization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualDesktopHostPool provides resource knowledge for
// Microsoft.DesktopVirtualization/hostPools.
//
// Sources:
//   - internal/services/desktopvirtualization/virtual_desktop_host_pool_resource.go
//     (schema + CRUD, lines 28-215)
//   - vendor/.../desktopvirtualization/2025-10-10/hostpool/model_hostpoolproperties.go
//   - vendor/.../desktopvirtualization/2025-10-10/hostpool/constants.go
//
// Notes:
//   - vm_template uses validation.StringIsJSON (semantic JSON check) — not
//     expressible as a declarative StringRule; skipped.
//   - scheduled_agent_updates maps to properties.agentUpdate; its schedule entries
//     (day_of_week / hour_of_day) live under properties.agentUpdate.maintenanceWindows[*],
//     an array-element path unsupported by azwise — skipped.
type VirtualDesktopHostPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualDesktopHostPool)(nil)

func NewVirtualDesktopHostPool() *VirtualDesktopHostPool {
	return &VirtualDesktopHostPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DesktopVirtualization/hostPools",
			ApiVersions:  []string{"2025-10-10"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.hostPoolType"},
				{PropertyPath: "properties.personalDesktopAssignmentType"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.hostPoolType",
					AllowedValues: []string{"BYODesktop", "Personal", "Pooled"},
				},
				{
					PropertyPath:  "properties.loadBalancerType",
					AllowedValues: []string{"BreadthFirst", "DepthFirst", "MultiplePersistent", "Persistent"},
				},
				{
					PropertyPath:  "properties.personalDesktopAssignmentType",
					AllowedValues: []string{"Automatic", "Direct"},
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled", "EnabledForClientsOnly", "EnabledForSessionHostsOnly"},
				},
				{
					PropertyPath:  "properties.preferredAppGroupType",
					AllowedValues: []string{"Desktop", "None", "RailApplications"},
				},
				{PropertyPath: "properties.friendlyName", MinLength: 1, MaxLength: 64},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 512},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.maxSessionLimit", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(999999))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.validationEnvironment", Value: false},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.maxSessionLimit", Value: float64(999999)},
				{PropertyPath: "properties.startVMOnConnect", Value: false},
				{PropertyPath: "properties.preferredAppGroupType", Value: "Desktop"},
			},
			// hostPoolType, loadBalancerType and preferredAppGroupType are non-omitempty
			// (required) in the ARM HostPoolProperties model.
			RequiredFields: []string{
				"properties.hostPoolType",
				"properties.loadBalancerType",
				"properties.preferredAppGroupType",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualDesktopHostPool()) }
