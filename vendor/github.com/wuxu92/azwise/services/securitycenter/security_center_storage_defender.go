// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterStorageDefender provides resource knowledge for
// Microsoft.Security/defenderForStorageSettings.
//
// Mirrors azurerm_security_center_storage_defender. Storage-account-scoped
// singleton named "current". Body carries properties.{isEnabled,
// overrideSubscriptionLevelSettings, malwareScanning.onUpload.*,
// sensitiveDataDiscovery.isEnabled}.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_storage_defender_resource.go
//     (arguments lines 49-88; create body lines 121-143; Create/Update 30m, Read 5m, Delete 30m)
//   - go-azure-sdk resource-manager/security/2025-06-01/defenderforstorage:
//     method_create.go (path .../defenderForStorageSettings/current),
//     model_defenderforstoragesettingproperties.go (isEnabled/overrideSubscriptionLevelSettings/malwareScanning/sensitiveDataDiscovery),
//     model_malwarescanningproperties.go (onUpload/scanResultsEventGridTopicResourceId),
//     model_onuploadproperties.go (isEnabled/capGBPerMonth *int64)
type SecurityCenterStorageDefender struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterStorageDefender)(nil)

// NewSecurityCenterStorageDefender returns knowledge for the defenderForStorageSettings resource.
func NewSecurityCenterStorageDefender() *SecurityCenterStorageDefender {
	return &SecurityCenterStorageDefender{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/defenderForStorageSettings",
			ApiVersions:  []string{"2025-06-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.overrideSubscriptionLevelSettings", Value: false},
				{PropertyPath: "properties.malwareScanning.onUpload.isEnabled", Value: false},
				// AzureRM default -1 = uncapped (schema Default: -1).
				{PropertyPath: "properties.malwareScanning.onUpload.capGBPerMonth", Value: float64(-1)},
				{PropertyPath: "properties.sensitiveDataDiscovery.isEnabled", Value: false},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.malwareScanning.onUpload.capGBPerMonth",
					// AzureRM validates IntAtLeast(1) OR the sentinel -1 (uncapped).
					// IntRule cannot express the disjoint gap at 0; the lower bound -1
					// is encoded, so a stray 0 is not rejected here (schema layer handles it).
					MinValue: azwise.Ptr(int64(-1)),
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterStorageDefender()) }
