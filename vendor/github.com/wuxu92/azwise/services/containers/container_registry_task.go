package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryTask provides resource knowledge for
// Microsoft.ContainerRegistry/registries/tasks.
//
// Contributing Terraform resource: azurerm_container_registry_task.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_task_resource.go
//     (schema L154-618, Create/expand L667-746, expand helpers L968-1597)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_task_name.go
//   - go-azure-sdk resource-manager/containerregistry/2019-06-01-preview/tasks:
//     model_taskproperties.go, model_platformproperties.go, model_triggerproperties.go,
//     model_baseimagetrigger.go, model_agentproperties.go, constants.go (enums).
//
// Notes:
//   - enabled (bool) maps to properties.status (Enabled/Disabled), default true → Enabled.
//   - is_system_task is ForceNew → properties.isSystemTask (default false).
//   - A task requires exactly one build step (docker_step/file_step/encoded_step) which
//     expands into properties.step (a polymorphic oneOf) — not expressible as a single
//     required path, left as a note.
//   - source_trigger and timer_trigger expand into properties.trigger.sourceTriggers[*] /
//     timerTriggers[*] (arrays); their nested enum validators (events, source_type,
//     token_type) are array-element paths, skipped per azwise policy.
//   - registry_credential.source.login_mode and .custom.* are array/sub-object credential
//     paths, skipped.
type ContainerRegistryTask struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryTask)(nil)

// NewContainerRegistryTask returns knowledge for the tasks resource.
func NewContainerRegistryTask() *ContainerRegistryTask {
	return &ContainerRegistryTask{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/tasks",
			ApiVersions:  []string{"2019-06-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// is_system_task is a ForceNew body property.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.isSystemTask"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryTaskName
				{
					Regex:   `^[\w-]*$`,
					Message: "only alpha numeric characters (optionally separated by dash) are allowed",
				},
				// ── platform.os → properties.platform.os
				{
					PropertyPath:  "properties.platform.os",
					AllowedValues: []string{"Windows", "Linux"},
					Message:       "must be one of Windows or Linux",
				},
				// ── platform.architecture → properties.platform.architecture
				{
					PropertyPath:  "properties.platform.architecture",
					AllowedValues: []string{"amd64", "arm", "arm64", "386", "x86"},
					Message:       "must be one of amd64, arm, arm64, 386 or x86",
				},
				// ── platform.variant → properties.platform.variant
				{
					PropertyPath:  "properties.platform.variant",
					AllowedValues: []string{"v6", "v7", "v8"},
					Message:       "must be one of v6, v7 or v8",
				},
				// ── enabled → properties.status
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				// ── base_image_trigger.type → properties.trigger.baseImageTrigger.baseImageTriggerType
				{
					PropertyPath:  "properties.trigger.baseImageTrigger.baseImageTriggerType",
					AllowedValues: []string{"All", "Runtime"},
					Message:       "must be one of All or Runtime",
				},
				// ── base_image_trigger.update_trigger_payload_type
				{
					PropertyPath:  "properties.trigger.baseImageTrigger.updateTriggerPayloadType",
					AllowedValues: []string{"Default", "Token"},
					Message:       "must be one of Default or Token",
				},
			},
			IntRules: []azwise.IntRule{
				// ── timeout_in_seconds → properties.timeout
				{
					PropertyPath: "properties.timeout",
					MinValue:     azwise.Ptr(int64(300)),
					MaxValue:     azwise.Ptr(int64(28800)),
				},
				// ── agent_setting.cpu → properties.agentConfiguration.cpu (IntInSlice{2})
				{
					PropertyPath: "properties.agentConfiguration.cpu",
					MinValue:     azwise.Ptr(int64(2)),
					MaxValue:     azwise.Ptr(int64(2)),
					Message:      "cpu must be 2",
				},
			},
			// agent_pool_name conflicts with agent_setting.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.agentPoolName", "properties.agentConfiguration"},
					Message: "agent_pool_name cannot be combined with agent_setting",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default true → Enabled.
				{PropertyPath: "properties.status", Value: "Enabled"},
				// is_system_task Default false.
				{PropertyPath: "properties.isSystemTask", Value: false},
				// timeout_in_seconds Default 3600.
				{PropertyPath: "properties.timeout", Value: int64(3600)},
			},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.creationDate",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistryTask()) }
