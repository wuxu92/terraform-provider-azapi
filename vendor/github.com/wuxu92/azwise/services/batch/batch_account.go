package batch

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BatchAccount provides resource knowledge for Microsoft.Batch/batchAccounts.
//
// Contributing Terraform resource: azurerm_batch_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/batch/batch_account_resource.go
//     (schema L51-187, Create body L233-297, expand helpers L505-571)
//   - terraform-provider-azurerm internal/services/batch/validate/account_name.go
//   - go-azure-sdk resource-manager/batch/2024-07-01/batchaccount:
//     model_batchaccountcreateproperties.go (Create body: allowedAuthenticationModes,
//     autoStorage, encryption, keyVaultReference, networkProfile, poolAllocationMode,
//     publicNetworkAccess), model_batchaccountproperties.go (GET-only fields),
//     constants.go (enums).
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; only name is ForceNew and it
//     is not an ARM-body property, so no body ForceNew rules are emitted (no create-body
//     property is ForceNew).
//   - public_network_access_enabled (bool) maps to properties.publicNetworkAccess (enum
//     Enabled/Disabled); AzureRM defaults it to Enabled.
//   - key_vault_reference is Required only when pool_allocation_mode == UserSubscription —
//     a value-conditional requirement that cannot be expressed declaratively, left as a note.
//   - primary_access_key / secondary_access_key are Computed read-only account keys returned
//     via a separate keys API, not create-body fields, so they are not encoded here.
type BatchAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BatchAccount)(nil)

// NewBatchAccount returns knowledge for the batchAccounts resource.
func NewBatchAccount() *BatchAccount {
	return &BatchAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Batch/batchAccounts",
			ApiVersions:  []string{"2024-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.AccountName
				{
					Regex:     `^[a-z0-9]+$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "may contain only lowercase letters and numbers (3-24 chars)",
				},
				// ── storage_account_authentication_mode → properties.autoStorage.authenticationMode
				{
					PropertyPath:  "properties.autoStorage.authenticationMode",
					AllowedValues: []string{"StorageKeys", "BatchAccountManagedIdentity"},
					Message:       "must be one of StorageKeys or BatchAccountManagedIdentity",
				},
				// ── pool_allocation_mode → properties.poolAllocationMode
				{
					PropertyPath:  "properties.poolAllocationMode",
					AllowedValues: []string{"BatchService", "UserSubscription"},
					Message:       "must be one of BatchService or UserSubscription",
				},
				// ── public_network_access → properties.publicNetworkAccess (full ARM SDK set)
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled", "SecuredByPerimeter"},
					Message:       "must be one of Enabled, Disabled, or SecuredByPerimeter",
				},
				// ── encryption.key_source → properties.encryption.keySource
				{
					PropertyPath:  "properties.encryption.keySource",
					AllowedValues: []string{"Microsoft.Batch", "Microsoft.KeyVault"},
					Message:       "must be one of Microsoft.Batch or Microsoft.KeyVault",
				},
				// ── allowed_authentication_modes[*] → properties.allowedAuthenticationModes[*]
				{
					PropertyPath:  "properties.allowedAuthenticationModes[*]",
					AllowedValues: []string{"SharedKey", "AAD", "TaskAuthenticationToken"},
					Message:       "each mode must be one of SharedKey, AAD, or TaskAuthenticationToken",
				},
				// ── network_profile.account_access.default_action
				{
					PropertyPath:  "properties.networkProfile.accountAccess.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be one of Allow or Deny",
				},
				// ── network_profile.node_management_access.default_action
				{
					PropertyPath:  "properties.networkProfile.nodeManagementAccess.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be one of Allow or Deny",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// pool_allocation_mode Default BatchService.
				{PropertyPath: "properties.poolAllocationMode", Value: "BatchService"},
				// public_network_access_enabled Default true → Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// network_profile endpoint access default_action Default Deny (only present
				// when the network_profile block is set).
				{PropertyPath: "properties.networkProfile.accountAccess.defaultAction", Value: "Deny"},
				{PropertyPath: "properties.networkProfile.nodeManagementAccess.defaultAction", Value: "Deny"},
			},
			// storage_account_id and storage_account_authentication_mode are mutually
			// RequiredWith (AzureRM schema); when the autoStorage object carries one, it must
			// carry the other. storage_account_node_identity also RequiresWith storage_account_id.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.autoStorage.storageAccountId", "properties.autoStorage.authenticationMode"},
					Message: "storageAccountId and authenticationMode must be set together",
				},
				{
					Paths:   []string{"properties.autoStorage.nodeIdentityReference", "properties.autoStorage.storageAccountId"},
					Message: "node identity requires storageAccountId to be set",
				},
			},
			// Read-only quota/state/endpoint properties returned by GET but absent from the
			// create body (BatchAccountCreateProperties).
			ComputedFields: []string{
				"properties.accountEndpoint",
				"properties.nodeManagementEndpoint",
				"properties.activeJobAndJobScheduleQuota",
				"properties.dedicatedCoreQuota",
				"properties.dedicatedCoreQuotaPerVMFamily",
				"properties.dedicatedCoreQuotaPerVMFamilyEnforced",
				"properties.lowPriorityCoreQuota",
				"properties.poolQuota",
				"properties.privateEndpointConnections",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewBatchAccount()) }
