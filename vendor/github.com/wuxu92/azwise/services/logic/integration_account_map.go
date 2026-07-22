package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountMap provides resource knowledge for
// Microsoft.Logic/integrationAccounts/maps.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_map.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_map_resource.go
//     (schema L42-79, Create body L105-120)
//   - terraform-provider-azurerm internal/services/logic/validate/integration_account_map_name.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountmaps:
//     model_integrationaccountmap.go, model_integrationaccountmapproperties.go,
//     constants.go (PossibleValuesForMapType), id_map.go
//
// Notes:
//   - name (ForceNew), integration_account_name (ForceNew parent segment) and
//     resource_group_name are envelope-owned; only the map-name regex is emitted.
//   - map_type maps to properties.mapType; the full MapType SDK enum is emitted.
//   - content → properties.content; contentType is derived by AzureRM from map_type
//     (Liquid→text/plain, else application/xml) and is provider-controlled, not user-facing.
type IntegrationAccountMap struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountMap)(nil)

// NewIntegrationAccountMap returns knowledge for the maps child resource.
func NewIntegrationAccountMap() *IntegrationAccountMap {
	return &IntegrationAccountMap{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/maps",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9-().]+$`,
					MaxLength:    80,
					Message:      "map name contains only letters, numbers, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath:  "properties.mapType",
					AllowedValues: []string{"Liquid", "NotSpecified", "Xslt", "Xslt30", "Xslt20"},
					Message:       "map_type must be a valid MapType",
				},
			},
			RequiredFields: []string{
				"properties.mapType",
				"properties.content",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountMap()) }
