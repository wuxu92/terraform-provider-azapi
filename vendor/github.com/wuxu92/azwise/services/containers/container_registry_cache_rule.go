package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryCacheRule provides resource knowledge for
// Microsoft.ContainerRegistry/registries/cacheRules.
//
// Contributing Terraform resource: azurerm_container_registry_cache_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_cache_rule_resource.go
//     (schema L27-62, Create L84-143)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_cache_rule_name.go
//   - go-azure-sdk resource-manager/containerregistry/2023-07-01/cacherules:
//     model_cacheruleproperties.go.
//
// Notes:
//   - source_repo (Required, ForceNew) maps to properties.sourceRepository.
//   - target_repo (Required, ForceNew) maps to properties.targetRepository.
//   - credential_set_id maps to properties.credentialSetResourceId.
type ContainerRegistryCacheRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryCacheRule)(nil)

// NewContainerRegistryCacheRule returns knowledge for the cacheRules resource.
func NewContainerRegistryCacheRule() *ContainerRegistryCacheRule {
	return &ContainerRegistryCacheRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/cacheRules",
			ApiVersions:  []string{"2023-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// source_repo and target_repo are ForceNew body properties.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sourceRepository"},
				{PropertyPath: "properties.targetRepository"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryCacheRuleName
				{
					Regex:     `^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`,
					MinLength: 5,
					MaxLength: 49,
					Message:   "alpha numeric characters optionally separated by '-', 5-49 chars",
				},
			},
			// source_repo and target_repo are Required for creation.
			RequiredFields: []string{
				"properties.sourceRepository",
				"properties.targetRepository",
			},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.creationDate",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistryCacheRule()) }
