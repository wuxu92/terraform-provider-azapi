package servicenetworking

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationLoadBalancerSecurityPolicy provides resource knowledge for
// Microsoft.ServiceNetworking/trafficControllers/securityPolicies.
//
// Mirrors azurerm_application_load_balancer_security_policy. name and
// application_load_balancer_id (parent) live on the envelope; the create body
// sets location + properties.wafPolicy.id.
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicenetworking/application_load_balancer_security_policy_resource.go
//     (schema Arguments lines 36-56, Create body lines 102-110, CRUD timeouts)
//   - go-azure-sdk resource-manager/servicenetworking/2025-01-01/securitypoliciesinterface:
//     id_securitypolicy.go (ARM type casing "securityPolicies"),
//     model_securitypolicy.go, model_securitypolicyproperties.go, model_wafpolicy.go
//
// Not encoded (deliberate):
//   - web_application_firewall_policy_id uses a WAF-policy resource-ID validator
//     (semantic) routed to an azapin customizer, not a declarative rule.
type ApplicationLoadBalancerSecurityPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationLoadBalancerSecurityPolicy)(nil)

// NewApplicationLoadBalancerSecurityPolicy returns knowledge for the securityPolicies resource.
func NewApplicationLoadBalancerSecurityPolicy() *ApplicationLoadBalancerSecurityPolicy {
	return &ApplicationLoadBalancerSecurityPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceNetworking/trafficControllers/securityPolicies",
			ApiVersions:  []string{"2025-01-01"},
			// location (commonschema.Location()) and web_application_firewall_policy_id
			// (ResourceIDReferenceRequiredForceNew -> properties.wafPolicy.id) are
			// ForceNew. name/application_load_balancer_id are envelope/parent refs.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.wafPolicy.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name validation (schema lines 42-45).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,62}[a-zA-Z0-9])?$`,
					Message:      "name must begin with a letter or number, end with a letter or number, be 1-64 characters, and contain only letters, numbers, underscores, periods, or hyphens",
				},
			},
			// wafPolicy.id (Required, model json:"id" no omitempty).
			RequiredFields: []string{
				"properties.wafPolicy.id",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationLoadBalancerSecurityPolicy()) }
