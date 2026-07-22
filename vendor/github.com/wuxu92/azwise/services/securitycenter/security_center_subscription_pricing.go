// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterSubscriptionPricing provides resource knowledge for
// Microsoft.Security/pricings.
//
// Mirrors azurerm_security_center_subscription_pricing. Subscription-scoped; the
// resource name (resource_type, e.g. VirtualMachines) is user-specified. Body:
// properties.pricingTier (Free/Standard, required), properties.subPlan (ForceNew),
// properties.extensions[].
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_subscription_pricing_resource.go
//     (schema lines 51-112: tier Required enum, subplan ForceNew, extension set;
//     create body lines 128-195; Create/Update/Delete 60m, Read 5m)
//   - go-azure-sdk resource-manager/security/2023-01-01/pricings:
//     id_pricing.go (segment "pricings", PricingName user-specified),
//     model_pricingproperties.go (json pricingTier/subPlan/extensions),
//     constants.go PricingTier (Free/Standard)
type SecurityCenterSubscriptionPricing struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterSubscriptionPricing)(nil)

// NewSecurityCenterSubscriptionPricing returns knowledge for the pricings resource.
func NewSecurityCenterSubscriptionPricing() *SecurityCenterSubscriptionPricing {
	return &SecurityCenterSubscriptionPricing{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/pricings",
			ApiVersions:  []string{"2023-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.subPlan"},
			},
			RequiredFields: []string{
				"properties.pricingTier",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.pricingTier",
					AllowedValues: []string{"Free", "Standard"},
					Message:       "pricingTier must be Free or Standard",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterSubscriptionPricing()) }
