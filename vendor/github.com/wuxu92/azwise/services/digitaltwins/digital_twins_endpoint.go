package digitaltwins

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DigitalTwinsEndpoint provides resource knowledge for
// Microsoft.DigitalTwins/digitalTwinsInstances/endpoints.
//
// AzureRM splits this single ARM type into three typed TF resources, one per
// endpoint-type variant of the DigitalTwinsEndpointResourceProperties
// discriminated union (discriminator = properties.endpointType):
//   - azurerm_digital_twins_endpoint_eventgrid  -> EventGrid variant
//   - azurerm_digital_twins_endpoint_eventhub   -> EventHub variant
//   - azurerm_digital_twins_endpoint_servicebus -> ServiceBus variant
//
// Only knowledge universal to every variant body is unioned here. Variant value
// constraints on a sub-object that only one kind sets ARE safe (they fire only
// when that sub-object is present), so per-variant secret paths are listed in
// SensitiveFields; but kind-specific RequiredFields are NOT unioned (see below).
//
// Sources:
//   - AzureRM internal/services/digitaltwins/digital_twins_endpoint_eventgrid_resource.go
//     :31-36 (timeouts), :48-86 (schema), :116-124 (EventGrid payload)
//   - AzureRM internal/services/digitaltwins/digital_twins_endpoint_eventhub_resource.go
//     :31-36 (timeouts), :48-83 (schema: sensitive conn strings), :113-120 (EventHub payload)
//   - AzureRM internal/services/digitaltwins/digital_twins_endpoint_servicebus_resource.go
//     :31-36 (timeouts), :48-83 (schema: sensitive conn strings), :113-120 (ServiceBus payload)
//   - AzureRM internal/services/digitaltwins/validate/digital_twins_instance_name.go
//     :11-33 (endpoint name uses DigitalTwinsInstanceName: length 3-63 + regex)
//   - go-azure-sdk resource-manager/digitaltwins/2023-01-31/endpoints:
//     model_digitaltwinsendpointresourceproperties.go (base: authenticationType,
//     deadLetterSecret, endpointType), model_eventgrid.go (accessKey1/accessKey2/TopicEndpoint),
//     model_eventhub.go (connectionStringPrimaryKey/SecondaryKey), model_servicebus.go
//     (primaryConnectionString/secondaryConnectionString), constants.go
//     (EndpointType + AuthenticationType enums)
//
// Not encoded (deliberate):
//   - name/digital_twins_id are envelope + parent references, not body properties.
//   - Kind-specific RequiredFields are NOT unioned: EventGrid requires
//     properties.accessKey1 + properties.TopicEndpoint; EventHub requires none of
//     the ServiceBus fields; etc. Unioning any of them would reject the other two
//     variants' legitimate bodies. Only properties.endpointType (the discriminator,
//     required for every body) is universal.
//   - eventgrid_topic_endpoint uses IsURLWithHTTPS (a semantic URL check, not a
//     clean enum/regex over the raw value) and maps to a single-variant path
//     (properties.TopicEndpoint), so it is left to a customizer validator, not a
//     declarative StringRule here.
//   - AzureRM always sends authenticationType = KeyBased; the ARM API also accepts
//     IdentityBased, so the enum below carries both (AzAPI users send raw ARM values).
type DigitalTwinsEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DigitalTwinsEndpoint)(nil)

// NewDigitalTwinsEndpoint returns knowledge for the endpoints resource,
// merged across the EventGrid/EventHub/ServiceBus TF variants.
func NewDigitalTwinsEndpoint() *DigitalTwinsEndpoint {
	return &DigitalTwinsEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DigitalTwins/digitalTwinsInstances/endpoints",
			ApiVersions:  []string{"2023-01-31"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath = name attribute);
					// all three variants use validate.DigitalTwinsInstanceName.
					Regex:     `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 chars, begin and end with a letter or number, and contain only letters, numbers, and hyphens",
				},
				{
					PropertyPath:  "properties.endpointType",
					AllowedValues: []string{"EventGrid", "EventHub", "ServiceBus"},
					Message:       "endpoint type must be one of EventGrid, EventHub, ServiceBus",
				},
				{
					PropertyPath:  "properties.authenticationType",
					AllowedValues: []string{"IdentityBased", "KeyBased"},
					Message:       "authentication type must be IdentityBased or KeyBased",
				},
			},
			// Secret body paths across all three variants. Each path exists only in
			// its own variant body, so listing them together is safe.
			SensitiveFields: []string{
				"properties.deadLetterSecret",
				"properties.connectionStringPrimaryKey",   // EventHub
				"properties.connectionStringSecondaryKey", // EventHub
				"properties.primaryConnectionString",      // ServiceBus
				"properties.secondaryConnectionString",    // ServiceBus
			},
			// Universal to every endpoint body; the discriminator must be present.
			RequiredFields: []string{
				"properties.endpointType",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDigitalTwinsEndpoint()) }
