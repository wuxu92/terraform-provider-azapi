package desktopvirtualization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualDesktopWorkspace provides resource knowledge for
// Microsoft.DesktopVirtualization/workspaces.
//
// Sources:
//   - internal/services/desktopvirtualization/virtual_desktop_workspace_resource.go
//     (schema + CRUD, lines 29-137)
//   - vendor/.../desktopvirtualization/2025-10-10/workspace/model_workspaceproperties.go
//   - vendor/.../desktopvirtualization/2025-10-10/workspace/constants.go
//
// Notes:
//   - The Terraform bool public_network_access_enabled maps to the ARM enum
//     properties.publicNetworkAccess (Enabled/Disabled). AzAPI users send the raw
//     ARM enum value directly.
type VirtualDesktopWorkspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualDesktopWorkspace)(nil)

func NewVirtualDesktopWorkspace() *VirtualDesktopWorkspace {
	return &VirtualDesktopWorkspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DesktopVirtualization/workspaces",
			ApiVersions:  []string{"2025-10-10"},
			// name (ForceNew) is an envelope field with no ARM body path.
			ForceNew:   []azwise.ForceNewRule{},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.friendlyName", MinLength: 1, MaxLength: 64},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 512},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualDesktopWorkspace()) }
