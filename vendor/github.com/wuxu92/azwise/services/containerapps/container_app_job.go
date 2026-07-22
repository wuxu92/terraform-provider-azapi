package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerAppJob provides resource knowledge for Microsoft.App/jobs.
//
// Mirrors azurerm_container_app_job.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_job_resource.go
//     schema (Arguments 66-204, Attributes 206-221) + Create (223-305): environment/trigger
//     ForceNew, replica timeout/retry ranges, trigger parallelism/completion defaults,
//     triggerType hardcoded per trigger block; timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/jobs
//     JobProperties (configuration, environmentId, template, workloadProfileName, read-only
//     eventStreamEndpoint/outboundIpAddresses/provisioningState), JobConfiguration
//     (replicaTimeout/replicaRetryLimit/triggerType + event/schedule/manual trigger configs),
//     constants.go TriggerType enum.
type ContainerAppJob struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerAppJob)(nil)

func NewContainerAppJob() *ContainerAppJob {
	return &ContainerAppJob{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/jobs",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// container_app_environment_id and the three trigger config blocks are ForceNew.
			// The trigger blocks are ExactlyOneOf, so changing the trigger type replaces
			// the job.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.environmentId"},
				{PropertyPath: "properties.configuration.eventTriggerConfig"},
				{PropertyPath: "properties.configuration.scheduleTriggerConfig"},
				{PropertyPath: "properties.configuration.manualTriggerConfig"},
			},
			RequiredFields: []string{
				"properties.environmentId",
				"properties.configuration.replicaTimeout",
				"properties.configuration.triggerType",
			},
			StringRules: []azwise.StringRule{
				// AzureRM sets triggerType based on which *_trigger_config block is present.
				{
					PropertyPath:  "properties.configuration.triggerType",
					AllowedValues: []string{"Event", "Manual", "Schedule"},
					Message:       "trigger type must be Event, Manual or Schedule",
				},
			},
			IntRules: []azwise.IntRule{
				// replica_timeout_in_seconds -> validation.IntAtLeast(1).
				{
					PropertyPath: "properties.configuration.replicaTimeout",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "replica timeout must be at least 1 second",
				},
				// replica_retry_limit -> validation.IntAtLeast(0).
				{
					PropertyPath: "properties.configuration.replicaRetryLimit",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "replica retry limit must be at least 0",
				},
				// parallelism / replica_completion_count -> validation.IntAtLeast(1).
				{PropertyPath: "properties.configuration.eventTriggerConfig.parallelism", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.configuration.eventTriggerConfig.replicaCompletionCount", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.configuration.scheduleTriggerConfig.parallelism", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.configuration.scheduleTriggerConfig.replicaCompletionCount", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.configuration.manualTriggerConfig.parallelism", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.configuration.manualTriggerConfig.replicaCompletionCount", MinValue: azwise.Ptr(int64(1))},
			},
			// Read-only properties populated by Azure in the GET response.
			ComputedFields: []string{
				"properties.outboundIpAddresses",
				"properties.eventStreamEndpoint",
				"properties.provisioningState",
			},
			// AzureRM schema defaults for the trigger sub-objects (only validated when the
			// corresponding trigger block is present).
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.configuration.eventTriggerConfig.parallelism", Value: float64(1)},
				{PropertyPath: "properties.configuration.eventTriggerConfig.replicaCompletionCount", Value: float64(1)},
				{PropertyPath: "properties.configuration.scheduleTriggerConfig.parallelism", Value: float64(1)},
				{PropertyPath: "properties.configuration.scheduleTriggerConfig.replicaCompletionCount", Value: float64(1)},
				{PropertyPath: "properties.configuration.manualTriggerConfig.parallelism", Value: float64(1)},
				{PropertyPath: "properties.configuration.manualTriggerConfig.replicaCompletionCount", Value: float64(1)},
			},
		},
	}
}

func init() { azwise.Register(NewContainerAppJob()) }
