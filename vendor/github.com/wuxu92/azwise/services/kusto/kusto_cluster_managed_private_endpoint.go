package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoClusterManagedPrivateEndpoint provides resource knowledge for
// Microsoft.Kusto/clusters/managedPrivateEndpoints.
//
// Mirrors azurerm_kusto_cluster_managed_private_endpoint. name / cluster_name /
// resource_group_name are envelope-owned; name is ForceNew (RequiresReplace).
// The ARM body carries no `location` field (it inherits the parent cluster's).
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_managed_private_endpoint_resource.go:26-96
//     (schema: ForceNew name/cluster_name/private_link_resource_id/group_id/
//     private_link_resource_region; request_message Optional; timeouts)
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_managed_private_endpoint_resource.go:120-133
//     (expand: groupId, privateLinkResourceId, privateLinkResourceRegion, requestMessage)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/managedprivateendpoints:
//     model_managedprivateendpoint.go:10-16 (envelope: name, properties),
//     model_managedprivateendpointproperties.go:6-12 (properties.* body paths),
//     id_managedprivateendpoint.go:110-127 (ARM path casing:
//     clusters/managedPrivateEndpoints)
//
// Not encoded (deliberate):
//   - name (validation.StringIsNotEmpty), group_id (StringIsNotEmpty),
//     private_link_resource_region (StringIsNotEmpty), request_message
//     (StringIsNotEmpty): non-empty-string checks add no declarative constraint
//     beyond presence and are not emitted as StringRules.
//   - private_link_resource_id (azure.ValidateResourceID): a generic "is an ARM
//     resource id" semantic validator; it maps to properties.privateLinkResourceId
//     but is not a declarative enum/regex/length rule, so it belongs in an azapin
//     customizer (schema/validators.AzureResourceID) rather than here.
//   - provisioningState is server-computed read-only (see ComputedFields).
type KustoClusterManagedPrivateEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoClusterManagedPrivateEndpoint)(nil)

// NewKustoClusterManagedPrivateEndpoint returns knowledge for the
// clusters/managedPrivateEndpoints resource.
func NewKustoClusterManagedPrivateEndpoint() *KustoClusterManagedPrivateEndpoint {
	return &KustoClusterManagedPrivateEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/managedPrivateEndpoints",
			ApiVersions:  []string{"2025-02-14"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.groupId"},
				{PropertyPath: "properties.privateLinkResourceId"},
				{PropertyPath: "properties.privateLinkResourceRegion"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.groupId",
				"properties.privateLinkResourceId",
			},
		},
	}
}

func init() { azwise.Register(NewKustoClusterManagedPrivateEndpoint()) }
