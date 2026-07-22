package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkSecurityPerimeterAssociation provides resource knowledge for
// Microsoft.Network/networkSecurityPerimeters/resourceAssociations.
//
// Mirrors azurerm_network_security_perimeter_association. name, resource_id and
// network_security_perimeter_profile_id are ForceNew; access_mode is mutable.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_security_perimeter_association_resource.go
//     (typed schema, Create/Update 30m, Read 5m)
//   - go-azure-sdk resource-manager/network/2025-01-01/networksecurityperimeterassociations:
//     model_nspassociationproperties.go, constants.go (AssociationAccessMode),
//     id_resourceassociation.go (segment casing
//     "networkSecurityPerimeters/resourceAssociations")
type NetworkSecurityPerimeterAssociation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkSecurityPerimeterAssociation)(nil)

// NewNetworkSecurityPerimeterAssociation returns knowledge for the
// networkSecurityPerimeters/resourceAssociations resource.
func NewNetworkSecurityPerimeterAssociation() *NetworkSecurityPerimeterAssociation {
	return &NetworkSecurityPerimeterAssociation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkSecurityPerimeters/resourceAssociations",
			ApiVersions:  []string{"2025-01-01"},
			// resource_id maps to properties.privateLinkResource.id and is ForceNew;
			// name is the resource name (envelope). Encode the body-path ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.privateLinkResource.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `(^[a-zA-Z0-9]+[a-zA-Z0-9_.-]{0,78}[a-zA-Z0-9_]+$)|(^[a-zA-Z0-9]$)`,
					Message: "name must be 1-80 chars, start with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, underscores, periods, or hyphens",
				},
				{
					PropertyPath:  "properties.accessMode",
					AllowedValues: []string{"Audit", "Enforced", "Learning"},
					Message:       "access_mode must be Audit, Enforced or Learning",
				},
			},
			RequiredFields: []string{
				"properties.accessMode",
				"properties.privateLinkResource.id",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkSecurityPerimeterAssociation()) }
