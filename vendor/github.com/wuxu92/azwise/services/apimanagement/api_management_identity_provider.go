package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementIdentityProvider provides resource knowledge for
// Microsoft.ApiManagement/service/identityProviders.
//
// This single ARM resource type backs six AzureRM Terraform resources, all of
// which PUT to the same identityProviders API and differ only by the `type`
// discriminator (the ARM resource name segment):
//
//   - azurerm_api_management_identity_provider_aad       (type "aad")
//   - azurerm_api_management_identity_provider_aadb2c    (type "aadB2C")
//   - azurerm_api_management_identity_provider_facebook  (type "facebook")
//   - azurerm_api_management_identity_provider_google    (type "google")
//   - azurerm_api_management_identity_provider_microsoft (type "microsoft")
//   - azurerm_api_management_identity_provider_twitter   (type "twitter")
//
// The knowledge below is the UNION of all six resources' body constraints.
// clientId/clientSecret are shared by every provider; the AAD/AADB2C-only fields
// (allowedTenants, signinTenant, authority, *PolicyName, clientLibrary) apply
// only to those two types, so they are declared but not marked globally required.
//
// Sources (terraform-provider-azurerm internal/services/apimanagement):
//   - api_management_identity_provider_aad_resource.go       (schema L40-78, expand L110-119)
//   - api_management_identity_provider_aadb2c_resource.go    (schema L40-108, expand L148-162)
//   - api_management_identity_provider_facebook_resource.go  (schema L39-56, expand L85-91)
//   - api_management_identity_provider_google_resource.go    (schema L40-57, expand L86-92; validate.GoogleClientID)
//   - api_management_identity_provider_microsoft_resource.go (schema L39-56, expand L85-91)
//   - api_management_identity_provider_twitter_resource.go   (schema L39-57, expand L86-92)
//   - SDK model IdentityProviderCreateContractProperties (clientId/clientSecret required;
//     allowedTenants/authority/*PolicyName/clientLibrary/signinTenant optional)
//   - identityprovider/constants.go IdentityProviderType enum + id_identityprovider.go
//
// TODO: The per-provider clientId validators cannot be unified in a merged file:
// aad/aadb2c/microsoft use validation.IsUUID, google uses a
// `*.apps.googleusercontent.com` regex (validate.GoogleClientID), facebook(app_id)
// and twitter(api_key) only require non-empty. The rule below applies the common
// denominator (non-empty). Twitter also marks its clientId (api_key) sensitive,
// but the ARM clientId path is not universally sensitive across providers so it is
// not listed in SensitiveFields.
type ApiManagementIdentityProvider struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementIdentityProvider)(nil)

// NewApiManagementIdentityProvider returns knowledge for the identityProviders resource.
func NewApiManagementIdentityProvider() *ApiManagementIdentityProvider {
	return &ApiManagementIdentityProvider{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/identityProviders",
			ApiVersions:  []string{"2022-08-01"},
			// clientId + clientSecret are required for every provider type.
			// AAD/AADB2C-specific required fields (allowedTenants, signinTenant,
			// authority, signupPolicyName, signinPolicyName) are type-conditional
			// and therefore omitted from the global required set.
			RequiredFields: []string{
				"properties.clientId",
				"properties.clientSecret",
			},
			SensitiveFields: []string{
				"properties.clientSecret",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name is the identity provider type discriminator.
				{AllowedValues: []string{"aad", "aadB2C", "facebook", "google", "microsoft", "twitter"}, Message: "identity provider name must be a valid IdentityProviderType"},
				{PropertyPath: "properties.clientId", MinLength: 1, Message: "client_id must not be empty"},
				{PropertyPath: "properties.clientSecret", MinLength: 1, Message: "client_secret must not be empty"},
				{PropertyPath: "properties.clientLibrary", MaxLength: 16, Message: "client_library must be at most 16 characters"},
				{PropertyPath: "properties.signinTenant", MinLength: 1, Message: "signin_tenant must not be empty"},
				{PropertyPath: "properties.authority", MinLength: 1, Message: "authority must not be empty"},
				{PropertyPath: "properties.signupPolicyName", MinLength: 1, Message: "signup_policy must not be empty"},
				{PropertyPath: "properties.signinPolicyName", MinLength: 1, Message: "signin_policy must not be empty"},
				{PropertyPath: "properties.profileEditingPolicyName", MinLength: 1, Message: "profile_editing_policy must not be empty"},
				{PropertyPath: "properties.passwordResetPolicyName", MinLength: 1, Message: "password_reset_policy must not be empty"},
				// allowedTenants elements: AAD requires UUIDs, AADB2C accepts an
				// arbitrary tenant string; common denominator is non-empty.
				{PropertyPath: "properties.allowedTenants[*]", MinLength: 1, Message: "allowed_tenant must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementIdentityProvider()) }
