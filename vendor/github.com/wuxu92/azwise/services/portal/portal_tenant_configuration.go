// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package portal

import (
	"time"

	"github.com/wuxu92/azwise"
)

// TenantConfiguration provides resource knowledge for
// Microsoft.Portal/tenantConfigurations.
//
// Mirrors azurerm_portal_tenant_configuration. This is a tenant-scoped singleton
// named "default" (no resource group or location); the only configurable value is
// private_markdown_storage_enforced -> properties.enforcePrivateMarkdownStorage.
//
// Sources:
//   - terraform-provider-azurerm internal/services/portal/portal_tenant_configuration_resource.go
//     Schema() lines 41-46, CreateUpdate() lines 50-88 (Configuration{Properties:
//     ConfigurationProperties{EnforcePrivateMarkdownStorage}})
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/portal/2019-01-01-preview/tenantconfiguration
//     model_configurationproperties.go (enforcePrivateMarkdownStorage) /
//     method_*.go: PUT/GET/DELETE /providers/Microsoft.Portal/tenantConfigurations/default
type TenantConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*TenantConfiguration)(nil)

// NewTenantConfiguration returns knowledge for the Portal tenantConfigurations
// singleton resource.
func NewTenantConfiguration() *TenantConfiguration {
	return &TenantConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Portal/tenantConfigurations",
			ApiVersions:  []string{"2019-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				// private_markdown_storage_enforced is Required in TF.
				"properties.enforcePrivateMarkdownStorage",
			},
		},
	}
}

func init() { azwise.Register(NewTenantConfiguration()) }
