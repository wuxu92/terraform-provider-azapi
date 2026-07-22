package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHubCluster provides resource knowledge for Microsoft.EventHub/clusters.
//
// Contributing Terraform resource: azurerm_eventhub_cluster.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_cluster_resource.go
//     (schema L29-73, Create body L98-103, expand sku L179-188)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/eventhubsclusters:
//     model_cluster.go, model_clustersku.go, constants.go (ClusterSkuName)
//
// Notes:
//   - name/location/resource_group_name are envelope-owned and ForceNew; not emitted as
//     body rules.
//   - sku_name is a composite Terraform field ("Dedicated_N") expanded into sku.name
//     (enum "Dedicated") and sku.capacity (int); the "^Dedicated_[1-9][0-9]*$" regex
//     applies to that composite representation, so only the sku.name enum is emitted.
//   - a cluster cannot be deleted until 4 hours after creation, hence the 300m delete.
type EventHubCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHubCluster)(nil)

// NewEventHubCluster returns knowledge for the clusters resource.
func NewEventHubCluster() *EventHubCluster {
	return &EventHubCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/clusters",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 300 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,48}[a-zA-Z0-9])?$`,
					Message:      "cluster name may contain only letters, numbers, periods, hyphens and underscores, up to 50 characters, beginning and ending with a letter or number",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Dedicated"},
					Message:       "sku name must be Dedicated",
				},
			},
			RequiredFields: []string{
				"sku.name",
			},
		},
	}
}

func init() { azwise.Register(NewEventHubCluster()) }
