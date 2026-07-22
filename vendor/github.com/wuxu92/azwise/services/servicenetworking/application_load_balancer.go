package servicenetworking

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationLoadBalancer provides resource knowledge for
// Microsoft.ServiceNetworking/trafficControllers.
//
// Mirrors azurerm_application_load_balancer. name, resource_group_name and
// location live on the operational envelope; the ARM create body carries only
// location + tags (no properties are user-settable).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicenetworking/application_load_balancer_resource.go
//     (schema Arguments/Attributes lines 36-60, Create body lines 101-104, CRUD timeouts)
//   - go-azure-sdk resource-manager/servicenetworking/2025-01-01/trafficcontrollerinterface:
//     id_trafficcontroller.go (ARM type casing "trafficControllers" under
//     "Microsoft.ServiceNetworking"), model_trafficcontroller.go,
//     model_trafficcontrollerproperties.go
type ApplicationLoadBalancer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationLoadBalancer)(nil)

// NewApplicationLoadBalancer returns knowledge for the trafficControllers resource.
func NewApplicationLoadBalancer() *ApplicationLoadBalancer {
	return &ApplicationLoadBalancer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceNetworking/trafficControllers",
			ApiVersions:  []string{"2025-01-01"},
			// location is ForceNew (commonschema.Location()); name is ForceNew but
			// lives on the resource-name envelope.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name validation (schema line 42).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,62}[a-zA-Z0-9]$`,
					Message:      "name must begin with a letter or number, end with a letter, number or underscore, be 1-64 characters, and contain only letters, numbers, underscores, periods, or hyphens",
				},
			},
			// configurationEndpoints is populated by Azure (read-only) and surfaces
			// as primary_configuration_endpoint (Attributes line 55-58).
			ComputedFields: []string{
				"properties.configurationEndpoints",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationLoadBalancer()) }
