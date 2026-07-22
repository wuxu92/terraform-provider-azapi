package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementDiagnostic provides resource knowledge for
// Microsoft.ApiManagement/service/diagnostics.
//
// Mirrors azurerm_api_management_diagnostic.
//
// Sources (terraform-provider-azurerm internal/services/apimanagement):
//   - api_management_diagnostic_resource.go (schema L45-134; expand L160-220;
//     additional-content block L140-176 in api_management_api_diagnostic_resource.go)
//   - SDK model DiagnosticContractProperties (loggerId required; alwaysLog/verbosity/
//     httpCorrelationProtocol/operationNameFormat/logClientIp/sampling optional)
//   - SDK models PipelineDiagnosticSettings/HTTPMessageDiagnostic/BodyDiagnosticSettings
//   - diagnostic/constants.go enums (Verbosity, HTTPCorrelationProtocol,
//     OperationNameFormat, AlwaysLog, DataMaskingMode)
//
// The resource name (`identifier`) is ForceNew and constrained to
// applicationinsights|azuremonitor.
//
// TODO: AzureRM CustomizeDiff forbids operation_name_format unless identifier ==
// "applicationinsights"; this value-conditional constraint cannot be expressed
// with the current declarative rule types.
// TODO: data_masking entities are arrays of {mode Hide|Mask, value} nested under
// frontend/backend request/response (e.g. properties.frontend.request.dataMasking.
// queryParams[*].mode). Array-element mode enums are skipped here.
type ApiManagementDiagnostic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementDiagnostic)(nil)

// NewApiManagementDiagnostic returns knowledge for the diagnostics resource.
func NewApiManagementDiagnostic() *ApiManagementDiagnostic {
	maxBodyBytes := int64(8192)
	minBodyBytes := int64(0)
	return &ApiManagementDiagnostic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/diagnostics",
			ApiVersions:  []string{"2022-08-01"},
			RequiredFields: []string{
				"properties.loggerId",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name is the diagnostic identifier.
				{AllowedValues: []string{"applicationinsights", "azuremonitor"}, Message: "identifier must be applicationinsights or azuremonitor"},
				{PropertyPath: "properties.verbosity", AllowedValues: []string{"verbose", "information", "error"}, Message: "verbosity must be verbose, information or error"},
				{PropertyPath: "properties.httpCorrelationProtocol", AllowedValues: []string{"None", "Legacy", "W3C"}, Message: "http_correlation_protocol must be None, Legacy or W3C"},
				{PropertyPath: "properties.operationNameFormat", AllowedValues: []string{"Name", "Url"}, Message: "operation_name_format must be Name or Url"},
				{PropertyPath: "properties.alwaysLog", AllowedValues: []string{"allErrors"}, Message: "always_log must be allErrors"},
			},
			FloatRules: []azwise.FloatRule{
				{PropertyPath: "properties.sampling.percentage", MinValue: azwise.Ptr(0.0), MaxValue: azwise.Ptr(100.0), Message: "sampling_percentage must be between 0 and 100"},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.frontend.request.body.bytes", MinValue: &minBodyBytes, MaxValue: &maxBodyBytes, Message: "body_bytes must be between 0 and 8192"},
				{PropertyPath: "properties.frontend.response.body.bytes", MinValue: &minBodyBytes, MaxValue: &maxBodyBytes, Message: "body_bytes must be between 0 and 8192"},
				{PropertyPath: "properties.backend.request.body.bytes", MinValue: &minBodyBytes, MaxValue: &maxBodyBytes, Message: "body_bytes must be between 0 and 8192"},
				{PropertyPath: "properties.backend.response.body.bytes", MinValue: &minBodyBytes, MaxValue: &maxBodyBytes, Message: "body_bytes must be between 0 and 8192"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementDiagnostic()) }
