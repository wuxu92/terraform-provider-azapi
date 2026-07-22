package streamanalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StreamAnalyticsOutput provides resource knowledge for
// Microsoft.StreamAnalytics/streamingjobs/outputs.
//
// One ARM type shared by ten TF resources, a discriminated union keyed by
// properties.datasource.type:
//   - azurerm_stream_analytics_output_blob            (Microsoft.Storage/Blob)
//   - azurerm_stream_analytics_output_cosmosdb        (Microsoft.Storage/DocumentDB)
//   - azurerm_stream_analytics_output_eventhub        (Microsoft.ServiceBus/EventHub)
//   - azurerm_stream_analytics_output_function        (Microsoft.AzureFunction)
//   - azurerm_stream_analytics_output_mssql           (Microsoft.Sql/Server/Database)
//   - azurerm_stream_analytics_output_powerbi         (PowerBI)
//   - azurerm_stream_analytics_output_servicebus_queue(Microsoft.ServiceBus/Queue)
//   - azurerm_stream_analytics_output_servicebus_topic(Microsoft.ServiceBus/Topic)
//   - azurerm_stream_analytics_output_synapse         (Microsoft.Sql/Server/DataWarehouse)
//   - azurerm_stream_analytics_output_table           (Microsoft.Storage/Table)
//
// Only knowledge universal to EVERY sink is unioned here. Per-sink value
// constraints (authentication_mode, blob_write_mode, batch sizes, connection
// details) live under the discriminated properties.datasource.properties.* subtree;
// those paths differ (and may collide with different allowed sets) between sinks,
// so they are intentionally NOT unioned to avoid corrupting validation of a sink
// that does not carry them.
//
// Sources:
//   - terraform-provider-azurerm internal/services/streamanalytics/stream_analytics_output_*_resource.go
//   - internal/services/streamanalytics/helpers_output.go:15-63 (serialization schema, universal)
//   - Timeouts (all sinks, e.g. blob :41-46): Create 30m / Read 5m / Update 30m / Delete 30m
//   - go-azure-sdk resource-manager/streamanalytics/2021-10-01-preview/outputs
//     model_outputproperties.go (datasource + serialization are the two required union members) /
//     constants.go PossibleValuesForEventSerializationType
//   - ARM type casing verified via outputs/id_output.go
//     StaticSegment("staticStreamingJobs", "streamingJobs") + StaticSegment("staticOutputs", "outputs").
//   - serialization.type AzureRM allows Avro/Csv/Json/Parquet; the SDK enum also
//     defines CustomClr/Delta, included since AzAPI forwards raw ARM values. The rule
//     fires only when properties.serialization is present (sinks like mssql/powerbi
//     omit it), so it is safe to union.
type StreamAnalyticsOutput struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StreamAnalyticsOutput)(nil)

// NewStreamAnalyticsOutput returns knowledge for the Stream Analytics output resource.
func NewStreamAnalyticsOutput() *StreamAnalyticsOutput {
	return &StreamAnalyticsOutput{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StreamAnalytics/streamingjobs/outputs",
			ApiVersions:  []string{"2021-10-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// serialization.type (fires only when serialization present). Full SDK
				// EventSerializationType set; AzureRM restricts to Avro/Csv/Json/Parquet.
				{
					PropertyPath:  "properties.serialization.type",
					AllowedValues: []string{"Avro", "Csv", "CustomClr", "Delta", "Json", "Parquet"},
					Message:       "invalid output serialization type",
				},
				// name: validation.StringIsNotEmpty (universal to all sinks).
				{PropertyPath: "", MinLength: 1, Message: "output name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewStreamAnalyticsOutput()) }
