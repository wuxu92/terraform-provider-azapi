package dataprotection

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ResourceGuard provides resource knowledge for
// Microsoft.DataProtection/resourceGuards (azurerm_data_protection_resource_guard).
//
// Sources:
//   - AzureRM internal/services/dataprotection/data_protection_resource_guard_resource.go
//     :35-40  (timeouts create/update/delete 30m, read 5m)
//     :47-69  (schema: name validator, vault_critical_operation_exclusion_list)
//     :95-101 (expand: vaultCriticalOperationExclusionList)
//   - AzureRM internal/services/dataprotection/validate/resource_guard_name.go
//     :8-21   (name: non-empty, max 260 characters)
//   - go-azure-sdk resource-manager/dataprotection/2025-07-01/resourceguardresources:
//     model_resourceguard.go, model_resourceguardresource.go, constants.go
type ResourceGuard struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ResourceGuard)(nil)

// NewResourceGuard returns knowledge for the resourceGuards resource.
func NewResourceGuard() *ResourceGuard {
	return &ResourceGuard{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataProtection/resourceGuards",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			// name and resource_group_name are envelope-owned. location is ForceNew
			// (commonschema.Location).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: non-empty, up to 260 characters.
					MinLength: 1,
					MaxLength: 260,
					Message:   "DataProtection ResourceGuard name cannot be empty and must not exceed 260 characters",
				},
			},
			// Response-only properties Azure populates in the GET body.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceGuardOperations",
			},
		},
	}
}

func init() { azwise.Register(NewResourceGuard()) }
