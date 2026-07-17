package azwise

import (
	"strings"
	"time"
)

// DocumentDBDatabaseAccount provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts (azurerm_cosmosdb_account).
//
// It overrides CheckForceNew to add the two value-conditional replacements the
// static ForceNew list cannot express (AzureRM CustomizeDiff): disabling
// analytical storage, and downgrading the backup policy from Continuous to
// Periodic.
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_account_resource.go
//     :204-211  (name StringMatch regex/length)
//     :190-195  (CRUD timeouts: create/update 180m, delete 300m)
//     :219-582  (schema: offer_type, kind, create_mode, minimal_tls_version,
//     default_identity_type, consistency_policy, geo_location,
//     capabilities, backup, analytical_storage, capacity, enums)
//     :142-159  (CustomizeDiff: backup.0.type C->P ForceNew, analytical_storage
//     disable ForceNew, capabilities ForceNewIf)
//     :885-892,1101-1113,2137-2149 (expand: public_network_access & network_acl
//     bypass bool->enum, disableKeyBasedMetadataWriteAccess inversion)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_databaseaccountcreateupdateproperties.go / ...parameters.go (ARM body paths),
//     model_databaseaccountgetproperties.go (read-only response-only fields),
//     model_consistencypolicy.go, model_apiproperties.go, model_capacity.go,
//     model_analyticalstorageconfiguration.go, model_backuppolicy.go, model_location.go,
//     constants.go (enum PossibleValuesFor* value sets)
//
// Not encoded (deliberate):
//   - Connection strings and access keys (primary_key, *_connection_string, etc.)
//     are Sensitive + Computed in AzureRM but are NOT properties of the
//     databaseAccounts body — they are returned by the separate listKeys /
//     listConnectionStrings data-plane APIs. There is no ARM body path to mark
//     Sensitive, so SensitiveFields is empty.
//   - properties.backupPolicy is a discriminated union (Periodic/Continuous). Its
//     sub-properties (tier, backupIntervalInMinutes, backupRetentionIntervalInHours,
//     backupStorageRedundancy) live under variant objects the generator emits as
//     separate blocks with an AtMostOneOf, and BackupPolicy is an interface in the
//     SDK model so a "properties.backupPolicy.*" path cannot be resolved. The
//     backup type/tier/interval/retention/storage-redundancy enum & range rules are
//     therefore left to the variant blocks and not re-emitted here.
//   - capabilities[*].name enum, geo_location/locations[*] (failover_priority
//     IntAtLeast(0), zone_redundant), ip_range_filter/ipRules[*], virtual_network_rule[*]
//     and network_acl_bypass_ids[*] are array-element paths; azwise/azapin cannot
//     lower or resolve a rule through an array element, so they are skipped.
//   - default_identity_type (properties.defaultIdentity) is a free-form string
//     validated by AzureRM with Any(regex "^UserAssignedIdentity(.)+$",
//     StringInSlice[FirstPartyIdentity, SystemAssignedIdentity]). It is not a clean
//     enum, so it is left as a residual customizer validator, not a StringRule.
//   - key_vault_key_id vs managed_hsm_key_id ConflictsWith: managed_hsm_key_id is a
//     deprecated (pre-5.0) Terraform-only alias that expands to the SAME ARM path
//     (properties.keyVaultKeyUri). There is no distinct ARM body path for the
//     conflict, so no RelationalRule is representable.
//   - capabilities ForceNewIf (resource :153-159) forces replacement only when a
//     capability is removed and the kind/removable-set logic disallows in-place
//     removal — value+kind-conditional over an array; left as a residual hook.
type DocumentDBDatabaseAccount struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*DocumentDBDatabaseAccount)(nil)

// CheckForceNew extends BaseKnowledge with AzureRM's value-conditional
// replacements (cosmosdb_account_resource.go:143-151): analytical storage cannot
// be disabled in place, and the backup policy can only move Periodic->Continuous,
// so a Continuous->Periodic change forces replacement.
func (s *DocumentDBDatabaseAccount) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}
	if documentDBBoolValue(oldBody, "properties.enableAnalyticalStorage") &&
		!documentDBBoolValue(newBody, "properties.enableAnalyticalStorage") {
		return true
	}
	oldBackup := extractStringValue(oldBody, "properties.backupPolicy.type")
	newBackup := extractStringValue(newBody, "properties.backupPolicy.type")
	return strings.EqualFold(oldBackup, "Continuous") && strings.EqualFold(newBackup, "Periodic")
}

// NewDocumentDBDatabaseAccount returns knowledge for the databaseAccounts resource.
func NewDocumentDBDatabaseAccount() *DocumentDBDatabaseAccount {
	return &DocumentDBDatabaseAccount{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts",
			ApiVersions:  []string{"2026-03-15"},
			SoftDelete:   false,
			// name and resource_group_name are envelope-owned. location
			// (commonschema.Location) replaces the account on change; kind,
			// create_mode, free_tier_enabled, key_vault_key_id and the whole restore
			// block are unconditionally ForceNew in AzureRM.
			ForceNew: []ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "kind"},
				{PropertyPath: "properties.createMode"},
				{PropertyPath: "properties.enableFreeTier"},
				{PropertyPath: "properties.keyVaultKeyUri"},
				{PropertyPath: "properties.restoreParameters"},
			},
			TimeoutsConfig: &Timeouts{
				Create: 180 * time.Minute,
				Read:   5 * time.Minute,
				Update: 180 * time.Minute,
				Delete: 300 * time.Minute,
			},
			StringRules: []StringRule{
				{
					// Resource name: 3-50 chars, lowercase letters, numbers, hyphens.
					Regex:     `^[-a-z0-9]{3,50}$`,
					MinLength: 3,
					MaxLength: 50,
					Message:   "Cosmos DB Account name must be 3-50 characters long and contain only lowercase letters, numbers and hyphens",
				},
				{
					PropertyPath:  "kind",
					AllowedValues: []string{"GlobalDocumentDB", "MongoDB", "Parse"},
					Message:       "must be one of GlobalDocumentDB, MongoDB or Parse",
				},
				{
					PropertyPath:  "properties.databaseAccountOfferType",
					AllowedValues: []string{"Standard"},
					Message:       "must be Standard",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Default", "Restore"},
					Message:       "must be Default or Restore",
				},
				{
					PropertyPath:  "properties.minimalTlsVersion",
					AllowedValues: []string{"Tls", "Tls11", "Tls12"},
					Message:       "must be one of Tls, Tls11 or Tls12",
				},
				{
					PropertyPath:  "properties.consistencyPolicy.defaultConsistencyLevel",
					AllowedValues: []string{"BoundedStaleness", "ConsistentPrefix", "Eventual", "Session", "Strong"},
					Message:       "must be one of BoundedStaleness, ConsistentPrefix, Eventual, Session or Strong",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled", "SecuredByPerimeter"},
					Message:       "must be one of Disabled, Enabled or SecuredByPerimeter",
				},
				{
					PropertyPath:  "properties.networkAclBypass",
					AllowedValues: []string{"AzureServices", "None"},
					Message:       "must be AzureServices or None",
				},
				{
					PropertyPath:  "properties.apiProperties.serverVersion",
					AllowedValues: []string{"3.2", "3.6", "4.0", "4.2", "5.0", "6.0", "7.0"},
					Message:       "must be a supported MongoDB server version",
				},
				{
					PropertyPath:  "properties.analyticalStorageConfiguration.schemaType",
					AllowedValues: []string{"FullFidelity", "WellDefined"},
					Message:       "must be FullFidelity or WellDefined",
				},
			},
			IntRules: []IntRule{
				{
					PropertyPath: "properties.capacity.totalThroughputLimit",
					MinValue:     ptr(int64(-1)),
					Message:      "must be -1 (unlimited) or a positive throughput limit",
				},
				{
					PropertyPath: "properties.consistencyPolicy.maxIntervalInSeconds",
					MinValue:     ptr(int64(5)),
					MaxValue:     ptr(int64(86400)),
					Message:      "must be between 5 and 86400 seconds",
				},
				{
					PropertyPath: "properties.consistencyPolicy.maxStalenessPrefix",
					MinValue:     ptr(int64(10)),
					MaxValue:     ptr(int64(2147483647)),
					Message:      "must be between 10 and 2147483647",
				},
			},
			// Connection strings / keys are data-plane (listKeys) values, not body
			// properties — see the doc block. No ARM body path to mark Sensitive.
			SensitiveFields: []string{},
			// Response-only properties absent from the create/update model.
			ComputedFields: []string{
				"properties.documentEndpoint",
				"properties.provisioningState",
				"properties.instanceId",
				"properties.failoverPolicies",
				"properties.readLocations",
				"properties.writeLocations",
				"properties.privateEndpointConnections",
			},
			DefaultValues: []DefaultValue{
				{PropertyPath: "kind", Value: "GlobalDocumentDB"},
				{PropertyPath: "properties.createMode", Value: "Default"},
				{PropertyPath: "properties.minimalTlsVersion", Value: "Tls12"},
				{PropertyPath: "properties.defaultIdentity", Value: "FirstPartyIdentity"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.networkAclBypass", Value: "None"},
				{PropertyPath: "properties.enableFreeTier", Value: false},
				{PropertyPath: "properties.enableAnalyticalStorage", Value: false},
				{PropertyPath: "properties.enableAutomaticFailover", Value: false},
				{PropertyPath: "properties.isVirtualNetworkFilterEnabled", Value: false},
				{PropertyPath: "properties.disableKeyBasedMetadataWriteAccess", Value: false},
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				{PropertyPath: "properties.enableMultipleWriteLocations", Value: false},
				{PropertyPath: "properties.enablePartitionMerge", Value: false},
				{PropertyPath: "properties.enableBurstCapacity", Value: false},
				{PropertyPath: "properties.consistencyPolicy.maxIntervalInSeconds", Value: int64(5)},
				{PropertyPath: "properties.consistencyPolicy.maxStalenessPrefix", Value: int64(100)},
			},
			RequiredFields: []string{
				"properties.databaseAccountOfferType",
				"properties.locations",
				"properties.consistencyPolicy",
				"properties.consistencyPolicy.defaultConsistencyLevel",
			},
		},
	}
}

// documentDBBoolValue returns the boolean at path, or false when absent/non-bool.
func documentDBBoolValue(body map[string]interface{}, path string) bool {
	value, ok := extractNestedValue(body, path).(bool)
	return ok && value
}
