package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHubNamespaceDisasterRecoveryConfig provides resource knowledge for
// Microsoft.EventHub/namespaces/disasterRecoveryConfigs.
//
// Contributing Terraform resource: azurerm_eventhub_namespace_disaster_recovery_config.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_namespace_disaster_recovery_config_resource.go
//     (schema L26-68, Create body L95-99)
//   - terraform-provider-azurerm internal/services/eventhub/validate/eventhub_names.go
//     (ValidateEventHubAuthorizationRuleName L35-40, reused for name)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/disasterrecoveryconfigs:
//     model_armdisasterrecovery.go, model_armdisasterrecoveryproperties.go
//
// Notes:
//   - name/namespace_name/resource_group_name are envelope / parent-reference fields;
//     not emitted as body rules.
//   - partner_namespace_id uses azure.ValidateResourceID (a semantic resource-id check);
//     it maps to properties.partnerNamespace and is emitted as a RequiredField. The
//     resource-id validation belongs on an azapin customizer validator, not a declarative
//     regex, so it is not emitted as a StringRule.
type EventHubNamespaceDisasterRecoveryConfig struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHubNamespaceDisasterRecoveryConfig)(nil)

// NewEventHubNamespaceDisasterRecoveryConfig returns knowledge for the disasterRecoveryConfigs resource.
func NewEventHubNamespaceDisasterRecoveryConfig() *EventHubNamespaceDisasterRecoveryConfig {
	return &EventHubNamespaceDisasterRecoveryConfig{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/namespaces/disasterRecoveryConfigs",
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
					Message:      "disaster recovery config name may contain only letters, numbers, periods, hyphens and underscores, up to 60 characters, beginning and ending with a letter or number",
				},
			},
			RequiredFields: []string{
				"properties.partnerNamespace",
			},
		},
	}
}

func init() { azwise.Register(NewEventHubNamespaceDisasterRecoveryConfig()) }
