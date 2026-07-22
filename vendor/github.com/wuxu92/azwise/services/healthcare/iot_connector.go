// Copyright (c) HashiCorp, Inc.

package healthcare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotConnector provides resource knowledge for
// Microsoft.HealthcareApis/workspaces/iotconnectors.
//
// Contributing TF resource:
//   - azurerm_healthcare_medtech_service
//
// Sources:
//   - AzureRM internal/services/healthcare/healthcare_medtech_service_resource.go
//     (schema :32-103, Create :105-162)
//   - AzureRM internal/services/healthcare/validate/medtech_service_name.go (name rule)
//   - go-azure-sdk .../healthcareapis/2022-12-01/iotconnectors:
//     id_iotconnector.go (segment casing: workspaces/iotConnectors),
//     model_iotconnector.go, model_iotconnectorproperties.go,
//     model_ioteventhubingestionendpointconfiguration.go, constants.go
type IotConnector struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotConnector)(nil)

func NewIotConnector() *IotConnector {
	return &IotConnector{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthcareApis/workspaces/iotconnectors",
			ApiVersions:  []string{"2022-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// NOTE: workspace_id is the parent resource reference (not a body path); omitted.
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.MedTechServiceName:
					// 3-24 chars, starts/ends alphanumeric, dashes allowed.
					Regex:   `^[0-9a-zA-Z][-0-9a-zA-Z]{1,22}[0-9a-zA-Z]$`,
					Message: "must be 3-24 characters, start and end with a letter or number, and contain only letters, numbers, and dashes",
				},
				// NOTE: eventhub_namespace_name (ValidateEventHubNamespaceName, and
				// AzureRM appends ".servicebus.windows.net" before storing in
				// properties.ingestionEndpointConfiguration.fullyQualifiedEventHubNamespace),
				// eventhub_name (ValidateEventHubName), and eventhub_consumer_group_name
				// (ValidateEventHubConsumerName) are resource-specific semantic validators
				// with value transforms; not representable as declarative StringRules.
			},
			RequiredFields: []string{
				// eventhub_name / eventhub_consumer_group_name / eventhub_namespace_name /
				// device_mapping_json are all Required.
				"properties.ingestionEndpointConfiguration.eventHubName",
				"properties.ingestionEndpointConfiguration.consumerGroup",
				"properties.ingestionEndpointConfiguration.fullyQualifiedEventHubNamespace",
				"properties.deviceMapping",
			},
			// Azure-populated, read-only field absent from the create body.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewIotConnector()) }
