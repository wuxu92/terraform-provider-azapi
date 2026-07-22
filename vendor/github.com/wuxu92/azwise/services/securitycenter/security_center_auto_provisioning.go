// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterAutoProvisioning provides resource knowledge for
// Microsoft.Security/autoProvisioningSettings.
//
// Mirrors azurerm_security_center_auto_provisioning (deprecated in AzureRM v5.0).
// Subscription-scoped singleton — the only valid name is "default", which cannot
// be created under another name nor deleted. The body carries a single enum
// property properties.autoProvision (On/Off).
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_auto_provisioning_resource.go
//     (schema lines 47-56: auto_provision Required enum On/Off; create id "default"; Create/Update/Delete 60m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/parse/auto_provisioning_setting.go
//     (ID .../providers/Microsoft.Security/autoProvisioningSettings/default)
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security AutoProvisioningSettingProperties.AutoProvision
//     (json "autoProvision", enum AutoProvisionOn/AutoProvisionOff). Preview SDK: no ARM API version pinned.
type SecurityCenterAutoProvisioning struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterAutoProvisioning)(nil)

// NewSecurityCenterAutoProvisioning returns knowledge for the autoProvisioningSettings resource.
func NewSecurityCenterAutoProvisioning() *SecurityCenterAutoProvisioning {
	return &SecurityCenterAutoProvisioning{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/autoProvisioningSettings",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			RequiredFields: []string{
				"properties.autoProvision",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.autoProvision",
					AllowedValues: []string{"On", "Off"},
					Message:       "autoProvision must be On or Off",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterAutoProvisioning()) }
