package web

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteSlotHostNameBinding provides resource knowledge for
// Microsoft.Web/sites/slots/hostNameBindings.
//
// Backing AzureRM Terraform resource: azurerm_app_service_slot_custom_hostname_binding.
//
// Sources:
//   - terraform-provider-azurerm internal/services/web/app_service_slot_custom_hostname_binding_resource.go:24-133
//     (schema: app_service_slot_id/hostname/ssl_state/thumbprint all ForceNew; ssl_state
//     StringInSlice(IpBasedEnabled, SniEnabled); create mapping to HostNameBindingProperties
//     ssl_state->properties.sslState, thumbprint->properties.thumbprint)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/id_slothostnamebinding.go:115-118
//     (ID casing: Microsoft.Web/sites/slots/hostNameBindings)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_hostnamebindingproperties.go:6-16
//     (properties JSON tags: sslState *SslState, thumbprint *string)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/constants.go:2239-2242
//     (SslState enum: Disabled, IpBasedEnabled, SniEnabled)
//
// Notes:
//   - ssl_state AllowedValues use the full ARM SDK SslState set; AzureRM restricts the
//     Terraform surface to IpBasedEnabled/SniEnabled, but AzAPI sends raw ARM values.
//   - No Update; every schema field is ForceNew.
type SiteSlotHostNameBinding struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteSlotHostNameBinding)(nil)

// NewSiteSlotHostNameBinding returns knowledge for the sites/slots/hostNameBindings resource.
func NewSiteSlotHostNameBinding() *SiteSlotHostNameBinding {
	return &SiteSlotHostNameBinding{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/slots/hostNameBindings",
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

func init() { azwise.Register(NewSiteSlotHostNameBinding()) }
