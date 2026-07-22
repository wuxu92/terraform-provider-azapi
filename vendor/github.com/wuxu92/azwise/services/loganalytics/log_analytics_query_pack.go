// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsQueryPack provides resource knowledge for
// Microsoft.OperationalInsights/queryPacks.
//
// Contributing Terraform resource: azurerm_log_analytics_query_pack.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_query_pack_resource.go
//     (Arguments L45-60, Create body L91-95)
//   - go-azure-sdk resource-manager/operationalinsights/2019-09-01/querypacks:
//     model_loganalyticsquerypack.go, model_loganalyticsquerypackproperties.go
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name is Required + ForceNew and
//     validated only with StringIsNotEmpty (no declarative regex/enum), so no name rule.
//   - properties carries no user-settable body fields (AzureRM sends an empty
//     LogAnalyticsQueryPackProperties), so no body rules are emitted.
type LogAnalyticsQueryPack struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsQueryPack)(nil)

// NewLogAnalyticsQueryPack returns knowledge for the queryPacks resource.
func NewLogAnalyticsQueryPack() *LogAnalyticsQueryPack {
	return &LogAnalyticsQueryPack{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/queryPacks",
			ApiVersions:  []string{"2019-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsQueryPack()) }
