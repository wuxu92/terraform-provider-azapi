package nginx

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NginxApiKey provides resource knowledge for Nginx.NginxPlus/nginxDeployments/apiKeys.
//
// Mirrors azurerm_nginx_api_key. This is a management-plane ARM resource (ARM type
// Nginx.NginxPlus/nginxDeployments/apiKeys, id validator nginxapikey.ValidateApiKeyID),
// not a data-plane-only resource.
//
// ARM type casing verified against go-azure-sdk nginxapikey/id_apikey.go Segments:
// "Nginx.NginxPlus" / "nginxDeployments" / "apiKeys".
//
// Sources:
//   - terraform-provider-azurerm internal/services/nginx/nginx_api_key_resource.go
//     Arguments (34-64: name StringIsNotEmpty ForceNew; nginx_deployment_id parent ref ForceNew;
//     end_date_time validate.EndDateTime RFC3339 within 2 years; secret_text StringIsNotEmpty
//     Sensitive), Attributes (66-73: hint Computed), Create (83-132), timeouts 5m across the board.
//   - go-azure-sdk resource-manager/nginx/2024-11-01-preview/nginxapikey
//     NginxDeploymentApiKeyRequestProperties (endDateTime/secretText settable),
//     NginxDeploymentApiKeyResponseProperties (endDateTime/hint; hint read-only).
//
// Note: end_date_time uses the semantic validator validate.EndDateTime (RFC3339 and no further out
// than 2 years); the 2-year bound cannot be expressed declaratively and is not emitted as a rule.
type NginxApiKey struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NginxApiKey)(nil)

// NewNginxApiKey returns knowledge for the Nginx apiKeys resource.
func NewNginxApiKey() *NginxApiKey {
	return &NginxApiKey{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Nginx.NginxPlus/nginxDeployments/apiKeys",
			ApiVersions:  []string{"2024-11-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Update: 5 * time.Minute,
				Delete: 5 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			RequiredFields: []string{
				"properties.endDateTime",
				"properties.secretText",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "name", MinLength: 1, Message: "name must not be empty"},
				{PropertyPath: "properties.secretText", MinLength: 1, Message: "secret_text must not be empty"},
			},
			SensitiveFields: []string{
				"properties.secretText",
			},
			// hint is only present in the response model — server-populated, read-only.
			ComputedFields: []string{
				"properties.hint",
			},
		},
	}
}

func init() { azwise.Register(NewNginxApiKey()) }
