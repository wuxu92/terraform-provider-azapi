package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGroup provides resource knowledge for
// Microsoft.ApiManagement/service/groups.
//
// Mirrors azurerm_api_management_group.
//
// Sources (terraform-provider-azurerm internal/services/apimanagement):
//   - api_management_group_resource.go (schema L41-76; expand L108-115)
//   - SDK model GroupCreateParametersProperties (displayName required; description/externalId/type optional)
//   - group/constants.go GroupType enum {custom, external, system}
//
// externalId and type are ForceNew in the AzureRM schema.
type ApiManagementGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGroup)(nil)

// NewApiManagementGroup returns knowledge for the groups resource.
func NewApiManagementGroup() *ApiManagementGroup {
	return &ApiManagementGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/groups",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.externalId"},
				{PropertyPath: "properties.type"},
			},
			RequiredFields: []string{
				"properties.displayName",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.type", Value: "custom"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				{PropertyPath: "properties.type", AllowedValues: []string{"custom", "external", "system"}, Message: "type must be one of custom, external, system"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGroup()) }
