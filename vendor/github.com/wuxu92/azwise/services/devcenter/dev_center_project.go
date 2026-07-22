package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterProject provides resource knowledge for Microsoft.DevCenter/projects.
//
// Contributing Terraform resource: azurerm_dev_center_project.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_project_resource.go
//     (schema L53-88)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/projects:
//     id_project.go (type segment "projects"), model_projectproperties.go.
//
// Key mappings:
//   - dev_center_id → properties.devCenterId (ForceNew)
//   - description → properties.description (ForceNew)
//   - maximum_dev_boxes_per_user → properties.maxDevBoxesPerUser
//   - dev_center_uri → properties.devCenterUri (read-only)
//
// Notes:
//   - name/location/resource_group_name/identity/tags are envelope-owned; not body rules.
type DevCenterProject struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterProject)(nil)

func NewDevCenterProject() *DevCenterProject {
	return &DevCenterProject{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/projects",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.devCenterId"},
				{PropertyPath: "properties.description"},
			},
			ComputedFields: []string{
				"properties.devCenterUri",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterProject()) }
