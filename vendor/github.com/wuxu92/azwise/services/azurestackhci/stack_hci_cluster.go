package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCICluster provides resource knowledge for Microsoft.AzureStackHCI/clusters.
//
// Mirrors azurerm_stack_hci_cluster. name / resource_group_name / location live on the
// operational envelope; location is surfaced here as ForceNew because AzureRM marks
// commonschema.Location RequiresReplace and users benefit from the plan-time guardrail.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_cluster_resource.go:47-98
//     (schema: ForceNew, validators, Optional+Computed tenant_id, computed attributes)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_cluster_resource.go:122-140
//     (create mapping: client_id->properties.aadClientId, tenant_id->properties.aadTenantId)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_cluster_resource.go:205-238
//     (read/flatten mapping: cloud_id, service_endpoint, resource_provider_object_id)
//   - terraform-provider-azurerm internal/services/azurestackhci/validate/cluster_name.go:8-21
//     (ClusterName: non-empty, max 260 chars)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/clusters:
//     model_clusterproperties.go:12-33 (aadClientId, aadTenantId, cloudId, serviceEndpoint,
//     resourceProviderObjectId body/read-only paths)
//
// Not encoded (deliberate):
//   - client_id / tenant_id carry validation.IsUUID — a generic semantic validator that
//     azwise cannot express declaratively; it belongs in an azapin customizer (validators.UUID)
//     attached to properties.aadClientId / properties.aadTenantId, not in a StringRule.
//   - automanage_configuration_id is not part of the cluster body; AzureRM writes it as a
//     separate Automanage ConfigurationProfileHCIAssignment resource, so no ARM path here.
//   - identity is a system-assigned block (commonschema) with no cross-property constraints.
type StackHCICluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCICluster)(nil)

// NewStackHCICluster returns knowledge for the Azure Stack HCI clusters resource.
func NewStackHCICluster() *StackHCICluster {
	return &StackHCICluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/clusters",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.aadClientId"},
				{PropertyPath: "properties.aadTenantId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// validate.ClusterName: non-empty, must not exceed 260 characters.
					MinLength: 1,
					MaxLength: 260,
					Message:   "cannot be empty and must not exceed 260 characters",
				},
			},
			ComputedFields: []string{
				// Server-assigned read-only properties (AzureRM Computed-only attributes).
				"properties.cloudId",
				"properties.serviceEndpoint",
				"properties.resourceProviderObjectId",
			},
			DefaultValues: []azwise.DefaultValue{
				// tenant_id is Optional+Computed: AzureRM defaults it to the account tenant
				// when omitted, so no fixed ARM-format default is emitted.
				{PropertyPath: "properties.aadTenantId"},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCICluster()) }
