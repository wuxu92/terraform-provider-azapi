package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterProjectEnvironmentType provides resource knowledge for
// Microsoft.DevCenter/projects/environmentTypes.
//
// Contributing Terraform resource: azurerm_dev_center_project_environment_type.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_project_environment_type_resource.go
//     (schema L60-115, Create body L157-174)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/environmenttypes:
//     id_environmenttype.go (type segment "environmentTypes", parent "projects"),
//     model_projectenvironmenttypeproperties.go.
//
// Key mappings:
//   - deployment_target_id → properties.deploymentTargetId (Required)
//   - creator_role_assignment_roles → properties.creatorRoleAssignment.roles (map[roleId]role)
//   - user_role_assignment → properties.userRoleAssignments (map[userId]{roles})
//
// Notes:
//   - name/location/identity/tags are envelope-owned; dev_center_project_id is the parent.
//   - creator_role_assignment_roles, user_role_assignment user_id and roles use
//     validation.IsUUID. These validate map keys / entries inside userRoleAssignments and
//     creatorRoleAssignment.roles, which have no single ARM body scalar path — they are
//     map-key/array semantic checks and are intentionally not emitted as declarative rules.
type DevCenterProjectEnvironmentType struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterProjectEnvironmentType)(nil)

func NewDevCenterProjectEnvironmentType() *DevCenterProjectEnvironmentType {
	return &DevCenterProjectEnvironmentType{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/projects/environmentTypes",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.deploymentTargetId",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterProjectEnvironmentType()) }
