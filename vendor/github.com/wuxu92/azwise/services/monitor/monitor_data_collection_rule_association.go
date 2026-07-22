package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataCollectionRuleAssociation provides resource knowledge for
// Microsoft.Insights/dataCollectionRuleAssociations.
//
// Mirrors azurerm_monitor_data_collection_rule_association. This is a scoped
// (extension) proxy resource attached to a target resource; its ID has the form
// /{scope}/providers/Microsoft.Insights/dataCollectionRuleAssociations/{name}.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_data_collection_rule_association_resource.go
//     Arguments (32-70): target_resource_id Required+ForceNew (scope, not a body property);
//     name Optional+ForceNew Default "configurationAccessEndpoint" (StringIsNotEmpty);
//     data_collection_endpoint_id / data_collection_rule_id each Optional, ExactlyOneOf
//     (rule_id RequiredWith name) → properties.dataCollectionEndpointId /
//     properties.dataCollectionRuleId; description Optional. Create/Update/Delete 30m,
//     Read 5m.
//   - go-azure-sdk resource-manager/insights/2023-03-11/datacollectionruleassociations:
//     DataCollectionRuleAssociation (model_datacollectionruleassociation.go) fields
//     metadata/provisioningState read-only; id_scopeddatacollectionruleassociation.go
//     type Microsoft.Insights/dataCollectionRuleAssociations.
//
// Note: data_collection_endpoint_id vs data_collection_rule_id is an ExactlyOneOf
// relational constraint that azwise cannot express declaratively; enforced by the
// service, documented here only.
type DataCollectionRuleAssociation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataCollectionRuleAssociation)(nil)

// NewDataCollectionRuleAssociation returns knowledge for the
// dataCollectionRuleAssociations resource.
func NewDataCollectionRuleAssociation() *DataCollectionRuleAssociation {
	return &DataCollectionRuleAssociation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/dataCollectionRuleAssociations",
			ApiVersions:  []string{"2023-03-11"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: StringIsNotEmpty (defaults to "configurationAccessEndpoint").
					MinLength: 1,
					Message:   "name must not be empty",
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.metadata",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewDataCollectionRuleAssociation()) }
