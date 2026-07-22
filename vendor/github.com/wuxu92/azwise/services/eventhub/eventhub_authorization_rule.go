package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHubAuthorizationRule provides resource knowledge for
// Microsoft.EventHub/namespaces/eventhubs/authorizationRules.
//
// Contributing Terraform resource: azurerm_eventhub_authorization_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_authorization_rule_resource.go
//     (schema L24-74, Create body L106-111, expand rights L214-230)
//   - terraform-provider-azurerm internal/services/eventhub/helpers.go
//     (listen/send/manage schema + CustomizeDiff L64-78)
//   - terraform-provider-azurerm internal/services/eventhub/validate/eventhub_names.go
//     (ValidateEventHubAuthorizationRuleName L35-40)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/authorizationruleseventhubs:
//     model_authorizationrule.go, model_authorizationruleproperties.go, constants.go
//
// Notes:
//   - name/namespace_name/eventhub_name/resource_group_name are envelope / parent-reference
//     fields; not emitted as body rules.
//   - listen/send/manage booleans expand into the properties.rights array of AccessRights
//     enums (Listen/Send/Manage); a many-to-array mapping not expressible as a single-path
//     rule, so only the required-array constraint is emitted. CustomizeDiff (helpers.go
//     L64-78) requires at least one right and manage implies listen+send.
//   - primary_connection_string / primary_key etc. are data-plane keys, not body
//     properties; omitted.
type EventHubAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHubAuthorizationRule)(nil)

// NewEventHubAuthorizationRule returns knowledge for the eventhubs/authorizationRules resource.
func NewEventHubAuthorizationRule() *EventHubAuthorizationRule {
	return &EventHubAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/namespaces/eventhubs/authorizationRules",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,58}[a-zA-Z0-9])?$`,
					Message:      "authorization rule name may contain only letters, numbers, periods, hyphens and underscores, up to 60 characters, beginning and ending with a letter or number",
				},
			},
			RequiredFields: []string{
				"properties.rights",
			},
		},
	}
}

func init() { azwise.Register(NewEventHubAuthorizationRule()) }
