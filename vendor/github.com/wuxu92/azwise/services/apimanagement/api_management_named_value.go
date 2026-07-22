package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementNamedValue provides resource knowledge for
// Microsoft.ApiManagement/service/namedValues.
//
// Mirrors azurerm_api_management_named_value.
//
// Sources (terraform-provider-azurerm internal/services/apimanagement):
//   - api_management_named_value_resource.go (schema L47-103; expand L136-154, L232-247)
//   - SDK model NamedValueCreateContractProperties (displayName required; secret/value/tags/keyVault optional)
//   - SDK model KeyVaultContractCreateProperties (secretIdentifier/identityClientId)
//   - validate.ApiManagementNamedValueDisplayName (regex ^[0-9a-zA-Z_.-]{1,256}$)
//
// TODO: AzureRM enforces `secret` == true whenever `value_from_key_vault` is set
// (a value-based cross-field check in the Create func, not a mere presence check),
// which RelationalRule cannot express; it is therefore not encoded here.
type ApiManagementNamedValue struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementNamedValue)(nil)

// NewApiManagementNamedValue returns knowledge for the namedValues resource.
func NewApiManagementNamedValue() *ApiManagementNamedValue {
	return &ApiManagementNamedValue{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/namedValues",
			ApiVersions:  []string{"2022-08-01"},
			RequiredFields: []string{
				"properties.displayName",
			},
			SensitiveFields: []string{
				"properties.value",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.secret", Value: false},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.displayName", Regex: `^[0-9a-zA-Z_.-]{1,256}$`, Message: "display_name may only contain alphanumeric characters, periods, underscores and dashes (max 256)"},
				{PropertyPath: "properties.value", MinLength: 1, Message: "value must not be empty"},
				{PropertyPath: "properties.keyVault.identityClientId", Regex: `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, Message: "identity_client_id must be a valid UUID"},
			},
			// value XOR value_from_key_vault (AzureRM ExactlyOneOf on value/value_from_key_vault).
			ExactlyOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.value", "properties.keyVault"}, Message: "exactly one of value or value_from_key_vault must be set"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementNamedValue()) }
