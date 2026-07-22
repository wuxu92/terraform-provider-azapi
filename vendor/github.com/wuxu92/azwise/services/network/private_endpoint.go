package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateEndpoint provides resource knowledge for
// Microsoft.Network/privateEndpoints.
//
// Mirrors azurerm_private_endpoint. name and resource_group_name are envelope-owned.
//
// azurerm_private_endpoint_application_security_group_association is NOT a separate
// ARM resource: it folds into this type's body
// (properties.applicationSecurityGroups). It attaches/detaches ASG references on the
// existing privateEndpoint, so it contributes no distinct ARM type and no separate
// knowledge file — documented here (its only field is an ASG resource ID, which has
// no declarative rule to add).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/private_endpoint_resource.go
//     (schema 66-299, expand 346-357, timeouts 59-64)
//   - internal/services/network/private_endpoint_application_security_group_association_resource.go
//     (folded-in ASG association)
//   - internal/services/network/validate/private_link_name.go:30 (name regex)
//   - go-azure-sdk resource-manager/network/2025-01-01/privateendpoints:
//     id_privateendpoint.go:104 (segment casing), model_privateendpointproperties.go
//
// Not encoded (deliberate):
//   - private_service_connection (Required) expands to
//     properties.{privateLinkServiceConnections,manualPrivateLinkServiceConnections}[*];
//     its ExactlyOneOf (resource_id vs alias) and per-connection rules live under
//     array elements and cannot be lowered by azwise/azapin.
type PrivateEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateEndpoint)(nil)

// NewPrivateEndpoint returns knowledge for the privateEndpoints resource.
func NewPrivateEndpoint() *PrivateEndpoint {
	return &PrivateEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateEndpoints",
			ApiVersions:  []string{"2025-01-01"},
			// location, edge_zone, subnet_id and custom_network_interface_name all
			// replace the endpoint.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation"},
				{PropertyPath: "properties.subnet.id"},
				{PropertyPath: "properties.customNetworkInterfaceName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: 1-80 chars, begin with letter/number, end with
					// letter/number/underscore, may contain letters/numbers/._-.
					Regex:   `^([a-zA-Z\d])([a-zA-Z\d-\_\.]{0,78})([a-zA-Z\d\_])$`,
					Message: "name must be 1-80 characters, begin with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, periods, hyphens or underscores",
				},
			},
			RequiredFields: []string{
				"properties.subnet.id",
			},
		},
	}
}

func init() { azwise.Register(NewPrivateEndpoint()) }
