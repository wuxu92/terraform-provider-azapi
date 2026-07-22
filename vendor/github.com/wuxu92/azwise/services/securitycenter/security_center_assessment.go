// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterAssessment provides resource knowledge for
// Microsoft.Security/assessments.
//
// Mirrors azurerm_security_center_assessment. The assessment is created under a
// target resource (target_resource_id, the parent scope) and named after its
// assessment policy metadata (both are envelope/scope inputs, not body ForceNew).
// The body carries properties.status (required) and properties.additionalData.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_assessment_resource.go
//     (schema lines 43-96; body expand lines 126-134: properties.status{code,cause,description},
//     properties.additionalData, properties.resourceDetails.source hardcoded "Azure"; Create/Update 30m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/parse/assessment.go
//     (ID .../providers/Microsoft.Security/assessments/<name>)
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security AssessmentProperties / AssessmentStatus
//     (status.code enum Healthy/NotApplicable/Unhealthy). Preview SDK: no ARM API version pinned.
type SecurityCenterAssessment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterAssessment)(nil)

// NewSecurityCenterAssessment returns knowledge for the assessments resource.
func NewSecurityCenterAssessment() *SecurityCenterAssessment {
	return &SecurityCenterAssessment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/assessments",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.status",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.status.code",
					AllowedValues: []string{"Healthy", "NotApplicable", "Unhealthy"},
					Message:       "status code must be one of Healthy, NotApplicable, or Unhealthy",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterAssessment()) }
