package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApi provides resource knowledge for
// Microsoft.ApiManagement/service/apis.
//
// Mirrors azurerm_api_management_api. `name`, `api_management_name` and
// `resource_group_name` are envelope references. `revision` is ForceNew and is
// folded into the ARM resource name (apiId = "name;rev=revision"), so it lives on
// the envelope, not the body.
//
// Top-level ARM body fields (api.ApiCreateOrUpdateProperties):
//   - display_name          -> properties.displayName (Optional+Computed)
//   - path                  -> properties.path (Optional+Computed; ApiManagementApiPath validator)
//   - protocols             -> properties.protocols[*] (enum http|https|ws|wss)
//   - api_type              -> properties.type (enum graphql|http|soap|websocket, default http)
//   - revision_description  -> properties.apiRevisionDescription
//   - version               -> properties.apiVersion (Optional+Computed)
//   - version_description   -> properties.apiVersionDescription
//   - version_set_id        -> properties.apiVersionSetId (Optional+Computed)
//   - description           -> properties.description
//   - service_url           -> properties.serviceUrl (Optional+Computed)
//   - source_api_id         -> properties.sourceApiId
//   - subscription_required -> properties.subscriptionRequired (default true)
//   - terms_of_service_url  -> properties.termsOfServiceUrl
//
// Nested blocks:
//   - contact                          -> properties.contact {email, name, url}
//   - license                          -> properties.license {name, url}
//   - subscription_key_parameter_names -> properties.subscriptionKeyParameterNames {header, query}
//   - oauth2_authorization             -> properties.authenticationSettings.oAuth2
//   - openid_authentication            -> properties.authenticationSettings.openid
//                                         {openidProviderId, bearerTokenSendingMethods[*]}
//   - import                           -> properties.value/format (import-only, two-phase PUT)
//
// TODO: several AzureRM validators are semantic/conditional and are not emitted
// as declarative rules:
//   - path (ApiManagementApiPath), source_api_id (ApiID), contact.email
//     (IsEmailAddress), contact.url/license.url/terms_of_service_url
//     (IsURLWithHTTPorHTTPS), oauth2.authorization_server_name /
//     openid.openid_provider_name (ApiManagementChildName) belong in azapin
//     semantic validators on their ARM paths.
//   - CustomizeDiff conditionals: `display_name` + `protocols` are required only
//     when `source_api_id` is not set; `service_url` is required only when
//     `api_type` is websocket. These conditional requirements cannot be expressed
//     with the current rule types, so no unconditional RequiredFields are emitted.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_resource.go
//     schema (lines 53-341); CustomizeDiff (lines 343-360); Create body
//     (lines 442-484); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/api
//     ApiCreateOrUpdateProperties + AuthenticationSettingsContract (oAuth2, openid);
//     enum constants ApiType, Protocol, BearerTokenSendingMethods (constants.go).
type ApiManagementApi struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApi)(nil)

// NewApiManagementApi returns knowledge for the apis resource.
func NewApiManagementApi() *ApiManagementApi {
	return &ApiManagementApi{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.type",
					AllowedValues: []string{"graphql", "http", "soap", "websocket"},
					Message:       "api_type must be one of graphql, http, soap or websocket",
				},
				{
					PropertyPath:  "properties.protocols[*]",
					AllowedValues: []string{"http", "https", "ws", "wss"},
					Message:       "protocols must be one of http, https, ws or wss",
				},
				{
					PropertyPath:  "properties.authenticationSettings.openid.bearerTokenSendingMethods[*]",
					AllowedValues: []string{"authorizationHeader", "query"},
					Message:       "bearer_token_sending_methods must be one of authorizationHeader or query",
				},
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				{PropertyPath: "properties.apiRevisionDescription", MinLength: 1, Message: "revision_description must not be empty"},
				{PropertyPath: "properties.apiVersionDescription", MinLength: 1, Message: "version_description must not be empty"},
			},
			// AzureRM: subscription_required Default true; api_type defaults to http
			// in the create body when unset.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.subscriptionRequired", Value: true},
				{PropertyPath: "properties.type", Value: "http"},
			},
			// AzureRM: oauth2_authorization ConflictsWith openid_authentication.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.authenticationSettings.oAuth2",
						"properties.authenticationSettings.openid",
					},
					Message: "`oauth2_authorization` conflicts with `openid_authentication`",
				},
			},
			// AzureRM CustomizeDiff: `version` requires `version_set_id`.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.apiVersion", "properties.apiVersionSetId"},
					Message: "`version` requires `version_set_id`",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApi()) }
