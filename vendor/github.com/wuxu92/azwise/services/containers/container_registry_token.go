package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryToken provides resource knowledge for
// Microsoft.ContainerRegistry/registries/tokens.
//
// Contributing Terraform resource: azurerm_container_registry_token.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_token_resource.go
//     (schema L44-72, Create L100-113)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_token_name.go
//   - go-azure-sdk resource-manager/containerregistry/2025-11-01/tokens:
//     model_tokenproperties.go, constants.go (TokenStatus).
//
// Notes:
//   - enabled (bool) maps to properties.status (enum enabled/disabled), default true → enabled.
//   - scope_map_id (Required) maps to properties.scopeMapId.
//   - token passwords are managed via a separate generateCredentials data-plane flow
//     (azurerm_container_registry_token_password) and are not part of this create body.
type ContainerRegistryToken struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryToken)(nil)

// NewContainerRegistryToken returns knowledge for the tokens resource.
func NewContainerRegistryToken() *ContainerRegistryToken {
	return &ContainerRegistryToken{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/tokens",
			ApiVersions:  []string{"2025-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryTokenName
				{
					Regex:     `^[a-zA-Z][a-zA-Z0-9-]{4,48}$`,
					MinLength: 5,
					MaxLength: 50,
					Message:   "alpha numeric characters (optionally separated by dash), 5-50 chars, starting with a letter",
				},
				// ── enabled → properties.status
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"enabled", "disabled"},
					Message:       "must be one of enabled or disabled",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default true → enabled.
				{PropertyPath: "properties.status", Value: "enabled"},
			},
			// scope_map_id is Required for creation.
			RequiredFields: []string{"properties.scopeMapId"},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.creationDate",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistryToken()) }
