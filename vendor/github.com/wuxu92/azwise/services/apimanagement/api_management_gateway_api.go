package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGatewayApi provides resource knowledge for
// Microsoft.ApiManagement/service/gateways/apis.
//
// Mirrors azurerm_api_management_gateway_api. Both schema fields (`api_id` and
// `gateway_id`) are envelope/parent references and are ForceNew by construction,
// so they live on the operational envelope (name + parent_id) rather than the
// ARM body. The ARM body is an empty AssociationContract with only the
// server-computed provisioningState, so there are no settable body properties.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_gateway_api_resource.go
//     schema (lines 46-59); Create body (empty AssociationContract, lines 97-100);
//     Timeouts 30m/5m/-/30m (lines 35-39).
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/gatewayapi
//     AssociationContract: only provisioningState (read-only).
type ApiManagementGatewayApi struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGatewayApi)(nil)

// NewApiManagementGatewayApi returns knowledge for the gateways/apis resource.
func NewApiManagementGatewayApi() *ApiManagementGatewayApi {
	return &ApiManagementGatewayApi{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/gateways/apis",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGatewayApi()) }
