package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiDiagnostic provides resource knowledge for
// Microsoft.ApiManagement/service/apis/diagnostics.
//
// Contributing Terraform resource: azurerm_api_management_api_diagnostic.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_diagnostic_resource.go:26-176
//     (schema: identifier ForceNew enum; api_management_logger_id Required;
//     sampling_percentage FloatBetween(0,100); verbosity/http_correlation_protocol/
//     operation_name_format enums; frontend/backend request+response blocks with
//     body_bytes IntBetween(0,8192), headers_to_log, data_masking Hide/Mask)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_diagnostic_resource.go:201-262
//     (create: DiagnosticContractProperties.LoggerId/OperationNameFormat/Sampling/
//     AlwaysLog/Verbosity/LogClientIP/HTTPCorrelationProtocol/Frontend/Backend)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_diagnostic_resource.go:365-468
//     (expand HTTPMessageDiagnostic body/headers/dataMasking, DataMaskingEntity mode/value)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apidiagnostic/constants.go:14-16,52-55,93-97,137-140,178-180,216-220
//     (AlwaysLog, DataMaskingMode, HTTPCorrelationProtocol, OperationNameFormat, SamplingType, Verbosity)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apidiagnostic/model_*.go
//     (json tags: alwaysLog/backend/frontend/httpCorrelationProtocol/logClientIp/loggerId/
//     operationNameFormat/sampling{percentage,samplingType}/verbosity;
//     PipelineDiagnosticSettings{request,response}; HTTPMessageDiagnostic{body{bytes},headers,
//     dataMasking{headers,queryParams}}; DataMaskingEntity{mode,value})
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apidiagnostic/id_apidiagnostic.go:122-135
//     (resource ID segments: .../apis/{apiId}/diagnostics/{diagnosticId})
//
// Intentionally skipped here:
//   - resource_group_name / api_management_name / api_name: AzAPI ID segments.
//   - CustomizeDiff (operation_name_format may only be set when identifier is
//     "applicationinsights"): a value-dependent cross-field rule keyed off the resource
//     name (identifier), which azwise's declarative rules cannot express; ARM/AzureRM
//     enforce it. Noted here for completeness.
//   - always_log_errors (bool) maps to properties.alwaysLog = "allErrors" (a one-way
//     bool->enum projection), not a settable enum field, so no StringRule is emitted.
//   - properties.sampling.samplingType is hardcoded to "fixed" by AzureRM whenever
//     sampling is configured; emitted below as a DefaultValue.
type ApiManagementApiDiagnostic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiDiagnostic)(nil)

// NewApiManagementApiDiagnostic returns knowledge for the
// Microsoft.ApiManagement/service/apis/diagnostics resource.
func NewApiManagementApiDiagnostic() *ApiManagementApiDiagnostic {
	maskingModes := []string{"Hide", "Mask"}
	byteBounds := func(path string) azwise.IntRule {
		return azwise.IntRule{
			PropertyPath: path,
			MinValue:     azwise.Ptr(int64(0)),
			MaxValue:     azwise.Ptr(int64(8192)),
			Message:      "must be between 0 and 8192",
		}
	}
	maskingRules := func(base string) []azwise.StringRule {
		return []azwise.StringRule{
			{PropertyPath: base + ".dataMasking.queryParams[*].mode", AllowedValues: maskingModes, Message: "must be one of Hide or Mask"},
			{PropertyPath: base + ".dataMasking.queryParams[*].value", MinLength: 1, Message: "must not be empty"},
			{PropertyPath: base + ".dataMasking.headers[*].mode", AllowedValues: maskingModes, Message: "must be one of Hide or Mask"},
			{PropertyPath: base + ".dataMasking.headers[*].value", MinLength: 1, Message: "must not be empty"},
		}
	}

	stringRules := []azwise.StringRule{
		{
			// identifier (resource name) — applicationinsights or azuremonitor.
			AllowedValues: []string{"applicationinsights", "azuremonitor"},
			Message:       "must be one of applicationinsights or azuremonitor",
		},
		{
			PropertyPath:  "properties.verbosity",
			AllowedValues: []string{"verbose", "information", "error"},
			Message:       "must be one of verbose, information or error",
		},
		{
			PropertyPath:  "properties.httpCorrelationProtocol",
			AllowedValues: []string{"None", "Legacy", "W3C"},
			Message:       "must be one of None, Legacy or W3C",
		},
		{
			PropertyPath:  "properties.operationNameFormat",
			AllowedValues: []string{"Name", "Url"},
			Message:       "must be one of Name or Url",
		},
	}
	for _, base := range []string{
		"properties.frontend.request",
		"properties.frontend.response",
		"properties.backend.request",
		"properties.backend.response",
	} {
		stringRules = append(stringRules, maskingRules(base)...)
	}

	return &ApiManagementApiDiagnostic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/diagnostics",
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
			StringRules: stringRules,
			FloatRules: []azwise.FloatRule{
				{
					PropertyPath: "properties.sampling.percentage",
					MinValue:     azwise.Ptr(0.0),
					MaxValue:     azwise.Ptr(100.0),
					Message:      "must be between 0.0 and 100.0",
				},
			},
			IntRules: []azwise.IntRule{
				byteBounds("properties.frontend.request.body.bytes"),
				byteBounds("properties.frontend.response.body.bytes"),
				byteBounds("properties.backend.request.body.bytes"),
				byteBounds("properties.backend.response.body.bytes"),
			},
			RequiredFields: []string{
				"properties.loggerId",
			},
			DefaultValues: []azwise.DefaultValue{
				// operation_name_format Default "Name".
				{PropertyPath: "properties.operationNameFormat", Value: "Name"},
				// AzureRM hardcodes samplingType to "fixed" whenever sampling is set.
				{PropertyPath: "properties.sampling.samplingType", Value: "fixed"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiDiagnostic()) }
