package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHubConsumerGroup provides resource knowledge for
// Microsoft.EventHub/namespaces/eventhubs/consumergroups.
//
// Contributing Terraform resource: azurerm_eventhub_consumer_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_consumer_group_resource.go
//     (schema L42-73, Create body L102-107, timeouts L116/L182/L203)
//   - terraform-provider-azurerm internal/services/eventhub/validate/eventhub_names.go
//     (ValidateEventHubConsumerName L28-33)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/consumergroups:
//     model_consumergroupproperties.go
//
// Notes:
//   - name/namespace_name/eventhub_name/resource_group_name are envelope / parent-reference
//     fields; not emitted as body rules.
type EventHubConsumerGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHubConsumerGroup)(nil)

// NewEventHubConsumerGroup returns knowledge for the consumergroups resource.
func NewEventHubConsumerGroup() *EventHubConsumerGroup {
	return &EventHubConsumerGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/namespaces/eventhubs/consumergroups",
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
					Regex:        `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,48}[a-zA-Z0-9])?$`,
					Message:      "consumer group name may contain only letters, numbers, periods, hyphens and underscores, up to 50 characters, beginning and ending with a letter or number",
				},
				{
					PropertyPath: "properties.userMetadata",
					MinLength:    1,
					MaxLength:    1024,
					Message:      "user_metadata must be between 1 and 1024 characters",
				},
			},
			ComputedFields: []string{
				"properties.createdAt",
				"properties.updatedAt",
			},
		},
	}
}

func init() { azwise.Register(NewEventHubConsumerGroup()) }
