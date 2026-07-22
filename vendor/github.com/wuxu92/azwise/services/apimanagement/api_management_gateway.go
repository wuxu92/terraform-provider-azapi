package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGateway provides resource knowledge for
// Microsoft.ApiManagement/service/gateways.
//
// Mirrors azurerm_api_management_gateway. `name` and `api_management_id` are
// envelope/parent references (ForceNew), so they are not repeated as body rules.
// The ARM body carries `description` and the required `location_data` object
// (azurerm's `location_data` list is MaxItems:1, projecting onto the single
// ResourceLocationDataContract). Within location_data, `name` is required and
// `region` maps to ARM `countryOrRegion`.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_gateway_resource.go
//     schema (lines 41-81); Create body (lines 112-120); location_data expand
//     (lines 185-207, region -> countryOrRegion); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/gateway
//     GatewayContractProperties: description, locationData
//     (ResourceLocationDataContract: name (required), city, district, countryOrRegion).
type ApiManagementGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGateway)(nil)

// NewApiManagementGateway returns knowledge for the gateways resource.
func NewApiManagementGateway() *ApiManagementGateway {
	return &ApiManagementGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/gateways",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// location_data (Required, MaxItems:1) and its `name` field (Required)
			// project onto the required ARM locationData object.
			RequiredFields: []string{
				"properties.locationData.name",
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGateway()) }
