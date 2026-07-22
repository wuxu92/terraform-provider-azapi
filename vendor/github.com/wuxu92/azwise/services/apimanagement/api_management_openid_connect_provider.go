package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementOpenIDConnectProvider provides resource knowledge for
// Microsoft.ApiManagement/service/openidConnectProviders.
//
// Mirrors azurerm_api_management_openid_connect_provider. client_id and
// client_secret are sensitive. The name segment is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_openid_connect_provider_resource.go
//     schema (lines 41-78), CreateUpdate body (lines 105-113), timeouts (30m/5m/30m/30m)
//   - go-azure-sdk apimanagement/2022-08-01/openidconnectprovider
//     OpenidConnectProviderContractProperties
type ApiManagementOpenIDConnectProvider struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementOpenIDConnectProvider)(nil)

func NewApiManagementOpenIDConnectProvider() *ApiManagementOpenIDConnectProvider {
	return &ApiManagementOpenIDConnectProvider{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/openidConnectProviders",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // name (SchemaApiManagementChildName, ForceNew)
			},
			RequiredFields: []string{
				"properties.clientId",
				"properties.displayName",
				"properties.metadataEndpoint",
			},
			SensitiveFields: []string{
				"properties.clientId",     // Sensitive: true in AzureRM schema
				"properties.clientSecret", // Sensitive: true in AzureRM schema
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── ApiManagementChildName
				{
					Regex:   `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					Message: "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length",
				},
				// ── client_id → properties.clientId ── StringIsNotEmpty
				{PropertyPath: "properties.clientId", MinLength: 1, Message: "client_id must not be empty"},
				// ── client_secret → properties.clientSecret ── StringIsNotEmpty
				{PropertyPath: "properties.clientSecret", MinLength: 1, Message: "client_secret must not be empty"},
				// ── display_name → properties.displayName ── StringIsNotEmpty
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				// ── metadata_endpoint → properties.metadataEndpoint ── StringIsNotEmpty
				{PropertyPath: "properties.metadataEndpoint", MinLength: 1, Message: "metadata_endpoint must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementOpenIDConnectProvider()) }
