package managedidentity

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FederatedIdentityCredential provides resource knowledge for
// Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials.
//
// Mirrors azurerm_federated_identity_credential. The credential name and its
// parent user_assigned_identity_id are envelope/parent references (Required +
// RequiresReplace by construction), so their ForceNew is not repeated as body
// rules. The body carries only properties.{audiences,issuer,subject}, all
// Required and freely updatable (no body-level ForceNew).
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedidentity/federated_identity_credential_resource.go
//   - Arguments(): name (ForceNew, envelope), user_assigned_identity_id (ForceNew, parent),
//     audience (Required, TypeList MaxItems:1 -> properties.audiences),
//     issuer (Required -> properties.issuer), subject (Required -> properties.subject)
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/managedidentity/2024-11-30/federatedidentitycredentials
//     model_federatedidentitycredentialproperties.go:
//     FederatedIdentityCredentialProperties.{Audiences []string, Issuer string, Subject string}
//     — no read-only fields, so nothing is stripped from the PUT body.
type FederatedIdentityCredential struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FederatedIdentityCredential)(nil)

// NewFederatedIdentityCredential returns knowledge for the
// federatedIdentityCredentials resource.
func NewFederatedIdentityCredential() *FederatedIdentityCredential {
	return &FederatedIdentityCredential{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials",
			ApiVersions:  []string{"2024-11-30"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// All three body properties are Required for a valid credential.
			RequiredFields: []string{
				"properties.audiences",
				"properties.issuer",
				"properties.subject",
			},
			// audience is a TypeList with MaxItems:1 mapping straight onto the
			// ARM audiences array.
			ArrayRules: []azwise.ArrayRule{
				{PropertyPath: "properties.audiences", MaxItems: 1},
			},
		},
	}
}

func init() { azwise.Register(NewFederatedIdentityCredential()) }
