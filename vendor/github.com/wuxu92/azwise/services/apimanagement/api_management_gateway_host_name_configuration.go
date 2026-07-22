package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGatewayHostNameConfiguration provides resource knowledge for
// Microsoft.ApiManagement/service/gateways/hostnameConfigurations.
//
// Mirrors azurerm_api_management_gateway_host_name_configuration. `name`,
// `api_management_id` (ForceNew) and `gateway_name` are envelope/parent
// references. The ARM body carries the required `host_name` (properties.hostname)
// and `certificate_id` (properties.certificateId), plus optional flags:
// `request_client_certificate_enabled` -> properties.negotiateClientCertificate,
// `tls10_enabled` -> properties.tls10Enabled, `tls11_enabled` ->
// properties.tls11Enabled, `http2_enabled` -> properties.http2Enabled (default true).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_gateway_host_name_configuration_resource.go
//     schema (lines 43-87); Create body (lines 118-127); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/gatewayhostnameconfiguration
//     GatewayHostnameConfigurationContractProperties: hostname, certificateId,
//     negotiateClientCertificate, tls10Enabled, tls11Enabled, http2Enabled.
type ApiManagementGatewayHostNameConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGatewayHostNameConfiguration)(nil)

// NewApiManagementGatewayHostNameConfiguration returns knowledge for the
// gateways/hostnameConfigurations resource.
func NewApiManagementGatewayHostNameConfiguration() *ApiManagementGatewayHostNameConfiguration {
	return &ApiManagementGatewayHostNameConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/gateways/hostnameConfigurations",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// host_name and certificate_id are Required in the AzureRM schema.
			RequiredFields: []string{
				"properties.hostname",
				"properties.certificateId",
			},
			// http2_enabled defaults to true (AzureRM schema Default: true).
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.http2Enabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGatewayHostNameConfiguration()) }
