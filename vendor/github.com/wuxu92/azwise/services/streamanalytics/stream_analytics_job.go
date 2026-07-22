package streamanalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StreamAnalyticsJob provides resource knowledge for
// Microsoft.StreamAnalytics/streamingjobs.
//
// Mirrors azurerm_stream_analytics_job. The sibling TF resources
// azurerm_stream_analytics_job_schedule (POSTs start/stop actions on the job — a
// pure operation with no distinct ARM resource body) and
// azurerm_stream_analytics_job_storage_account (mutates the job's
// properties.jobStorageAccount sub-object via the same streamingjobs API) do not
// introduce new ARM resource types; their body knowledge is folded here
// (jobStorageAccount enums/defaults/sensitive fields) or skipped (schedule = action).
//
// Sources:
//   - terraform-provider-azurerm internal/services/streamanalytics/stream_analytics_job_resource.go:56-342
//   - internal/services/streamanalytics/stream_analytics_job_storage_account_resource.go (jobStorageAccount sub-object)
//   - Timeouts (:49-54): Create 30m / Read 5m / Update 30m / Delete 30m
//   - go-azure-sdk resource-manager/streamanalytics/2021-10-01-preview/streamingjobs
//     model_streamingjobproperties.go / model_jobstorageaccount.go /
//     model_transformationproperties.go / constants.go
//   - ARM type casing verified via streamingjobs/id_streamingjob.go
//     StaticSegment("staticStreamingJobs", "streamingJobs").
//   - compatibility_level and sku_name include values AzureRM accepts beyond the SDK
//     constant set ("1.1"; "StandardV2") because AzAPI forwards raw ARM values.
//   - Semantic CustomizeDiff (account_key forbidden when authentication_mode == Msi)
//     is a cross-field check with no declarative azwise form — routed to a customizer.
type StreamAnalyticsJob struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StreamAnalyticsJob)(nil)

// NewStreamAnalyticsJob returns knowledge for the Stream Analytics streaming job resource.
func NewStreamAnalyticsJob() *StreamAnalyticsJob {
	return &StreamAnalyticsJob{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StreamAnalytics/streamingjobs",
			ApiVersions:  []string{"2021-10-01-preview", "2020-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// type -> properties.jobType is ForceNew. name/location/resource_group are envelope.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.jobType"},
			},
			// transformation_query is Required (inline transformation on create).
			RequiredFields: []string{
				"properties.transformation.properties.query",
			},
			// job_id is Computed-only (server-assigned).
			ComputedFields: []string{
				"properties.jobId",
			},
			SensitiveFields: []string{
				"properties.jobStorageAccount.accountKey",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.dataLocale", Value: "en-US"},
				{PropertyPath: "properties.eventsLateArrivalMaxDelayInSeconds", Value: float64(5)},
				{PropertyPath: "properties.eventsOutOfOrderMaxDelayInSeconds", Value: float64(0)},
				{PropertyPath: "properties.eventsOutOfOrderPolicy", Value: "Adjust"},
				{PropertyPath: "properties.outputErrorPolicy", Value: "Drop"},
				{PropertyPath: "properties.jobType", Value: "Cloud"},
				{PropertyPath: "properties.contentStoragePolicy", Value: "SystemAccount"},
				{PropertyPath: "properties.sku.name", Value: "Standard"},
				// job_storage_account.authentication_mode default.
				{PropertyPath: "properties.jobStorageAccount.authenticationMode", Value: "ConnectionString"},
			},
			IntRules: []azwise.IntRule{
				// events_late_arrival_max_delay_in_seconds: IntBetween(-1, 1814399).
				{PropertyPath: "properties.eventsLateArrivalMaxDelayInSeconds", MinValue: azwise.Ptr(int64(-1)), MaxValue: azwise.Ptr(int64(1814399))},
				// events_out_of_order_max_delay_in_seconds: IntBetween(0, 599).
				{PropertyPath: "properties.eventsOutOfOrderMaxDelayInSeconds", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(599))},
				// streaming_units: IntBetween(1, 120).
				{PropertyPath: "properties.transformation.properties.streamingUnits", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(120))},
			},
			StringRules: []azwise.StringRule{
				// compatibility_level (SDK: 1.0/1.2; AzureRM also accepts 1.1).
				{PropertyPath: "properties.compatibilityLevel", AllowedValues: []string{"1.0", "1.1", "1.2"}, Message: "invalid compatibility_level"},
				// events_out_of_order_policy.
				{PropertyPath: "properties.eventsOutOfOrderPolicy", AllowedValues: []string{"Adjust", "Drop"}, Message: "invalid events_out_of_order_policy"},
				// output_error_policy.
				{PropertyPath: "properties.outputErrorPolicy", AllowedValues: []string{"Drop", "Stop"}, Message: "invalid output_error_policy"},
				// type -> jobType.
				{PropertyPath: "properties.jobType", AllowedValues: []string{"Cloud", "Edge"}, Message: "invalid job type"},
				// content_storage_policy.
				{PropertyPath: "properties.contentStoragePolicy", AllowedValues: []string{"SystemAccount", "JobStorageAccount"}, Message: "invalid content_storage_policy"},
				// sku_name (SDK: Standard; AzureRM also accepts StandardV2).
				{PropertyPath: "properties.sku.name", AllowedValues: []string{"Standard", "StandardV2"}, Message: "invalid sku_name"},
				// job_storage_account.authentication_mode (full SDK AuthenticationMode set).
				{PropertyPath: "properties.jobStorageAccount.authenticationMode", AllowedValues: []string{"ConnectionString", "Msi", "UserToken"}, Message: "invalid authentication_mode"},
				// name: validation.StringIsNotEmpty.
				{PropertyPath: "", MinLength: 1, Message: "job name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewStreamAnalyticsJob()) }
