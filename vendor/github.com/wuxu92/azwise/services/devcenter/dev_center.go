package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenter provides resource knowledge for Microsoft.DevCenter/devCenters.
//
// Contributing Terraform resource: azurerm_dev_center.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_resource.go
//     (schema L51-77, map L212-258)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/devcenters:
//     id_devcenter.go (type segment "devCenters"), model_devcenterproperties.go,
//     model_devcenterprojectcatalogsettings.go, constants.go (CatalogItemSyncEnableStatus).
//
// Key mappings:
//   - project_catalog_item_sync_enabled (bool) →
//     properties.projectCatalogSettings.catalogItemSyncEnableStatus (Enabled/Disabled)
//   - dev_center_uri → properties.devCenterUri (read-only)
//
// Notes:
//   - name/location/resource_group_name/identity/tags are envelope-owned; not body rules.
type DevCenter struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenter)(nil)

func NewDevCenter() *DevCenter {
	return &DevCenter{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/devCenters",
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
					PropertyPath:  "properties.projectCatalogSettings.catalogItemSyncEnableStatus",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM defaults project_catalog_item_sync_enabled = false.
				{PropertyPath: "properties.projectCatalogSettings.catalogItemSyncEnableStatus", Value: "Disabled"},
			},
			ComputedFields: []string{
				"properties.devCenterUri",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenter()) }
