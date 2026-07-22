// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterAutomation provides resource knowledge for
// Microsoft.Security/automations.
//
// Mirrors azurerm_security_center_automation. Resource-group scoped; name and
// location are envelope-owned and ForceNew (location encoded below as a body
// top-level path). Body carries properties.{isEnabled,description,scopes,actions,sources}.
//
// The action trigger_url / connection_string are Sensitive but live at
// properties.actions[*].* (array-element paths) — not expressible as scalar
// SensitiveFields, so they are documented here rather than encoded. The
// event_source / operator / property_type / action type enums are likewise
// array-element validators and are left to the schema layer.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_automation_resource.go
//     (schema lines 50-202: name+location ForceNew, enabled default true, scopes/action/source Required min 1;
//     Create/Update 30m, Read 5m)
//   - go-azure-sdk resource-manager/security/2019-01-01-preview/automations:
//     id_automation.go (segment "automations"), model_automationproperties.go
//     (json isEnabled/description/scopes/actions/sources)
type SecurityCenterAutomation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterAutomation)(nil)

// NewSecurityCenterAutomation returns knowledge for the automations resource.
func NewSecurityCenterAutomation() *SecurityCenterAutomation {
	return &SecurityCenterAutomation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/automations",
			ApiVersions:  []string{"2019-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			RequiredFields: []string{
				"properties.scopes",
				"properties.actions",
				"properties.sources",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isEnabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterAutomation()) }
