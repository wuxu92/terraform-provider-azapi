package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGatewayCertificateAuthority provides resource knowledge for
// Microsoft.ApiManagement/service/gateways/certificateAuthorities.
//
// Mirrors azurerm_api_management_gateway_certificate_authority. `api_management_id`
// (ForceNew), `gateway_name` and `certificate_name` are envelope/parent references.
// The only ARM body field is `is_trusted` (properties.isTrusted, optional bool).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_gateway_certificate_authority_resource.go
//     schema (lines 42-58); Create body (lines 89-93); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/gatewaycertificateauthority
//     GatewayCertificateAuthorityContractProperties: isTrusted (optional bool).
type ApiManagementGatewayCertificateAuthority struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGatewayCertificateAuthority)(nil)

// NewApiManagementGatewayCertificateAuthority returns knowledge for the
// gateways/certificateAuthorities resource.
func NewApiManagementGatewayCertificateAuthority() *ApiManagementGatewayCertificateAuthority {
	return &ApiManagementGatewayCertificateAuthority{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/gateways/certificateAuthorities",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGatewayCertificateAuthority()) }
