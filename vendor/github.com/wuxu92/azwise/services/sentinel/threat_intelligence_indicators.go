// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ThreatIntelligenceIndicators provides resource knowledge for
// Microsoft.SecurityInsights/threatIntelligence/indicators (an extension
// resource on a Microsoft.OperationalInsights/workspaces scope). The parent
// threatIntelligence resource is the fixed singleton named "main"; that name is
// not part of the ARM resource type.
//
// Contributing Terraform resource: azurerm_sentinel_threat_intelligence_indicator.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_threat_intelligence_indicator_resource.go
//     (schema L107-311, body L433-461, timeouts L389)
//   - go-azure-sdk resource-manager/securityinsights/2022-10-01-preview/threatintelligence:
//     id_indicator.go (threatIntelligence "main" / indicators), constants.go
//     (ThreatIntelligenceResourceKindEnum=indicator),
//     model_threatintelligenceindicatorproperties.go (pattern, patternType,
//     confidence, source, displayName, validFrom, revoked)
//
// Notes:
//   - name (guid) is server-generated (Computed) — no name rule.
//   - pattern_type is an AzureRM-restricted subset; the ARM field is a free string.
type ThreatIntelligenceIndicators struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ThreatIntelligenceIndicators)(nil)

// NewThreatIntelligenceIndicators returns knowledge for the threatIntelligence/indicators resource type.
func NewThreatIntelligenceIndicators() *ThreatIntelligenceIndicators {
	return &ThreatIntelligenceIndicators{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/threatIntelligence/indicators",
			ApiVersions:  []string{"2022-10-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.source"},
			},
			RequiredFields: []string{
				"kind",
				"properties.pattern",
				"properties.patternType",
				"properties.source",
				"properties.displayName",
				"properties.validFrom",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "kind",
					AllowedValues: []string{"indicator"},
					Message:       "kind must be indicator",
				},
				{
					PropertyPath:  "properties.patternType",
					AllowedValues: []string{"domain-name", "file", "ipv4-addr", "ipv6-addr", "url"},
					Message:       "pattern_type must be one of domain-name, file, ipv4-addr, ipv6-addr, url",
				},
				{
					PropertyPath: "properties.source",
					MinLength:    1,
					Message:      "source must not be empty",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					Message:      "display_name must not be empty",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.confidence",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(100)),
					Message:      "confidence must be between 0 and 100",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.revoked", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewThreatIntelligenceIndicators()) }
