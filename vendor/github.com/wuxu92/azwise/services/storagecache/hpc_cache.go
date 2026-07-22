package storagecache

import (
	"time"

	"github.com/wuxu92/azwise"
)

// HPCCache provides resource knowledge for Microsoft.StorageCache/caches.
//
// Mirrors azurerm_hpc_cache. The azurerm_hpc_cache_access_policy resource mutates
// the SAME ARM resource (caches): it inserts/updates entries in
// properties.securitySettings.accessPolicies[*] (an array of NFS access policies,
// keyed by policy name). Because that is an array-element mutation with no single
// scalar body path, its per-rule constraints (scope/access enums, anonymous uid/gid
// >= 0) cannot be expressed as declarative rules here and are folded in as the note
// below rather than emitted as a separate file.
//
// Sources:
//   - internal/services/storagecache/hpc_cache_resource.go
//     (resource 33-55: timeouts Create/Update 60m, Read 5m, Delete 60m;
//     schema 713-1029: name ForceNew StringIsNotEmpty; cache_size_in_gb Required
//     ForceNew IntInSlice{3072,6144,12288,21623,24576,43246,49152,86491};
//     subnet_id Required ForceNew; sku_name Required ForceNew StringInSlice(6 SKUs);
//     mtu Optional Default 1500 IntBetween(576,1500); ntp_server Optional Default
//     "time.windows.com"; identity OptionalForceNew; key_vault_key_id +
//     automatically_rotate_key_to_latest_enabled; mount_addresses Computed;
//     create 133-195 → Cache{Properties{CacheSizeGB, Subnet, NetworkSettings,
//     SecuritySettings.AccessPolicies, DirectoryServicesSettings, EncryptionSettings},
//     Sku.Name}; SKU/cache-size combo validation 89-99).
//   - internal/services/storagecache/hpc_cache_access_policy_resource.go
//     (folds into caches securitySettings.accessPolicies[*]; access_rule scope/access
//     enums, anonymous_uid/gid IntAtLeast(0)).
//   - go-azure-sdk resource-manager/storagecache/2023-05-01/caches:
//     model_cacheproperties.go (cacheSizeGB/subnet/networkSettings/securitySettings/
//     directoryServicesSettings/encryptionSettings/mountAddresses/provisioningState),
//     model_cachenetworksettings.go (mtu/ntpServer/dnsServers/dnsSearchDomain),
//     model_cachesku.go (name), model_cacheencryptionsettings.go
//     (keyEncryptionKey/rotationToLatestKeyVersionEnabled),
//     model_keyvaultkeyreference.go (keyUrl/sourceVault),
//     constants.go (NfsAccessRuleScope default/host/network, NfsAccessRuleAccess no/ro/rw),
//     id_cache.go (type segment casing "caches").
type HPCCache struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*HPCCache)(nil)

// NewHPCCache returns knowledge for the caches resource.
func NewHPCCache() *HPCCache {
	return &HPCCache{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageCache/caches",
			ApiVersions:  []string{"2023-05-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.cacheSizeGB"},
				{PropertyPath: "properties.subnet"},
				{PropertyPath: "sku.name"},
				// identity uses SystemAssignedUserAssignedIdentityOptionalForceNew.
				{PropertyPath: "identity"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Standard_2G",
						"Standard_4G",
						"Standard_8G",
						"Standard_L4_5G",
						"Standard_L9G",
						"Standard_L16G",
					},
				},
			},
			IntRules: []azwise.IntRule{
				{
					// mtu: validation.IntBetween(576, 1500).
					PropertyPath: "properties.networkSettings.mtu",
					MinValue:     azwise.Ptr(int64(576)),
					MaxValue:     azwise.Ptr(int64(1500)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.networkSettings.mtu", Value: 1500},
				{PropertyPath: "properties.networkSettings.ntpServer", Value: "time.windows.com"},
			},
			RequiredFields: []string{
				"properties.cacheSizeGB",
				"properties.subnet",
				"sku.name",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.mountAddresses",
				"properties.provisioningState",
			},
			// NOTE: cache_size_in_gb is validated by IntInSlice against a discrete set
			// {3072, 6144, 12288, 21623, 24576, 43246, 49152, 86491} plus a create-time
			// SKU/cache-size combo check (Standard_L4_5G→21623, Standard_L9G→43246,
			// Standard_L16G→86491; 21623/43246/86491 reserved for read-only SKUs). This
			// discrete-set + cross-field constraint is not a numeric range and cannot be
			// expressed as an IntRule.
			// NOTE: default_access_policy maps to properties.securitySettings.accessPolicies[*]
			// (an array of NfsAccessPolicy), the same array mutated by
			// azurerm_hpc_cache_access_policy. Its element rules (access_rule scope in
			// {default,host,network}, access in {no,ro,rw}, anonymous_uid/gid >= 0) are
			// array-element paths and cannot be declarative rules here.
			// NOTE: directory_active_directory / directory_flat_file / directory_ldap are
			// mutually exclusive single blocks mapping to
			// properties.directoryServicesSettings.{activeDirectory,usernameDownload}; their
			// nested validators (netbios regexes ^[-0-9a-zA-Z]{1,15}$, IPv4 addresses)
			// are nested-object paths not expressible declaratively.
			// NOTE: key_vault_key_id maps to
			// properties.encryptionSettings.keyEncryptionKey.keyUrl and
			// automatically_rotate_key_to_latest_enabled to
			// properties.encryptionSettings.rotationToLatestKeyVersionEnabled; key_vault_key_id
			// is a Key Vault nested-item ID (RequiredWith identity) validated semantically,
			// and cannot be added/removed after create — neither is a simple declarative rule.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewHPCCache()) }
