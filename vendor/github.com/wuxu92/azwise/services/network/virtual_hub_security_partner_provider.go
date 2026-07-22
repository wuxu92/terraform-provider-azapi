package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHubSecurityPartnerProvider provides resource knowledge for
// Microsoft.Network/securityPartnerProviders.
//
// Mirrors azurerm_virtual_hub_security_partner_provider. name and
// resource_group_name are envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_security_partner_provider_resource.go
//     (schema 45-75, expand 100-112, timeouts 33-38)
//   - go-azure-sdk resource-manager/network/2025-01-01/securitypartnerproviders:
//     id_securitypartnerprovider.go:104 (segment casing),
//     model_securitypartnerproviderpropertiesformat.go, constants.go (SecurityProviderName 108-112)
type VirtualHubSecurityPartnerProvider struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHubSecurityPartnerProvider)(nil)

// NewVirtualHubSecurityPartnerProvider returns knowledge for the securityPartnerProviders resource.
func NewVirtualHubSecurityPartnerProvider() *VirtualHubSecurityPartnerProvider {
	return &VirtualHubSecurityPartnerProvider{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/securityPartnerProviders",
			ApiVersions:  []string{"2025-01-01"},
			// location, security_provider_name and virtual_hub_id all replace the resource.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.securityProviderName"},
				{PropertyPath: "properties.virtualHub.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.securityProviderName",
					AllowedValues: []string{"ZScaler", "IBoss", "Checkpoint"},
					Message:       "security_provider_name must be ZScaler, IBoss or Checkpoint",
				},
			},
			RequiredFields: []string{
				"properties.securityProviderName",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHubSecurityPartnerProvider()) }
