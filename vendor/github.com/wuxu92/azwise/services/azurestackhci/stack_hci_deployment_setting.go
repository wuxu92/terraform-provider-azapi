package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCIDeploymentSetting provides resource knowledge for
// Microsoft.AzureStackHCI/clusters/deploymentSettings.
//
// Mirrors azurerm_stack_hci_deployment_setting. stack_hci_cluster_id is the parent scope
// (envelope); the deployment setting is always created with the fixed child name "default".
// Every top-level argument on this resource is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_deployment_setting_resource.go:157-260
//     (top-level schema: stack_hci_cluster_id, arc_resource_ids, version, scale_unit)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_deployment_setting_resource.go:709-763
//     (create mapping: arcNodeResourceIds, deploymentMode, deploymentConfiguration.version/scaleUnits)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/deploymentsettings:
//     model_deploymentsettingsproperties.go:6-12 (arcNodeResourceIds, deploymentConfiguration,
//     deploymentMode, provisioningState, reportedProperties),
//     model_deploymentconfiguration.go:6-9 (scaleUnits, version)
//
// Not encoded (deliberate):
//   - stack_hci_cluster_id (commonschema.ResourceIDReferenceRequiredForceNew) is the parent
//     scope; its resource-ID validator belongs in an azapin customizer, not a StringRule.
//   - arc_resource_ids is a list of machines.ValidateMachineID elements on
//     properties.arcNodeResourceIds[*]; azwise cannot resolve a path through an array element,
//     so the per-element resource-ID rule is documented, not emitted.
//   - The entire scale_unit block (cluster name/azure_service_endpoint/cloud_account_name/
//     witness_type enum/witness_path, domain_fqdn regex, host_network intents/storage networks,
//     infrastructure_network ip_pool, optional_service, physical_node, storage
//     configuration_mode, and the observability/security boolean settings) lives under the
//     properties.deploymentConfiguration.scaleUnits[*] array element. azwise/azapin cannot
//     lower or resolve a path through an array element, so all of it is skipped here
//     (documented, not emitted).
//   - scale_unit MinItems:1 cannot be expressed (ArrayRule supports MaxItems only).
//   - deploymentMode is hard-coded by AzureRM (Validate then Deploy); it is a Required ARM
//     field with no user-facing default, listed in RequiredFields.
type StackHCIDeploymentSetting struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCIDeploymentSetting)(nil)

// NewStackHCIDeploymentSetting returns knowledge for the Azure Stack HCI deploymentSettings resource.
func NewStackHCIDeploymentSetting() *StackHCIDeploymentSetting {
	return &StackHCIDeploymentSetting{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/clusters/deploymentSettings",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.arcNodeResourceIds"},
				{PropertyPath: "properties.deploymentConfiguration.version"},
				{PropertyPath: "properties.deploymentConfiguration.scaleUnits"},
			},
			RequiredFields: []string{
				"properties.arcNodeResourceIds",
				"properties.deploymentMode",
				"properties.deploymentConfiguration.version",
				"properties.deploymentConfiguration.scaleUnits",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 6 * time.Hour,
				Read:   5 * time.Minute,
				Delete: 1 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.deploymentConfiguration.version",
					Regex:        `^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$`,
					Message:      "the version must be a set of numbers separated by dots, for example `10.0.0.1`",
				},
			},
			ComputedFields: []string{
				// Server-assigned read-only properties (present in GET, not settable).
				"properties.provisioningState",
				"properties.reportedProperties",
			},
		},
	}
}

func init() { azwise.Register(NewStackHCIDeploymentSetting()) }
