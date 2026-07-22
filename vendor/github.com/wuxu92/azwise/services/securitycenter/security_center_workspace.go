// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterWorkspace provides resource knowledge for
// Microsoft.Security/workspaceSettings.
//
// Mirrors azurerm_security_center_workspace. Subscription-scoped singleton — the
// only valid name is "default". Body carries properties.workspaceId and
// properties.scope, both required.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_workspace_resource.go
//     (schema lines 47-61: scope + workspace_id Required; name const "default"; Create/Update/Delete 60m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/parse/workspace.go
//     (ID .../providers/Microsoft.Security/workspaceSettings/default)
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security WorkspaceSettingProperties
//     (json workspaceId/scope). Preview SDK: no ARM API version pinned.
type SecurityCenterWorkspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterWorkspace)(nil)

// NewSecurityCenterWorkspace returns knowledge for the workspaceSettings resource.
func NewSecurityCenterWorkspace() *SecurityCenterWorkspace {
	return &SecurityCenterWorkspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/workspaceSettings",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			RequiredFields: []string{
				"properties.workspaceId",
				"properties.scope",
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterWorkspace()) }
