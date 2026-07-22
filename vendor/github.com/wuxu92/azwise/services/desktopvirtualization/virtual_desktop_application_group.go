package desktopvirtualization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualDesktopApplicationGroup provides resource knowledge for
// Microsoft.DesktopVirtualization/applicationGroups.
//
// Sources:
//   - internal/services/desktopvirtualization/virtual_desktop_application_group_resource.go
//     (schema + CRUD, lines 31-105)
//   - vendor/.../desktopvirtualization/2025-10-10/applicationgroup/model_applicationgroupproperties.go
//   - vendor/.../desktopvirtualization/2025-10-10/applicationgroup/constants.go
//
// Notes:
//   - default_desktop_display_name is not part of the applicationGroup body; it is
//     applied via a separate PATCH to the child Desktop resource — skipped here.
type VirtualDesktopApplicationGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualDesktopApplicationGroup)(nil)

func NewVirtualDesktopApplicationGroup() *VirtualDesktopApplicationGroup {
	return &VirtualDesktopApplicationGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DesktopVirtualization/applicationGroups",
			ApiVersions:  []string{"2025-10-10"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.applicationGroupType"},
				{PropertyPath: "properties.hostPoolArmPath"},
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
					PropertyPath:  "properties.applicationGroupType",
					AllowedValues: []string{"Desktop", "RemoteApp"},
				},
				{PropertyPath: "properties.friendlyName", MinLength: 1, MaxLength: 64},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 512},
			},
			// applicationGroupType and hostPoolArmPath are non-omitempty (required)
			// in the ARM ApplicationGroupProperties model.
			RequiredFields: []string{
				"properties.applicationGroupType",
				"properties.hostPoolArmPath",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualDesktopApplicationGroup()) }
