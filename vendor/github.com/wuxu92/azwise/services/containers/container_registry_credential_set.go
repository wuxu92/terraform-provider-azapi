package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryCredentialSet provides resource knowledge for
// Microsoft.ContainerRegistry/registries/credentialSets.
//
// Contributing Terraform resource: azurerm_container_registry_credential_set.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_credential_set_resource.go
//     (schema L28-77, Create L104-160)
//   - go-azure-sdk resource-manager/containerregistry/2023-07-01/credentialsets:
//     model_credentialsetproperties.go, model_authcredential.go.
//
// Notes:
//   - login_server (Required, ForceNew) maps to properties.loginServer.
//   - authentication_credentials (Required, MaxItems 1) maps to properties.authCredentials.
//   - identity is a Required system-assigned identity (envelope), not a body property.
type ContainerRegistryCredentialSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryCredentialSet)(nil)

// NewContainerRegistryCredentialSet returns knowledge for the credentialSets resource.
func NewContainerRegistryCredentialSet() *ContainerRegistryCredentialSet {
	return &ContainerRegistryCredentialSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/credentialSets",
			ApiVersions:  []string{"2023-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// login_server is a ForceNew body property.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.loginServer"},
			},
			ArrayRules: []azwise.ArrayRule{
				// authentication_credentials MaxItems 1.
				{PropertyPath: "properties.authCredentials", MaxItems: 1},
			},
			// login_server and authentication_credentials are Required for creation.
			RequiredFields: []string{
				"properties.loginServer",
				"properties.authCredentials",
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
func init() { azwise.Register(NewContainerRegistryCredentialSet()) }
