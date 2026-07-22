package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryIntegrationRuntime provides resource knowledge for
// Microsoft.DataFactory/factories/integrationruntimes.
//
// This ARM type is split by AzureRM into three typed Terraform resources,
// discriminated by the ARM body's properties.type field:
//   - azurerm_data_factory_integration_runtime_azure       -> properties.type = "Managed"
//   - azurerm_data_factory_integration_runtime_azure_ssis   -> properties.type = "Managed"
//   - azurerm_data_factory_integration_runtime_self_hosted -> properties.type = "SelfHosted"
//
// Only knowledge universal across all three contributors is unioned here.
// Type-specific value constraints live inside the discriminated properties.*
// union (typeProperties/computeProperties/ssisProperties/…), whose paths do not
// resolve through the SDK's interface-typed Properties field, so they are
// documented rather than emitted.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_integration_runtime_azure_resource.go:30-144
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_integration_runtime_azure_ssis_resource.go:54-210
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_integration_runtime_self_hosted_resource.go:51-113
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/integrationruntimes/model_integrationruntimeresource.go:13-20
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/integrationruntimes/model_integrationruntime.go:12-25
//     (IntegrationRuntime interface; ManagedIntegrationRuntime / SelfHostedIntegrationRuntime)
//
// Intentionally skipped here:
//   - name: the three contributors use *different* name regexes — azure and
//     azure_ssis use `^([a-zA-Z0-9](-|-?[a-zA-Z0-9]+)+[a-zA-Z0-9])$` (min 3, no
//     consecutive dashes), self_hosted uses `^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`.
//     A single StringRule would reject valid names for one contributor, so no
//     universal name rule is emitted.
//   - data_factory_id: parent reference (AzAPI ID segment), not a body property.
//   - Type-specific fields, all inside the discriminated union under properties.*
//     and therefore non-resolving through the interface Properties field:
//       * Managed (azure): location (ForceNew), compute_type (DataFlowComputeType
//         enum {General,ComputeOptimized,MemoryOptimized}, default General),
//         core_count IntInSlice{8,16,32,48,80,144,272} (default 8), time_to_live_min,
//         cleanup_enabled, virtual_network_enabled (ForceNew).
//       * Managed (azure_ssis): node_size StringInSlice, number_of_nodes IntBetween(1,10),
//         max_parallel_executions_per_node IntBetween(1,16), edition
//         {Standard,Enterprise}, license_type {LicenseIncluded,BasePrice},
//         catalog/express/proxy custom setup blocks.
//       * SelfHosted: rbac_authorization (ForceNew), self_contained_interactive_authoring.
type DataFactoryIntegrationRuntime struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryIntegrationRuntime)(nil)

// NewDataFactoryIntegrationRuntime returns merged knowledge for the
// Microsoft.DataFactory/factories/integrationruntimes resource.
func NewDataFactoryIntegrationRuntime() *DataFactoryIntegrationRuntime {
	return &DataFactoryIntegrationRuntime{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/integrationruntimes",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryIntegrationRuntime()) }
