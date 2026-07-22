package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryAgentPool provides resource knowledge for
// Microsoft.ContainerRegistry/registries/agentPools.
//
// Contributing Terraform resource: azurerm_container_registry_agent_pool.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_agent_pool_resource.go
//     (schema L46-92, Create L117-131)
//   - go-azure-sdk resource-manager/containerregistry/2019-06-01-preview/agentpools:
//     model_agentpoolproperties.go.
//
// Notes:
//   - instance_count maps to properties.count (default 1).
//   - tier is ForceNew → properties.tier (default S1); AzureRM restricts to S1/S2/S3/I6
//     though the ARM field is a free-form string.
//   - virtual_network_subnet_id is ForceNew → properties.virtualNetworkSubnetResourceId.
type ContainerRegistryAgentPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryAgentPool)(nil)

// NewContainerRegistryAgentPool returns knowledge for the agentPools resource.
func NewContainerRegistryAgentPool() *ContainerRegistryAgentPool {
	return &ContainerRegistryAgentPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/agentPools",
			ApiVersions:  []string{"2019-06-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// tier and virtual_network_subnet_id are ForceNew body properties.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.tier"},
				{PropertyPath: "properties.virtualNetworkSubnetResourceId"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringLenBetween(3, 20)
				{
					MinLength: 3,
					MaxLength: 20,
					Message:   "name must be 3-20 characters",
				},
				// ── tier → properties.tier
				{
					PropertyPath:  "properties.tier",
					AllowedValues: []string{"S1", "S2", "S3", "I6"},
					Message:       "must be one of S1, S2, S3 or I6",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// instance_count Default 1.
				{PropertyPath: "properties.count", Value: int64(1)},
				// tier Default S1.
				{PropertyPath: "properties.tier", Value: "S1"},
			},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistryAgentPool()) }
