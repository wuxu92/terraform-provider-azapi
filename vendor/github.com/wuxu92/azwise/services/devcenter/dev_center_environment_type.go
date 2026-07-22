package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterEnvironmentType provides resource knowledge for
// Microsoft.DevCenter/devCenters/environmentTypes.
//
// Contributing Terraform resource: azurerm_dev_center_environment_type.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_environment_type_resource.go
//     (schema L46-59, Create body L97-99)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/environmenttypes:
//     id_devcenterenvironmenttype.go (type segment "environmentTypes", parent "devCenters"),
//     model_environmenttype.go.
//
// Notes:
//   - name is envelope-owned; dev_center_id is the parent, not a body field.
//   - The only settable payload is tags (envelope); no body property rules apply.
type DevCenterEnvironmentType struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterEnvironmentType)(nil)

func NewDevCenterEnvironmentType() *DevCenterEnvironmentType {
	return &DevCenterEnvironmentType{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/devCenters/environmentTypes",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterEnvironmentType()) }
