// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterSetting provides resource knowledge for
// Microsoft.Security/settings.
//
// Mirrors azurerm_security_center_setting. Subscription-scoped; the resource
// name (setting_name) is a fixed enum and ForceNew. The body carries a single
// property properties.enabled (present on both DataExportSettings and
// AlertSyncSettings kinds).
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_setting_resource.go
//     (schema lines 48-58: setting_name Required ForceNew enum, enabled Required; Create/Update/Delete 10m, Read 5m)
//   - go-azure-sdk resource-manager/security/2022-05-01/settings:
//     id_setting.go (segment "settings"), constants.go SettingName
//     (current/MCAS/Sentinel/WDATP/WDATP_EXCLUDE_LINUX_PUBLIC_PREVIEW/WDATP_UNIFIED_SOLUTION),
//     model_dataexportsettingproperties.go / model_alertsyncsettingproperties.go (json "enabled")
type SecurityCenterSetting struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterSetting)(nil)

// NewSecurityCenterSetting returns knowledge for the settings resource.
func NewSecurityCenterSetting() *SecurityCenterSetting {
	return &SecurityCenterSetting{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/settings",
			ApiVersions:  []string{"2022-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 10 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			RequiredFields: []string{
				"properties.enabled",
			},
			StringRules: []azwise.StringRule{
				{
					// resource name attribute (PropertyPath == "")
					AllowedValues: []string{
						"current",
						"MCAS",
						"Sentinel",
						"WDATP",
						"WDATP_EXCLUDE_LINUX_PUBLIC_PREVIEW",
						"WDATP_UNIFIED_SOLUTION",
					},
					Message: "setting name must be one of the supported Security Center setting names",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterSetting()) }
