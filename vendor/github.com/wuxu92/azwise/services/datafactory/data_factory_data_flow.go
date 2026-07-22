package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryDataFlow provides resource knowledge for
// Microsoft.DataFactory/factories/dataflows.
//
// This ARM type is split by AzureRM into two typed Terraform resources,
// discriminated by the ARM body's properties.type field:
//   - azurerm_data_factory_data_flow         -> properties.type = "MappingDataFlow"
//   - azurerm_data_factory_flowlet_data_flow -> properties.type = "Flowlet"
//
// Both contributors share an identical schema shape (script/script_lines,
// source/sink/transformation blocks, description, folder). Only universal
// knowledge is unioned here; the type-specific payload lives inside the
// discriminated properties.typeProperties union, whose paths do not resolve
// through the SDK's interface-typed Properties field.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_data_flow_resource.go:30-99
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_flowlet_data_flow_resource.go:30-99
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/dataflows/model_dataflow.go:12-27
//     (DataFlow interface; MappingDataFlow / Flowlet / WranglingDataFlow)
//
// Intentionally skipped here:
//   - name: Required+ForceNew with no ValidateFunc in either resource — nothing to
//     express as a declarative StringRule.
//   - data_factory_id: parent reference (AzAPI ID segment), not a body property.
//   - script / script_lines AtLeastOneOf, source/sink/transformation blocks, folder:
//     all inside the discriminated properties.typeProperties union and validated with
//     StringIsNotEmpty only (not expressible declaratively).
type DataFactoryDataFlow struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryDataFlow)(nil)

// NewDataFactoryDataFlow returns merged knowledge for the
// Microsoft.DataFactory/factories/dataflows resource.
func NewDataFactoryDataFlow() *DataFactoryDataFlow {
	return &DataFactoryDataFlow{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/dataflows",
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
func init() { azwise.Register(NewDataFactoryDataFlow()) }
