package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateLinkService provides resource knowledge for
// Microsoft.Network/privateLinkServices.
//
// Mirrors azurerm_private_link_service. name and resource_group_name are
// envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/private_link_service_resource.go
//     (schema 55-169, expand 226-244, timeouts 48-53)
//   - internal/services/network/validate/private_link_name.go:30 (name regex)
//   - go-azure-sdk resource-manager/network/2025-01-01/privatelinkservices:
//     id_privatelinkservice.go:104 (segment casing), model_privatelinkserviceproperties.go
//
// Not encoded (deliberate):
//   - auto_approval_subscription_ids / visibility_subscription_ids are UUID (or "*")
//     arrays; their per-element IsUUID checks live under array elements and cannot be
//     lowered by azwise/azapin.
//   - nat_ip_configuration (Required) expands to properties.ipConfigurations[*]; its
//     per-config rules (name, private_ip_address_version IPv4) live under array
//     elements and are documented here, not emitted.
type PrivateLinkService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateLinkService)(nil)

// NewPrivateLinkService returns knowledge for the privateLinkServices resource.
func NewPrivateLinkService() *PrivateLinkService {
	return &PrivateLinkService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateLinkServices",
			ApiVersions:  []string{"2025-01-01"},
			// location and load_balancer_frontend_ip_configuration_ids replace the service.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.loadBalancerFrontendIpConfigurations"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableProxyProtocol", Value: false},
			},
			RequiredFields: []string{
				"properties.ipConfigurations",
			},
			// destination_ip_address and load_balancer_frontend_ip_configuration_ids
			// are ExactlyOneOf.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.destinationIPAddress",
						"properties.loadBalancerFrontendIpConfigurations",
					},
					Message: "exactly one of destination_ip_address or load_balancer_frontend_ip_configuration_ids must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPrivateLinkService()) }
