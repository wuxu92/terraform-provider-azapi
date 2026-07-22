package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageManagementPolicy provides resource knowledge for
// Microsoft.Storage/storageAccounts/managementPolicies.
//
// In AzureRM this is the azurerm_storage_management_policy resource. It is an ARM
// singleton — the policy is always PUT to managementPolicies/default (see the SDK
// CreateOrUpdate path below) — modelled as a sub-resource of the storage account.
//
// The entire lifecycle-rule configuration (rule name/enabled, filters,
// blob_types, match_blob_index_tag, actions with tier/delete/version/snapshot
// day thresholds) lives under properties.policy.rules[*], i.e. array-element
// paths. azwise does not represent array-element paths, so the numerous nested
// StringInSlice / day-threshold validators are intentionally not encoded here;
// only the account-level knowledge (type, api version, timeouts) is captured.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_management_policy_resource.go:26-307
//     (schema), :309-358 (create → managementpolicies.CreateOrUpdate)
//   - go-azure-sdk resource-manager/storage/2025-08-01/managementpolicies:
//     method_createorupdate.go (path .../managementPolicies/default),
//     model_managementpolicyproperties.go, model_managementpolicyschema.go
type StorageManagementPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageManagementPolicy)(nil)

// NewStorageManagementPolicy returns knowledge for the managementPolicies sub-resource.
func NewStorageManagementPolicy() *StorageManagementPolicy {
	return &StorageManagementPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts/managementPolicies",
			ApiVersions:  []string{"2025-08-01"},
			// storage_account_id is ForceNew (schema L45-50) but is the envelope
			// parent reference, not a body field. The rule set (properties.policy.rules)
			// is Required, but its content is entirely array-element paths (see doc).
			RequiredFields: []string{
				"properties.policy.rules",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewStorageManagementPolicy()) }
