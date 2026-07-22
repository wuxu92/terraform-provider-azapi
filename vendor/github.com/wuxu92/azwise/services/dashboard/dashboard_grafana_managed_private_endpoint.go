package dashboard

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DashboardGrafanaManagedPrivateEndpoint provides resource knowledge for
// Microsoft.Dashboard/grafana/managedPrivateEndpoints.
//
// Contributing Terraform resource:
//   - azurerm_dashboard_grafana_managed_private_endpoint
//
// Sources:
//   - terraform-provider-azurerm internal/services/dashboard/dashboard_grafana_managed_private_endpoint_resource.go
//     (schema Arguments L69-131, Create timeout L139 + body L164-175, Read L193,
//     Update L258, Delete L239)
//   - go-azure-sdk resource-manager/dashboard/2025-08-01/managedprivateendpointmodels:
//     model_managedprivateendpointmodel.go, model_managedprivateendpointmodelproperties.go,
//     id_managedprivateendpoint.go (type segment "managedPrivateEndpoints" confirmed L110/L125).
//
// Notes:
//   - grafana_id is the parent Grafana resource ID (envelope/ID), not an ARM body property.
//   - private_link_service_url maps to properties.privateLinkServiceUrl (note the SDK json
//     tag is lowercase "Url", though the Go field is PrivateLinkServiceURL).
//   - name/grafana_id/private_link_resource_id/group_ids/private_link_resource_region are all
//     ForceNew; request_message and private_link_service_url are updatable.
type DashboardGrafanaManagedPrivateEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DashboardGrafanaManagedPrivateEndpoint)(nil)

// NewDashboardGrafanaManagedPrivateEndpoint returns knowledge for the
// Microsoft.Dashboard/grafana/managedPrivateEndpoints resource.
func NewDashboardGrafanaManagedPrivateEndpoint() *DashboardGrafanaManagedPrivateEndpoint {
	return &DashboardGrafanaManagedPrivateEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Dashboard/grafana/managedPrivateEndpoints",
			ApiVersions:  []string{"2025-08-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 5 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.privateLinkResourceId"},
				{PropertyPath: "properties.groupIds"},
				{PropertyPath: "properties.privateLinkResourceRegion"},
			},
			StringRules: []azwise.StringRule{
				// name (resource name; empty PropertyPath). 2-20 chars.
				{
					PropertyPath: "",
					Regex:        `\A([a-zA-Z]{1}[a-zA-Z0-9\-]{1,19}[a-zA-Z0-9]{1})\z`,
					MinLength:    2,
					MaxLength:    20,
					Message:      "name must be 2-20 characters, alphanumeric or dashes, begin with a letter and end with a letter or digit",
				},
				{
					PropertyPath: "properties.privateLinkServiceUrl",
					Regex:        `^([0-9A-Za-z\-]+\.){2,}([0-9A-Za-z\-]+)\.?$`,
					Message:      "the URL must contain at least 3 dot-separated parts of alphanumeric characters and hyphens",
				},
			},
			// Server-populated, read-only ARM properties (absent from create body).
			ComputedFields: []string{
				"properties.connectionState",
				"properties.privateLinkServicePrivateIP",
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.privateLinkResourceId",
			},
		},
	}
}

// Self-registers into the azwise registry.
func init() { azwise.Register(NewDashboardGrafanaManagedPrivateEndpoint()) }
