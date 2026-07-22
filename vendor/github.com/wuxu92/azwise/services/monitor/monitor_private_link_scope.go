package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateLinkScope provides resource knowledge for
// Microsoft.Insights/privateLinkScopes.
//
// Mirrors azurerm_monitor_private_link_scope.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_private_link_scope_resource.go
//     Schema (42-67): name Required+ForceNew (PrivateLinkScopeName); ingestion_access_mode /
//     query_access_mode Optional Default "Open" enum Open/PrivateOnly →
//     properties.accessModeSettings.ingestionAccessMode / .queryAccessMode;
//     resource_group_name; tags. CreateUpdate (100-110) hardcodes location "Global".
//     Timeouts Create/Update/Delete 30m, Read 5m.
//   - internal/services/monitor/validate/private_link_scope_name.go: name 1-255 chars,
//     alphanumeric plus periods/underscores/hyphens/parenthesis, cannot end in a period.
//   - go-azure-sdk resource-manager/insights/2021-07-01-preview/privatelinkscopesapis:
//     AzureMonitorPrivateLinkScopeProperties.accessModeSettings (ingestionAccessMode /
//     queryAccessMode non-omitempty = required); provisioningState /
//     privateEndpointConnections read-only; AccessMode{Open,PrivateOnly};
//     id_privatelinkscope.go type Microsoft.Insights/privateLinkScopes.
type PrivateLinkScope struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateLinkScope)(nil)

// NewPrivateLinkScope returns knowledge for the privateLinkScopes resource.
func NewPrivateLinkScope() *PrivateLinkScope {
	return &PrivateLinkScope{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/privateLinkScopes",
			ApiVersions:  []string{"2021-07-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: PrivateLinkScopeName.
					Regex:     `^[a-zA-Z0-9()-_.]*[a-zA-Z0-9_-]$`,
					MinLength: 1,
					MaxLength: 255,
					Message:   "name must be 1-255 characters, alphanumeric plus periods, underscores, hyphens and parenthesis, and cannot end in a period",
				},
				{
					PropertyPath:  "properties.accessModeSettings.ingestionAccessMode",
					AllowedValues: []string{"Open", "PrivateOnly"},
					Message:       "ingestion_access_mode must be one of Open, PrivateOnly",
				},
				{
					PropertyPath:  "properties.accessModeSettings.queryAccessMode",
					AllowedValues: []string{"Open", "PrivateOnly"},
					Message:       "query_access_mode must be one of Open, PrivateOnly",
				},
			},
			// accessModeSettings and its two access modes are non-omitempty (required by ARM).
			RequiredFields: []string{
				"properties.accessModeSettings.ingestionAccessMode",
				"properties.accessModeSettings.queryAccessMode",
			},
			// Both access modes default to "Open".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.accessModeSettings.ingestionAccessMode", Value: "Open"},
				{PropertyPath: "properties.accessModeSettings.queryAccessMode", Value: "Open"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.privateEndpointConnections",
			},
		},
	}
}

func init() { azwise.Register(NewPrivateLinkScope()) }
