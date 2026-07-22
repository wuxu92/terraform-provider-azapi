package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterDevBoxDefinition provides resource knowledge for
// Microsoft.DevCenter/devCenters/devBoxDefinitions.
//
// Contributing Terraform resource: azurerm_dev_center_dev_box_definition.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_dev_box_definition_resource.go
//     (schema L52-81, Create body L124-134)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/devboxdefinitions:
//     id_devcenterdevboxdefinition.go (type segment "devBoxDefinitions", parent "devCenters"),
//     model_devboxdefinitionproperties.go, model_sku.go, constants.go (HibernateSupport).
//
// Key mappings:
//   - image_reference_id → properties.imageReference.id (Required)
//   - sku_name → properties.sku.name (Required)
//   - hibernate_support_enabled (bool) → properties.hibernateSupport (Enabled/Disabled)
//
// Notes:
//   - name/location/tags are envelope-owned; dev_center_id is the parent, not a body field.
type DevCenterDevBoxDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterDevBoxDefinition)(nil)

func NewDevCenterDevBoxDefinition() *DevCenterDevBoxDefinition {
	return &DevCenterDevBoxDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/devCenters/devBoxDefinitions",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.hibernateSupport",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM defaults hibernate_support_enabled = false.
				{PropertyPath: "properties.hibernateSupport", Value: "Disabled"},
			},
			RequiredFields: []string{
				"properties.imageReference.id",
				"properties.sku.name",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterDevBoxDefinition()) }
