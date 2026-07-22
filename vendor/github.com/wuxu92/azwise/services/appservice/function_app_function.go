package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FunctionAppFunction provides resource knowledge for Microsoft.Web/sites/functions.
//
// Mirrors azurerm_function_app_function. The ARM resource name is the function
// name (azurerm name); the parent envelope is the function app (azurerm
// function_app_id). Note the enabled -> isDisabled inversion: AzureRM's
// enabled=true maps to ARM isDisabled=false.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/function_app_function_resource.go:25-274,341-395
//     (schema: name/function_app_id/file ForceNew; config_json Required; enabled Default true ->
//     isDisabled false; language enum; create maps config/testData/language/isDisabled/files;
//     timeouts Create 30m / Read 5m / Update 30m / Delete 5m)
//   - terraform-provider-azurerm internal/services/appservice/validate/function_app_function_name.go:18
//     (name regex)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_functionenvelopeproperties.go:6-20
//     (settable: config, files, isDisabled, language, test_data; read-only hrefs/URLs)
//
// config_json maps to properties.config, which is an arbitrary JSON object in
// ARM (not a string) so its StringIsJSON validation is not expressible declaratively.
type FunctionAppFunction struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FunctionAppFunction)(nil)

// NewFunctionAppFunction returns knowledge for the sites/functions resource.
func NewFunctionAppFunction() *FunctionAppFunction {
	return &FunctionAppFunction{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/functions",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.files"},
			},
			RequiredFields: []string{
				"properties.config",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 5 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `^[0-9a-zA-Z](([-_0-9a-zA-Z-]{0,126})[-_0-9a-zA-Z])?$`,
					Message: "function name must start with a letter, contain only alphanumeric characters, dashes and underscores, up to 128 characters",
				},
				{
					PropertyPath:  "properties.language",
					AllowedValues: []string{"CSharp", "Custom", "Java", "Javascript", "Python", "PowerShell", "TypeScript"},
					Message:       "language must be one of CSharp, Custom, Java, Javascript, Python, PowerShell, TypeScript",
				},
				{PropertyPath: "properties.test_data", MinLength: 1, Message: "test_data must not be empty"},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM enabled defaults true -> ARM isDisabled false.
				{PropertyPath: "properties.isDisabled", Value: false},
			},
			ComputedFields: []string{
				"properties.config_href",
				"properties.function_app_id",
				"properties.href",
				"properties.invoke_url_template",
				"properties.script_href",
				"properties.script_root_path_href",
				"properties.secrets_file_href",
				"properties.test_data_href",
			},
		},
	}
}

func init() { azwise.Register(NewFunctionAppFunction()) }
