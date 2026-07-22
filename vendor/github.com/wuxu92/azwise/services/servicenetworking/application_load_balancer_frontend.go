package servicenetworking

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationLoadBalancerFrontend provides resource knowledge for
// Microsoft.ServiceNetworking/trafficControllers/frontends.
//
// Mirrors azurerm_application_load_balancer_frontend. name and
// application_load_balancer_id (parent) live on the envelope; the create body
// carries only location + tags with an empty FrontendProperties block.
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicenetworking/application_load_balancer_frontend_resource.go
//     (schema Arguments/Attributes lines 35-62, Create body lines 120-124, CRUD timeouts)
//   - go-azure-sdk resource-manager/servicenetworking/2025-01-01/frontendsinterface:
//     id_frontend.go (ARM type casing "frontends"), model_frontend.go,
//     model_frontendproperties.go
type ApplicationLoadBalancerFrontend struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationLoadBalancerFrontend)(nil)

// NewApplicationLoadBalancerFrontend returns knowledge for the frontends resource.
func NewApplicationLoadBalancerFrontend() *ApplicationLoadBalancerFrontend {
	return &ApplicationLoadBalancerFrontend{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceNetworking/trafficControllers/frontends",
			ApiVersions:  []string{"2025-01-01"},
			// name and application_load_balancer_id are ForceNew but both live on the
			// resource-name/parent envelope, so no body ForceNew rules apply.
			ForceNew: []azwise.ForceNewRule{},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name is StringIsNotEmpty (schema line 41).
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "name must not be empty",
				},
			},
			// fqdn is populated by Azure (read-only) and surfaces as
			// fully_qualified_domain_name (Attributes line 57-60).
			ComputedFields: []string{
				"properties.fqdn",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationLoadBalancerFrontend()) }
