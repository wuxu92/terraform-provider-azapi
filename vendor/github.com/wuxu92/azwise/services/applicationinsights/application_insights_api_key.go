package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsightsAPIKey provides resource knowledge for
// Microsoft.Insights/components/apiKeys.
//
// Mirrors azurerm_application_insights_api_key. The ARM resource name is a
// server-generated keyId, so the user-supplied name is a genuine flat body
// property (APIKeyRequest.name), not an envelope segment. The request body is flat
// (name / linkedReadProperties / linkedWriteProperties). Everything is ForceNew:
// api keys are immutable (no Update).
//
// Deliberately not encoded here:
//   - read_permissions / write_permissions are validated against enum allow-sets
//     (agentconfig/aggregate/api/draft/extendqueries/search and annotations), but
//     AzureRM expands each element to "<applicationInsightsId>/<permission>"
//     (helpers.go:13-23) before sending, so the ARM linkedReadProperties[*] /
//     linkedWriteProperties[*] values are full resource-scoped strings that a raw
//     enum StringRule would not match. The permission enums are therefore not
//     expressible as declarative element rules.
//   - application_insights_id (components.ValidateComponentID) is the parent
//     reference on the envelope, not a body property.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_api_key_resource.go:22-158
//     (schema: name/read_permissions/write_permissions ForceNew, api_key
//     Computed+Sensitive, "at least one read or write permission" runtime check,
//     timeouts 30m/5m/-/30m; Create maps name/linkedReadProperties/linkedWriteProperties)
//   - terraform-provider-azurerm internal/services/applicationinsights/helpers.go:13-23
//     (linked property expand prefixes the component ID)
//   - go-azure-sdk resource-manager/applicationinsights/2015-05-01/componentapikeysapis
//     model_apikeyrequest.go (create body), model_applicationinsightscomponentapikey.go
//     (response: apiKey read-only), id_apikey.go (ARM type + 2015-05-01)
type ApplicationInsightsAPIKey struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsightsAPIKey)(nil)

// NewApplicationInsightsAPIKey returns knowledge for the apiKeys resource.
func NewApplicationInsightsAPIKey() *ApplicationInsightsAPIKey {
	return &ApplicationInsightsAPIKey{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/components/apiKeys",
			ApiVersions:  []string{"2015-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "linkedReadProperties"},
				{PropertyPath: "linkedWriteProperties"},
			},
			StringRules: []azwise.StringRule{
				// name validation.NoZeroValues.
				{PropertyPath: "name", MinLength: 1, Message: "name must not be empty"},
			},
			RequiredFields: []string{
				"name",
			},
			// AzureRM enforces "at least one read or write permission must be defined"
			// (api_key_resource.go:129-131).
			AtLeastOneOf: []azwise.RelationalRule{
				{Paths: []string{"linkedReadProperties", "linkedWriteProperties"}},
			},
			// apiKey is present only in the GET/create response, absent from the
			// APIKeyRequest create body — a genuine read-only field safe to strip.
			ComputedFields: []string{
				"apiKey",
			},
			SensitiveFields: []string{
				"apiKey",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsightsAPIKey()) }
