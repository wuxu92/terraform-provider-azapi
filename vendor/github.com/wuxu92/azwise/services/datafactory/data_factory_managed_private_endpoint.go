package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryManagedPrivateEndpoint provides resource knowledge for
// Microsoft.DataFactory/factories/managedVirtualNetworks/managedPrivateEndpoints.
//
// The ARM parent is the factory's managed virtual network: the Terraform ID is
// built as factories/<factory>/managedVirtualNetworks/<mvnet>/managedPrivateEndpoints/<name>
// (managedprivateendpoints.NewManagedPrivateEndpointID). AzureRM resolves the
// managedVirtualNetwork name automatically, but the ARM resource type includes it.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_managed_private_endpoint_resource.go:29-175
//     (schema: name/data_factory_id/target_resource_id/subresource_name/fqdns all ForceNew;
//     Create/Read/Delete timeouts, no Update; create mapping to ManagedPrivateEndpoint)
//   - terraform-provider-azurerm internal/services/datafactory/validate/datafactory.go:36-50
//     (DataFactoryManagedPrivateEndpointName regex)
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/managedprivateendpoints/model_managedprivateendpoint.go:6-13
//
// Intentionally skipped here:
//   - name: resource-name attribute (validated via StringRule with empty PropertyPath).
//   - data_factory_id: parent reference (AzAPI ID segment), not a body property.
//   - target_resource_id privatelinkservices/azure.ValidateResourceID and the
//     "subresource_name vs fqdns" mutual constraint: that check is a runtime
//     CustomizeDiff branch on whether the target is a Private Link Service, not a
//     single-field declarative rule; documented, not emitted.
type DataFactoryManagedPrivateEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryManagedPrivateEndpoint)(nil)

// NewDataFactoryManagedPrivateEndpoint returns knowledge for the
// Microsoft.DataFactory/factories/managedVirtualNetworks/managedPrivateEndpoints resource.
func NewDataFactoryManagedPrivateEndpoint() *DataFactoryManagedPrivateEndpoint {
	return &DataFactoryManagedPrivateEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/managedVirtualNetworks/managedPrivateEndpoints",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.privateLinkResourceId"}, // target_resource_id
				{PropertyPath: "properties.groupId"},               // subresource_name
				{PropertyPath: "properties.fqdns"},                 // fqdns
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath).
					Regex:   `^([[:alnum:]][-._[:alnum:]]{0,78}[_[:alnum:]])$`,
					Message: "invalid Data Factory Managed Private Endpoint name",
				},
			},
			RequiredFields: []string{
				"properties.privateLinkResourceId", // target_resource_id (Required)
			},
			ComputedFields: []string{
				"properties.connectionState",   // read-only approval status
				"properties.isReserved",        // read-only
				"properties.provisioningState", // read-only
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryManagedPrivateEndpoint()) }
