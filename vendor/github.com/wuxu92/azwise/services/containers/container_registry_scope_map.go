package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryScopeMap provides resource knowledge for
// Microsoft.ContainerRegistry/registries/scopeMaps.
//
// Contributing Terraform resource: azurerm_container_registry_scope_map.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_scope_map_resource.go
//     (schema L44-76, Create L101-106)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_scope_map_name.go
//   - go-azure-sdk resource-manager/containerregistry/2025-11-01/scopemaps:
//     model_scopemapproperties.go.
//
// Notes:
//   - actions (Required, MinItems 1) maps to properties.actions.
//   - description maps to properties.description (length 1-256).
type ContainerRegistryScopeMap struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryScopeMap)(nil)

// NewContainerRegistryScopeMap returns knowledge for the scopeMaps resource.
func NewContainerRegistryScopeMap() *ContainerRegistryScopeMap {
	return &ContainerRegistryScopeMap{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/scopeMaps",
			ApiVersions:  []string{"2025-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryScopeMapName
				{
					Regex:     `^[a-zA-Z0-9\-]+$`,
					MinLength: 5,
					MaxLength: 49,
					Message:   "alpha numeric characters and hyphens only, 5-49 chars",
				},
				// ── description → properties.description
				{
					PropertyPath: "properties.description",
					MinLength:    1,
					MaxLength:    256,
					Message:      "description must be 1-256 characters",
				},
			},
			// actions is Required for creation.
			RequiredFields: []string{"properties.actions"},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.creationDate",
				"properties.provisioningState",
				"properties.type",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistryScopeMap()) }
