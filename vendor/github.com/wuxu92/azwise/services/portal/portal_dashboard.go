// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package portal

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Dashboard provides resource knowledge for Microsoft.Portal/dashboards.
//
// Mirrors azurerm_portal_dashboard. dashboard_properties is a raw JSON string
// unmarshalled directly into the freeform properties object (lenses/metadata),
// so no scalar body rules apply beyond the resource name.
//
// Sources:
//   - terraform-provider-azurerm internal/services/portal/portal_dashboard_resource.go
//     Schema() lines 44-64, CreateUpdate() lines 68-114 (Dashboard{Location, Tags,
//     Properties: DashboardProperties})
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - internal/services/portal/validate/dashboard_name.go (<=160 chars, ^[-\w]+$)
//   - go-azure-sdk resource-manager/portal/2019-01-01-preview/dashboard
//     id_dashboard.go: /providers/Microsoft.Portal/dashboards/{dashboardName}
type Dashboard struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Dashboard)(nil)

// NewDashboard returns knowledge for the Portal dashboards resource.
func NewDashboard() *Dashboard {
	return &Dashboard{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Portal/dashboards",
			ApiVersions:  []string{"2019-01-01-preview"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name/resource_group are envelope.
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name: validate.DashboardName (<=160 chars, alphanumeric + hyphen).
				{
					PropertyPath: "",
					Regex:        `^[-\w]+$`,
					MaxLength:    160,
					Message:      "name must be at most 160 characters of alphanumeric and hyphen characters",
				},
			},
			// dashboard_properties maps to the entire freeform properties object; its
			// content is validated by AzureRM only as well-formed JSON, so there is no
			// scalar RequiredFields/StringRules entry for it.
		},
	}
}

func init() { azwise.Register(NewDashboard()) }
