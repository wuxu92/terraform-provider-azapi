package servicenetworking

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationLoadBalancerSubnetAssociation provides resource knowledge for
// Microsoft.ServiceNetworking/trafficControllers/associations.
//
// Mirrors azurerm_application_load_balancer_subnet_association. name and
// application_load_balancer_id (parent) live on the envelope; the create body
// sets properties.subnet.id and hardcodes properties.associationType = "subnets".
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicenetworking/application_load_balancer_subnet_association_resource.go
//     (schema Arguments lines 34-48, Create body lines 106-115, CRUD timeouts)
//   - go-azure-sdk resource-manager/servicenetworking/2025-01-01/associationsinterface:
//     id_association.go (ARM type casing "associations"), model_association.go,
//     model_associationproperties.go, model_associationsubnet.go, constants.go
//     (AssociationType "subnets")
//
// Not encoded (deliberate):
//   - name uses validate.ApplicationLoadBalancerSubnetAssociationName() (semantic
//     custom validator) and subnet_id uses a Subnet resource-ID validator; both are
//     semantic checks routed to an azapin customizer, not declarative rules.
type ApplicationLoadBalancerSubnetAssociation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationLoadBalancerSubnetAssociation)(nil)

// NewApplicationLoadBalancerSubnetAssociation returns knowledge for the associations resource.
func NewApplicationLoadBalancerSubnetAssociation() *ApplicationLoadBalancerSubnetAssociation {
	return &ApplicationLoadBalancerSubnetAssociation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceNetworking/trafficControllers/associations",
			ApiVersions:  []string{"2025-01-01"},
			// name and application_load_balancer_id are ForceNew envelope/parent refs.
			// subnet_id is NOT ForceNew (updatable).
			ForceNew: []azwise.ForceNewRule{},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// associationType is required and hardcoded to "subnets" by AzureRM.
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.associationType",
					AllowedValues: []string{"subnets"},
					Message:       "associationType must be `subnets`",
				},
			},
			// subnet.id (Required) and associationType (Required, hardcoded "subnets").
			RequiredFields: []string{
				"properties.subnet.id",
				"properties.associationType",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.associationType", Value: "subnets"},
			},
		},
	}
}

func init() { azwise.Register(NewApplicationLoadBalancerSubnetAssociation()) }
