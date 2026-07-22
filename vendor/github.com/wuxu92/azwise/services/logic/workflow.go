package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Workflow provides resource knowledge for Microsoft.Logic/workflows.
//
// Contributing Terraform resource: azurerm_logic_app_workflow.
//
// Folded into this parent (NOT distinct ARM types — they live inside
// properties.definition.triggers / properties.definition.actions of the workflow
// definition JSON, so no separate knowledge file is emitted for them):
//   - azurerm_logic_app_trigger_custom, azurerm_logic_app_trigger_http_request,
//     azurerm_logic_app_trigger_recurrence
//   - azurerm_logic_app_action_custom, azurerm_logic_app_action_http
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_workflow_resource.go
//     (schema L52-276, Create body L301-361, Read/computed L498-569)
//   - go-azure-sdk resource-manager/logic/2019-05-01/workflows:
//     model_workflow.go, model_workflowproperties.go, model_resourcereference.go,
//     constants.go (PossibleValuesForWorkflowState)
//
// Notes:
//   - name/location/resource_group_name are envelope-owned and ForceNew; not emitted
//     as body rules. The name regex is emitted against the resource name (PropertyPath "").
//   - workflow_schema and workflow_version are ForceNew and are folded into the opaque
//     properties.definition JSON blob ($schema / contentVersion keys); emitted as
//     ForceNew paths so a change is flagged, but the definition body is otherwise freeform.
//   - `enabled` (bool, default true) maps to properties.state (Enabled/Disabled); the full
//     WorkflowState SDK enum is emitted, with a DefaultValue of "Enabled".
//   - integration_service_environment_id / logic_app_integration_account_id are Azure
//     resource-ID references (properties.integrationServiceEnvironment.id /
//     properties.integrationAccount.id); their ID-format validation is semantic and belongs
//     in an azapin customizer (validators.AzureResourceID), not a declarative StringRule.
//   - access_endpoint, connector/workflow endpoint+outbound IP addresses, provisioningState,
//     version and created/changed times are server-computed read-only outputs.
type Workflow struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Workflow)(nil)

// NewWorkflow returns knowledge for the workflows resource.
func NewWorkflow() *Workflow {
	return &Workflow{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/workflows",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.integrationServiceEnvironment.id"},
				{PropertyPath: "properties.definition.$schema"},
				{PropertyPath: "properties.definition.contentVersion"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[-()_.A-Za-z0-9]{1,80}$`,
					Message:      "The Logic app name can contain only letters, numbers, periods (.), hyphens (-), brackets (()) and underscores (_), up to 80 characters",
				},
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"Completed", "Deleted", "Disabled", "Enabled", "NotSpecified", "Suspended"},
					Message:       "workflow state (from `enabled`) must be a valid WorkflowState",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.state", Value: "Enabled"},
			},
			ComputedFields: []string{
				"properties.accessEndpoint",
				"properties.endpointsConfiguration",
				"properties.provisioningState",
				"properties.version",
				"properties.createdTime",
				"properties.changedTime",
			},
		},
	}
}

func init() { azwise.Register(NewWorkflow()) }
