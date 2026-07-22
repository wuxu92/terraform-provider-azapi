package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHubConsumerGroup provides resource knowledge for
// Microsoft.Devices/IotHubs/eventHubEndpoints/ConsumerGroups.
//
// Contributing Terraform resource: azurerm_iothub_consumer_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_consumer_group_resource.go
//     (schema L45-67, Create body L95-105, timeouts L39-43)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_consumer_group.go
//     (IoTHubConsumerGroupName L10-18)
//   - terraform-provider-azurerm internal/services/iothub/parse/consumer_group.go
//     (ARM path .../iotHubs/{}/eventHubEndpoints/{}/consumerGroups/{} L46)
//   - go-azure-sdk (kermit) sdk/iothub/2022-04-30-preview/iothub:
//     EventHubConsumerGroupBodyDescription / EventHubConsumerGroupName (Name only)
//
// Notes:
//   - name/iothub_name/eventhub_endpoint_name/resource_group_name are envelope /
//     parent-reference fields. The body only carries the consumer group name; there
//     are no settable properties, so only the name validator is emitted.
type IotHubConsumerGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHubConsumerGroup)(nil)

// NewIotHubConsumerGroup returns knowledge for the eventHubEndpoints/ConsumerGroups resource.
func NewIotHubConsumerGroup() *IotHubConsumerGroup {
	return &IotHubConsumerGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Devices/IotHubs/eventHubEndpoints/ConsumerGroups",
			ApiVersions:  []string{"2022-04-30-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-zA-Z-._]{1,}$`,
					Message:      "consumer group name may only contain alphanumeric characters, dashes, periods and underscores",
				},
			},
		},
	}
}

func init() { azwise.Register(NewIotHubConsumerGroup()) }
