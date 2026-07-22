package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoClusterAttachedDatabaseConfiguration provides resource knowledge for
// Microsoft.Kusto/clusters/attachedDatabaseConfigurations.
//
// Mirrors azurerm_kusto_attached_database_configuration. name / cluster_name /
// resource_group_name live on the operational envelope; name is ForceNew
// (RequiresReplace) and surfaced only via the resource-name StringRule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_attached_database_configuration_resource.go:30-220
//     (schema: ForceNew name/cluster_name/database_name/cluster_id, default_principal_modification_kind
//     default+enum, sharing block arrays, ConflictsWith override/prefix, CustomizeDiff "*" guard, timeouts)
//   - terraform-provider-azurerm internal/services/kusto/kusto_attached_database_configuration_resource.go:335-402
//     (expand: clusterResourceId, databaseName, databaseNameOverride, databaseNamePrefix,
//     defaultPrincipalsModificationKind, tableLevelSharingProperties.*)
//   - terraform-provider-azurerm internal/services/kusto/validate/name.go:11-27
//     (DataConnectionName: charclass regex + max length 40; used for `name`)
//   - terraform-provider-azurerm internal/services/kusto/validate/name.go:61-77
//     (DatabaseName: charclass regex + max length 260; used for override/prefix)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/attacheddatabaseconfigurations:
//     model_attacheddatabaseconfiguration.go:6-12 (envelope: location, name, properties),
//     model_attacheddatabaseconfigurationproperties.go:6-15 (properties.* body paths),
//     model_tablelevelsharingproperties.go:6-15 (tableLevelSharingProperties.* arrays),
//     id_attacheddatabaseconfiguration.go:110-127 (ARM path casing:
//     clusters/attachedDatabaseConfigurations),
//     constants.go:50-64 (DefaultPrincipalsModificationKind: None/Replace/Union)
//
// Not encoded (deliberate):
//   - CustomizeDiff: `database_name == "*" && database_name_override != ""` is rejected
//     (resource.go:47-55). This is a value-conditional cross-field guard (a specific
//     literal value of one field forbids another being set), not a presence relation, so
//     it cannot be expressed as an azwise RelationalRule. Left to a customizer/hook.
//   - database_name uses validation.Any(DatabaseName, "*"): the "*" (all databases)
//     escape hatch means the DatabaseName charclass regex does NOT universally apply to
//     properties.databaseName, so no StringRule is emitted for it (override/prefix, which
//     do not accept "*", carry the regex instead).
//   - tableLevelSharingProperties.{tablesToInclude,...} are string arrays with no element
//     validator (utils.ExpandStringSlice); azwise cannot lower array-element constraints,
//     and there are none here anyway.
//   - attached_database_names (properties.attachedDatabaseNames) and provisioningState are
//     server-computed read-only (see ComputedFields).
type KustoClusterAttachedDatabaseConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoClusterAttachedDatabaseConfiguration)(nil)

// NewKustoClusterAttachedDatabaseConfiguration returns knowledge for the
// clusters/attachedDatabaseConfigurations resource.
func NewKustoClusterAttachedDatabaseConfiguration() *KustoClusterAttachedDatabaseConfiguration {
	return &KustoClusterAttachedDatabaseConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/attachedDatabaseConfigurations",
			ApiVersions:  []string{"2025-02-14"},
			// cluster_id -> properties.clusterResourceId and database_name ->
			// properties.databaseName are Required + ForceNew. location
			// (commonschema.Location()) is Required + ForceNew (envelope).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.clusterResourceId"},
				{PropertyPath: "properties.databaseName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name (PropertyPath == ""): DataConnectionName validator.
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9\s.-]+$`,
					MaxLength:    40,
					Message:      "must only contain letters, digits, whitespaces, dashes and dots, and be at most 40 characters",
				},
				{
					PropertyPath: "properties.databaseNameOverride",
					Regex:        `^[a-zA-Z0-9\s._-]+$`,
					MaxLength:    260,
					Message:      "must only contain alphanumeric characters, whitespaces, dashes, underscores and dots, and be at most 260 characters",
				},
				{
					PropertyPath: "properties.databaseNamePrefix",
					Regex:        `^[a-zA-Z0-9\s._-]+$`,
					MaxLength:    260,
					Message:      "must only contain alphanumeric characters, whitespaces, dashes, underscores and dots, and be at most 260 characters",
				},
				{
					PropertyPath:  "properties.defaultPrincipalsModificationKind",
					AllowedValues: []string{"None", "Replace", "Union"},
					Message:       "must be None, Replace or Union",
				},
			},
			// database_name_override and database_name_prefix are mutually exclusive.
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.databaseNameOverride", "properties.databaseNamePrefix"}},
			},
			ComputedFields: []string{
				"properties.attachedDatabaseNames",
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.defaultPrincipalsModificationKind", Value: "None"},
			},
			RequiredFields: []string{
				"properties.clusterResourceId",
				"properties.databaseName",
			},
		},
	}
}

func init() { azwise.Register(NewKustoClusterAttachedDatabaseConfiguration()) }
