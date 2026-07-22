// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AdvancedThreatProtection provides resource knowledge for
// Microsoft.Security/advancedThreatProtectionSettings.
//
// Mirrors azurerm_advanced_threat_protection. This is an extension-style
// singleton named "current" that hangs off any target resource (the
// target_resource_id is the parent scope, not part of the body). The only body
// property is properties.isEnabled.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/advanced_threat_protection_resource.go
//     (schema lines 48-60: target_resource_id ForceNew scope + enabled Required; Create/Update 30m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/parse/advanced_threat_protection.go
//     (ID .../providers/Microsoft.Security/advancedThreatProtectionSettings/current)
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security AdvancedThreatProtectionProperties.IsEnabled
//     (json "isEnabled"). Data-plane preview SDK: no ARM API version pinned.
type AdvancedThreatProtection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AdvancedThreatProtection)(nil)

// NewAdvancedThreatProtection returns knowledge for the advancedThreatProtectionSettings resource.
func NewAdvancedThreatProtection() *AdvancedThreatProtection {
	return &AdvancedThreatProtection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/advancedThreatProtectionSettings",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.isEnabled",
			},
		},
	}
}

func init() { azwise.Register(NewAdvancedThreatProtection()) }
