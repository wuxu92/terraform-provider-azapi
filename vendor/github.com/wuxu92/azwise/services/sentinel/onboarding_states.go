// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// OnboardingStates provides resource knowledge for
// Microsoft.SecurityInsights/onboardingStates (an extension resource on a
// Microsoft.OperationalInsights/workspaces scope). AzureRM only ever creates the
// singleton "default" onboarding state.
//
// Contributing Terraform resource: azurerm_sentinel_log_analytics_workspace_onboarding.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_log_analytics_workspace_onboarding_resource.go
//     (schema L40-57, body L92-96, timeouts L65)
//   - go-azure-sdk resource-manager/securityinsights/2022-11-01/sentinelonboardingstates:
//     id_onboardingstate.go (Microsoft.SecurityInsights/onboardingStates),
//     model_sentinelonboardingstateproperties.go (customerManagedKey)
type OnboardingStates struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*OnboardingStates)(nil)

// NewOnboardingStates returns knowledge for the onboardingStates resource type.
func NewOnboardingStates() *OnboardingStates {
	return &OnboardingStates{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/onboardingStates",
			ApiVersions:  []string{"2022-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.customerManagedKey"},
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM only supports the singleton "default" onboarding state.
					PropertyPath:  "",
					AllowedValues: []string{"default"},
					Message:       "onboarding state name must be \"default\"",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.customerManagedKey", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewOnboardingStates()) }
