// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Metadata provides resource knowledge for
// Microsoft.SecurityInsights/metadata (an extension resource on a
// Microsoft.OperationalInsights/workspaces scope).
//
// Contributing Terraform resource: azurerm_sentinel_metadata.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_metadata_resource.go
//     (schema L75-133, timeouts L353-354)
//   - go-azure-sdk resource-manager/securityinsights/2022-10-01-preview/metadata:
//     id_metadata.go (Microsoft.SecurityInsights/metadata), constants.go (Kind),
//     model_metadataproperties.go (contentId, kind, parentId)
//
// Notes:
//   - parent_id (azure.ValidateResourceID) is a semantic resource-ID check and is
//     not expressed declaratively here.
type Metadata struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Metadata)(nil)

// NewMetadata returns knowledge for the metadata resource type.
func NewMetadata() *Metadata {
	return &Metadata{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/metadata",
			ApiVersions:  []string{"2022-10-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.kind",
				"properties.parentId",
				"properties.contentId",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "metadata name must not be empty",
				},
				{
					PropertyPath: "properties.kind",
					// Full ARM SDK metadata Kind set.
					AllowedValues: []string{
						"AnalyticsRule", "AnalyticsRuleTemplate", "AutomationRule",
						"AzureFunction", "DataConnector", "DataType", "HuntingQuery",
						"InvestigationQuery", "LogicAppsCustomConnector", "Parser",
						"Playbook", "PlaybookTemplate", "Solution", "Watchlist",
						"WatchlistTemplate", "Workbook", "WorkbookTemplate",
					},
					Message: "kind must be a supported Sentinel metadata kind",
				},
				{
					PropertyPath: "properties.contentId",
					MinLength:    1,
					Message:      "content_id must not be empty",
				},
			},
		},
	}
}

func init() { azwise.Register(NewMetadata()) }
