package streamanalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StreamAnalyticsInput provides resource knowledge for
// Microsoft.StreamAnalytics/streamingjobs/inputs.
//
// One ARM type shared by six TF resources, a discriminated union keyed by
// properties.type (Stream / Reference) and properties.datasource.type:
//   - azurerm_stream_analytics_reference_input_blob    (type=Reference, Microsoft.Storage/Blob)
//   - azurerm_stream_analytics_reference_input_mssql   (type=Reference, Microsoft.Sql/Server/Database)
//   - azurerm_stream_analytics_stream_input_blob       (type=Stream, Microsoft.Storage/Blob)
//   - azurerm_stream_analytics_stream_input_eventhub   (type=Stream, Microsoft.ServiceBus/EventHub)
//   - azurerm_stream_analytics_stream_input_eventhub_v2(type=Stream, Microsoft.EventHub/EventHub)
//   - azurerm_stream_analytics_stream_input_iothub     (type=Stream, Microsoft.Devices/IotHubs)
//
// Only knowledge universal to EVERY input shape is unioned here. Per-source value
// constraints and Required fields live under the discriminated
// properties.datasource.properties.* subtree (which differs per source and may
// collide across sources), so they are intentionally NOT unioned — folding a
// source-specific rule would corrupt validation for inputs that omit it.
//
// Sources:
//   - terraform-provider-azurerm internal/services/streamanalytics/stream_analytics_*_input_*_resource.go
//   - internal/services/streamanalytics/helpers_input.go:15-53 (serialization schema, universal)
//   - Timeouts (all inputs, e.g. stream_input_eventhub :40-45): Create 30m / Read 5m / Update 30m / Delete 30m
//   - go-azure-sdk resource-manager/streamanalytics/2020-03-01/inputs
//     model_streaminputproperties.go (type="Stream") / model_referenceinputproperties.go (type="Reference")
//     (datasource + serialization are the union members) / constants.go PossibleValuesForEventSerializationType
//   - ARM type casing verified via inputs/id_input.go
//     StaticSegment("staticStreamingJobs", "streamingJobs") + StaticSegment("staticInputs", "inputs").
//   - serialization.type: AzureRM stream inputs allow Avro/Csv/Json; the SDK enum also
//     defines Parquet, included since AzAPI forwards raw ARM values. The rule fires only
//     when properties.serialization is present (reference_input_mssql omits it), so it is
//     safe to union.
type StreamAnalyticsInput struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StreamAnalyticsInput)(nil)

// NewStreamAnalyticsInput returns knowledge for the Stream Analytics input resource.
func NewStreamAnalyticsInput() *StreamAnalyticsInput {
	return &StreamAnalyticsInput{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StreamAnalytics/streamingjobs/inputs",
			ApiVersions:  []string{"2020-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// serialization.type (fires only when serialization present). Full SDK
				// EventSerializationType set; AzureRM stream inputs restrict to Avro/Csv/Json.
				{
					PropertyPath:  "properties.serialization.type",
					AllowedValues: []string{"Avro", "Csv", "Json", "Parquet"},
					Message:       "invalid input serialization type",
				},
				// name: validation.StringIsNotEmpty (universal to all input kinds).
				{PropertyPath: "", MinLength: 1, Message: "input name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewStreamAnalyticsInput()) }
