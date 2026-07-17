package azwise

import "time"

// KustoClusterDatabase provides resource knowledge for
// Microsoft.Kusto/clusters/databases.
//
// The ARM body is a DISCRIMINATED ROOT keyed by `kind` (ReadWrite /
// ReadOnlyFollowing). The native generator emits one Optional variant block per
// discriminator value (read_write / read_only_following) plus a synthesized
// ExactlyOneOf over them; the mapper flattens the selected variant's `properties`
// object directly into the tagged ARM body (kind + properties.*). So a ReadWrite
// database serialises as {location, kind:"ReadWrite", properties:{...}} — the
// property paths below (properties.hotCachePeriod, properties.softDeletePeriod)
// resolve against the go-azure-sdk ReadWriteDatabaseProperties model after that
// flattening, and are equally valid for the ReadOnlyFollowing variant (same field
// shapes and ISO8601 semantics).
//
// azurerm_kusto_database models only the ReadWrite variant.
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_database_resource.go:26-87
//     (schema: ForceNew name/cluster_name/location, timeouts, ISO8601 validators,
//     computed size)
//   - terraform-provider-azurerm internal/services/kusto/kusto_database_resource.go:113-116,195-207
//     (create maps to databases.ReadWriteDatabase{Location, Properties:
//     ReadWriteDatabaseProperties{SoftDeletePeriod, HotCachePeriod}})
//   - terraform-provider-azurerm internal/services/kusto/kusto_database_resource.go:160-172
//     (read/flatten: hotCachePeriod, softDeletePeriod, statistics.size)
//   - terraform-provider-azurerm internal/services/kusto/validate/name.go:61-77
//     (DatabaseName: charclass regex + max length 260; not-all-whitespace check)
//   - terraform-provider-azurerm helpers/validate/time.go:16-27
//     (ISO8601Duration validator applied to hot_cache_period / soft_delete_period)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/databases:
//     model_readwritedatabase.go:13-23 (base: location, kind),
//     model_readwritedatabaseproperties.go:6-14 (hotCachePeriod, softDeletePeriod,
//     isFollowed, provisioningState, statistics, suspensionDetails,
//     keyVaultProperties), constants.go (enums are all read-only)
//   - internal/native/services/kusto/kusto_database_gen.go:296-414 (ReadWrite
//     variant schema flags), :432-434 (generator-synthesized ExactlyOneOf)
//
// Intentionally skipped here (documented, not emitted):
//   - name / cluster_name ForceNew: envelope-owned. `name` (RequiresReplace) and
//     the parent `cluster_id` (RequiresReplace) are already replace-forcing in the
//     generated native schema; they are not Microsoft.Kusto/clusters/databases body
//     properties, so no body ForceNew rule is repeated for them.
//   - ExactlyOneOf(read_write, read_only_following): synthesized by the generator
//     from the discriminated root (kusto_database_gen.go:432-434). Its members are
//     Terraform variant block names, not ARM body dot-paths (both flatten onto the
//     same `properties` object tagged by `kind`), so it cannot be expressed as an
//     azwise RelationalRule and must not be duplicated here.
//   - is_followed / provisioning_state / statistics / suspension_details (ReadWrite)
//     and attached_database_configuration_name / database_share_origin /
//     leader_cluster_resource_id / original_database_name / principals_modification_kind
//     / provisioning_state / soft_delete_period / statistics / suspension_details
//     (ReadOnlyFollowing): all bicep ReadOnly (Computed in the generated schema),
//     so the generator already marks them Computed and strips them from the PUT
//     body — no azwise ComputedFields entry is needed (same rationale as
//     virtual_network.go's resourceGuid/provisioningState).
//   - hot_cache_period / soft_delete_period defaults: Optional+Computed with no
//     explicit AzureRM default (server decides), so no DefaultValue is emitted.
//   - The DatabaseName "must not consist of whitespace only" rule (name.go:64-66)
//     is a negative constraint (^[\s]+$ must NOT match) that RE2 cannot express as
//     a single positive pattern (no lookahead); the positive charclass regex below
//     still admits an all-whitespace name. This residual belongs in a customizer/
//     hook if strict parity is required.
//   - key_vault_properties (ReadWrite CMK block: keyName/keyVaultUri/keyVersion/
//     userIdentity/federatedIdentityClientId) carries no AzureRM validator on
//     azurerm_kusto_database (the field is not surfaced there), so no rule is added.
type KustoClusterDatabase struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*KustoClusterDatabase)(nil)

// NewKustoClusterDatabase returns knowledge for the clusters/databases resource.
func NewKustoClusterDatabase() *KustoClusterDatabase {
	return &KustoClusterDatabase{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/databases",
			ApiVersions:  []string{"2025-02-14"},
			// AzureRM marks location Required + ForceNew (commonschema.Location()); a
			// database's location is pinned to its cluster, so changing it replaces the
			// resource. name and cluster (parent) are envelope-owned RequiresReplace
			// (see doc). The Required schema flag is promoted in the customizer
			// (RequiredFields below is planner metadata, not a native schema flag).
			ForceNew: []ForceNewRule{
				{PropertyPath: "location"},
			},
			SoftDelete: false,
			TimeoutsConfig: &Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []StringRule{
				// Resource name (PropertyPath == ""): DatabaseName validator.
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9\s._-]+$`,
					MaxLength:    260,
					Message:      "must only contain alphanumeric characters, whitespaces, dashes, underscores and dots, and be at most 260 characters",
				},
				// hot_cache_period / soft_delete_period are ISO8601 durations
				// (validate.ISO8601Duration -> period.Parse). Paths resolve against the
				// flattened ReadWriteDatabaseProperties (also valid for ReadOnlyFollowing).
				{
					PropertyPath: "properties.hotCachePeriod",
					Regex:        `^P(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T(\d+H)?(\d+M)?(\d+S)?)?$`,
					Message:      "must be an ISO8601 duration (e.g. P30D)",
				},
				{
					PropertyPath: "properties.softDeletePeriod",
					Regex:        `^P(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T(\d+H)?(\d+M)?(\d+S)?)?$`,
					Message:      "must be an ISO8601 duration (e.g. P31D)",
				},
			},
			IntRules:        []IntRule{},
			FloatRules:      []FloatRule{},
			ArrayRules:      []ArrayRule{},
			SensitiveFields: []string{},
			ComputedFields:  []string{},
			DefaultValues:   []DefaultValue{},
			// location is Required in AzureRM (commonschema.Location()); the schema flag
			// is promoted in the kusto_cluster_database customizer.
			RequiredFields: []string{"location"},
		},
	}
}
