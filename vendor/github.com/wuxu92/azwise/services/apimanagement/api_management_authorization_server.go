package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementAuthorizationServer provides resource knowledge for
// Microsoft.ApiManagement/service/authorizationServers.
//
// Contributing Terraform resource: azurerm_api_management_authorization_server.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_authorization_server_resource.go:23-190
//     (schema: name/rg/apim ForceNew; required endpoints/methods/client_id/grant_types;
//     optional client_authentication_method/client_secret/default_scope/description/
//     resource_owner_*/support_state/token_body_parameter/token_endpoint)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_authorization_server_resource.go:233-268
//     (create: AuthorizationServerContractProperties field mapping)
//   - terraform-provider-azurerm internal/services/apimanagement/validate/api_management.go:12-21
//     (ApiManagementChildName regex via schemaz.SchemaApiManagementChildName)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/authorizationserver/constants.go:14-23,73-76,114-117,155-160
//     (AuthorizationMethod, BearerTokenSendingMethod, ClientAuthenticationMethod, GrantType)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/authorizationserver/id_authorizationserver.go:116-127
//     (resource ID segments: .../authorizationServers/{authorizationServerName})
//
// Intentionally skipped here:
//   - resource_group_name / api_management_name: AzAPI ID segments, not body props.
type ApiManagementAuthorizationServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementAuthorizationServer)(nil)

// NewApiManagementAuthorizationServer returns knowledge for the
// Microsoft.ApiManagement/service/authorizationServers resource.
func NewApiManagementAuthorizationServer() *ApiManagementAuthorizationServer {
	return &ApiManagementAuthorizationServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/authorizationServers",
			ApiVersions:  []string{"2022-08-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name — validate.ApiManagementChildName.
					Regex:     `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					MaxLength: 80,
					Message:   "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length, beginning and ending with an alphanumeric character",
				},
				{
					PropertyPath: "properties.authorizationEndpoint",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath: "properties.clientId",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath: "properties.clientRegistrationEndpoint",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath:  "properties.authorizationMethods[*]",
					AllowedValues: []string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE"},
					Message:       "must be one of DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT or TRACE",
				},
				{
					PropertyPath:  "properties.grantTypes[*]",
					AllowedValues: []string{"authorizationCode", "clientCredentials", "implicit", "resourceOwnerPassword"},
					Message:       "must be one of authorizationCode, clientCredentials, implicit or resourceOwnerPassword",
				},
				{
					PropertyPath:  "properties.bearerTokenSendingMethods[*]",
					AllowedValues: []string{"authorizationHeader", "query"},
					Message:       "must be one of authorizationHeader or query",
				},
				{
					PropertyPath:  "properties.clientAuthenticationMethod[*]",
					AllowedValues: []string{"Basic", "Body"},
					Message:       "must be one of Basic or Body",
				},
				{
					PropertyPath: "properties.tokenBodyParameters[*].name",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath: "properties.tokenBodyParameters[*].value",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			SensitiveFields: []string{
				"properties.clientSecret",
				"properties.resourceOwnerPassword",
			},
			RequiredFields: []string{
				"properties.authorizationEndpoint",
				"properties.authorizationMethods",
				"properties.clientId",
				"properties.clientRegistrationEndpoint",
				"properties.displayName",
				"properties.grantTypes",
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementAuthorizationServer()) }
