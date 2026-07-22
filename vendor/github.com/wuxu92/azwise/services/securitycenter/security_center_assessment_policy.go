// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterAssessmentPolicy provides resource knowledge for
// Microsoft.Security/assessmentMetadata.
//
// Mirrors azurerm_security_center_assessment_policy. The subscription-scoped
// metadata resource is named with a server-generated GUID (computed name).
// properties.assessmentType is hardcoded by AzureRM to "CustomerManaged".
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_assessment_policy_resource.go
//     (schema lines 43-135; create params lines 162-197: description/displayName Required,
//     severity default "Medium", categories/threats arrays, implementationEffort/userImpact enums;
//     Create/Update 30m, Read 5m)
//   - go-azure-sdk resource-manager/security/2021-06-01/assessmentsmetadata:
//     id_providerassessmentmetadata.go (segment "assessmentMetadata"),
//     model_securityassessmentmetadatapropertiesresponse.go (json tags),
//     constants.go (Severity High/Low/Medium, ImplementationEffort High/Low/Moderate, UserImpact High/Low/Moderate)
type SecurityCenterAssessmentPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterAssessmentPolicy)(nil)

// NewSecurityCenterAssessmentPolicy returns knowledge for the assessmentMetadata resource.
func NewSecurityCenterAssessmentPolicy() *SecurityCenterAssessmentPolicy {
	return &SecurityCenterAssessmentPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/assessmentMetadata",
			ApiVersions:  []string{"2021-06-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.description",
				"properties.displayName",
				// AzureRM hardcodes assessmentType = "CustomerManaged" for user-created policies.
				"properties.assessmentType",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.severity", Value: "Medium"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.severity",
					AllowedValues: []string{"High", "Low", "Medium"},
					Message:       "severity must be one of High, Low, or Medium",
				},
				{
					PropertyPath:  "properties.implementationEffort",
					AllowedValues: []string{"High", "Low", "Moderate"},
					Message:       "implementationEffort must be one of High, Low, or Moderate",
				},
				{
					PropertyPath:  "properties.userImpact",
					AllowedValues: []string{"High", "Low", "Moderate"},
					Message:       "userImpact must be one of High, Low, or Moderate",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterAssessmentPolicy()) }
