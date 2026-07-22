package web

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteHostNameBinding provides resource knowledge for
// Microsoft.Web/sites/hostNameBindings.
//
// Backing AzureRM Terraform resources:
//   - azurerm_app_service_custom_hostname_binding (creates the binding)
//   - azurerm_app_service_certificate_binding (updates ssl_state/thumbprint on an existing
//     binding; a pure sub-setting operation over the same ARM type, no separate registration)
//
// Sources:
//   - terraform-provider-azurerm internal/services/web/app_service_custom_hostname_binding_resource.go:25-131
//     (schema: hostname/app_service_name/ssl_state/thumbprint all ForceNew; ssl_state
//     StringInSlice(IpBasedEnabled, SniEnabled); create mapping to HostNameBindingProperties
//     ssl_state->properties.sslState, thumbprint->properties.thumbprint, siteName derived)
//   - terraform-provider-azurerm internal/services/web/app_service_certificate_binding_resource.go:25-82
//     (certificate binding: ssl_state Required StringInSlice(IpBasedEnabled, SniEnabled),
//     mapped onto the existing hostNameBinding's properties.sslState/thumbprint)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/id_hostnamebinding.go:109-112
//     (ID casing: Microsoft.Web/sites/hostNameBindings)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_hostnamebindingproperties.go:6-16
//     (properties JSON tags: sslState *SslState, thumbprint *string, siteName *string)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/constants.go:2239-2242
//     (SslState enum: Disabled, IpBasedEnabled, SniEnabled)
//
// Notes:
//   - ssl_state AllowedValues use the full ARM SDK SslState set (Disabled, IpBasedEnabled,
//     SniEnabled); AzureRM restricts the Terraform surface to the latter two, but AzAPI sends
//     raw ARM values.
//   - This resource has no Update; every schema field is ForceNew, so an in-place change is a
//     replacement.
//   - properties.siteName is derived from the parent site by AzureRM, not user-supplied; it is
//     not encoded as a RequiredField.
type SiteHostNameBinding struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteHostNameBinding)(nil)

// NewSiteHostNameBinding returns knowledge for the sites/hostNameBindings resource.
func NewSiteHostNameBinding() *SiteHostNameBinding {
	return &SiteHostNameBinding{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/hostNameBindings",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.sslState"},
				{PropertyPath: "properties.thumbprint"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.sslState",
					AllowedValues: []string{
						"Disabled",
						"IpBasedEnabled",
						"SniEnabled",
					},
					Message: "must be a valid SSL state",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSiteHostNameBinding()) }
