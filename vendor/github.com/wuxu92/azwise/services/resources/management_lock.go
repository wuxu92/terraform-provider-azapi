package resources

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagementLock provides resource knowledge for Microsoft.Authorization/locks.
//
// Mirrors azurerm_management_lock. The lock is a scoped resource: its name and scope live
// on the ID/envelope, while lock_level and notes form the ARM body under `properties`.
// The resource has no Update — every field is replace-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/resource/management_lock_resource.go:22-98
//     (name=ManagementLockName ForceNew, scope ForceNew, lock_level enum ForceNew,
//     notes StringLenBetween(0,512) ForceNew, 30m/5m/30m timeouts, no Update)
//   - terraform-provider-azurerm internal/services/resource/validate/management_lock_name.go:11-22
//     (name: alphanumeric/dash/underscore, maximum 260 chars — len>=260 rejected)
//   - go-azure-sdk resource-manager/resources/2020-05-01/managementlocks/model_managementlockproperties.go:6-10
//     (ManagementLockProperties{level,notes,owners} — ARM body json tags)
//   - go-azure-sdk resource-manager/resources/2020-05-01/managementlocks/constants.go:12-26
//     (LockLevel: CanNotDelete/NotSpecified/ReadOnly — full SDK set; AzureRM restricts to
//     CanNotDelete/ReadOnly but azwise sends raw ARM values)
//   - go-azure-sdk resource-manager/resources/2020-05-01/managementlocks/id_scopedlock.go:96-110
//     (.../providers/Microsoft.Authorization/locks/{lockName} — ARM type casing)
//
// Intentionally skipped (documented, no rule emitted):
//   - scope: the target resource ID the lock is applied to; lives on the ID, not the body.
type ManagementLock struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagementLock)(nil)

// NewManagementLock returns knowledge for the locks resource.
func NewManagementLock() *ManagementLock {
	return &ManagementLock{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/locks",
			ApiVersions:  []string{"2020-05-01"},
			SoftDelete:   false,
			// No Update function: the body properties force replacement.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.level"},
				{PropertyPath: "properties.notes"},
			},
			RequiredFields: []string{
				"properties.level",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM: alphanumeric, dashes and
					// underscores; maximum 260 characters (len>=260 is rejected, so 259 max).
					MaxLength: 259,
					Message:   "name may only contain alphanumeric characters, dashes and underscores, and be at most 260 characters",
				},
				{
					// lock_level -> properties.level. Full ARM SDK set.
					PropertyPath:  "properties.level",
					AllowedValues: []string{"CanNotDelete", "NotSpecified", "ReadOnly"},
					Message:       "lock_level must be one of CanNotDelete, NotSpecified, ReadOnly",
				},
				{
					// notes -> properties.notes. AzureRM StringLenBetween(0, 512).
					PropertyPath: "properties.notes",
					MaxLength:    512,
					Message:      "notes must be at most 512 characters",
				},
			},
		},
	}
}

func init() { azwise.Register(NewManagementLock()) }
