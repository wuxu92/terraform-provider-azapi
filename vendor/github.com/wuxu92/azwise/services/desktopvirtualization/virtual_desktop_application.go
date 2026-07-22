package desktopvirtualization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualDesktopApplication provides resource knowledge for
// Microsoft.DesktopVirtualization/applicationGroups/applications.
//
// Sources:
//   - internal/services/desktopvirtualization/virtual_desktop_application_resource.go
//     (schema + CRUD, lines 26-171)
//   - vendor/.../desktopvirtualization/2025-10-10/application/model_applicationproperties.go
//   - vendor/.../desktopvirtualization/2025-10-10/application/constants.go
type VirtualDesktopApplication struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualDesktopApplication)(nil)

func NewVirtualDesktopApplication() *VirtualDesktopApplication {
	return &VirtualDesktopApplication{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DesktopVirtualization/applicationGroups/applications",
			ApiVersions:  []string{"2025-10-10"},
			// name (ForceNew) and application_group_id (parent ref) are envelope
			// fields with no ARM body path.
			ForceNew:   []azwise.ForceNewRule{},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// resource name validation (PropertyPath empty = name attribute)
					Regex:   "^[-a-zA-Z0-9]{1,260}$",
					Message: "Virtual desktop application name must be 1 - 260 characters long, contain only letters, numbers and hyphens.",
				},
				{
					PropertyPath:  "properties.commandLineSetting",
					AllowedValues: []string{"Allow", "DoNotAllow", "Require"},
				},
				{PropertyPath: "properties.friendlyName", MinLength: 1, MaxLength: 64},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 512},
			},
			DefaultValues: []azwise.DefaultValue{
				// O+C: the API defaults friendlyName to the application name.
				{PropertyPath: "properties.friendlyName"},
				// O+C: the API defaults iconPath to filePath.
				{PropertyPath: "properties.iconPath"},
			},
			// commandLineSetting is non-omitempty (required) in the ARM model;
			// filePath (Terraform `path`) is required in the schema.
			RequiredFields: []string{
				"properties.commandLineSetting",
				"properties.filePath",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualDesktopApplication()) }
