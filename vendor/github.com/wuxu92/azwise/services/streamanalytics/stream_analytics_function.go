package streamanalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StreamAnalyticsFunction provides resource knowledge for
// Microsoft.StreamAnalytics/streamingjobs/functions.
//
// One ARM type shared by two TF resources, discriminated by properties.type:
//   - azurerm_stream_analytics_function_javascript_udf  -> properties.type == "Scalar"
//   - azurerm_stream_analytics_function_javascript_uda  -> properties.type == "Aggregate"
// Only knowledge universal to BOTH shapes is unioned here; per-kind name validation
// (uda's validate.FunctionName regex) is NOT unioned as it would reject valid udf
// names, so the looser MinLength constraint is used for the name.
//
// Sources:
//   - terraform-provider-azurerm internal/services/streamanalytics/stream_analytics_function_javascript_udf_resource.go:24-163
//   - internal/services/streamanalytics/stream_analytics_function_javascript_uda_resource.go:25-162
//   - Timeouts (both, :40-45 / :41-46): Create 30m / Read 5m / Update 30m / Delete 30m
//   - go-azure-sdk resource-manager/streamanalytics/2020-03-01/functions
//     model_function.go / model_scalarfunctionproperties.go (type="Scalar") /
//     model_aggregatefunctionproperties.go (type="Aggregate") / model_functionconfiguration.go /
//     model_javascriptfunctionbinding.go (type="Microsoft.StreamAnalytics/JavascriptUdf") /
//     model_javascriptfunctionbindingproperties.go / model_functionoutput.go / model_functioninput.go
//   - ARM type casing verified via functions/id_function.go
//     StaticSegment("staticStreamingJobs", "streamingJobs") + StaticSegment("staticFunctions", "functions").
//   - uda name uses validate.FunctionName (regex ^[a-zA-Z0-9-]{3,63}$); not applied here
//     (kind-specific). input[*] is an array element path (dataType / isConfigurationParameter),
//     so its type enum is skipped (no declarative array-element form).
type StreamAnalyticsFunction struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StreamAnalyticsFunction)(nil)

// NewStreamAnalyticsFunction returns knowledge for the Stream Analytics function resource.
func NewStreamAnalyticsFunction() *StreamAnalyticsFunction {
	return &StreamAnalyticsFunction{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StreamAnalytics/streamingjobs/functions",
			ApiVersions:  []string{"2020-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Universal to both udf and uda: the JS script and the output data type are Required.
			RequiredFields: []string{
				"properties.properties.binding.properties.script",
				"properties.properties.output.dataType",
			},
			StringRules: []azwise.StringRule{
				// output.type -> output.dataType (identical allowed set for udf and uda).
				{
					PropertyPath:  "properties.properties.output.dataType",
					AllowedValues: []string{"any", "array", "bigint", "datetime", "float", "nvarchar(max)", "record"},
					Message:       "invalid function output type",
				},
				// name: universal minimum (udf: StringIsNotEmpty). uda's stricter
				// FunctionName regex is kind-specific and intentionally not unioned.
				{PropertyPath: "", MinLength: 1, Message: "function name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewStreamAnalyticsFunction()) }
