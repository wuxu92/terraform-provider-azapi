package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryPipeline provides resource knowledge for
// Microsoft.DataFactory/factories/pipelines.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_pipeline_resource.go:26-116
//     (schema: name/data_factory_id ForceNew, timeouts, concurrency IntBetween(1,50),
//     folder/monitor_metrics_after_duration StringIsNotEmpty)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_pipeline_resource.go:118-216
//     (create mapping to pipelines.PipelineResource / pipelines.Pipeline:
//     description, parameters, variables, activities, annotations, concurrency,
//     policy.elapsedTimeMetric.duration, folder.name)
//   - terraform-provider-azurerm internal/services/datafactory/validate/datafactory.go:14-23
//     (DataFactoryPipelineAndTriggerName regex)
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/pipelines/model_pipelineresource.go:10-17
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/pipelines/model_pipeline.go:11-21
//
// Intentionally skipped here:
//   - name: resource-name attribute (validated via StringRule with empty PropertyPath),
//     not a Microsoft.DataFactory/factories/pipelines body property.
//   - data_factory_id: parent reference (AzAPI ID segment), not a body property.
//   - folder StringIsNotEmpty / monitor_metrics_after_duration StringIsNotEmpty:
//     non-empty checks are not expressible as a declarative StringRule (no regex/enum/length).
type DataFactoryPipeline struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryPipeline)(nil)

// NewDataFactoryPipeline returns knowledge for the
// Microsoft.DataFactory/factories/pipelines resource.
func NewDataFactoryPipeline() *DataFactoryPipeline {
	return &DataFactoryPipeline{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/pipelines",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). Data Factory pipeline naming rule.
					Regex:   `^[A-Za-z0-9_][^<>*#.%&:\\+?/]*$`,
					Message: "invalid Data Factory pipeline name, see https://docs.microsoft.com/en-us/azure/data-factory/naming-rules",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.concurrency",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(50)),
					Message:      "concurrency must be between 1 and 50",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryPipeline()) }
