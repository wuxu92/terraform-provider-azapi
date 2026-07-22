package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoClusterDatabaseScript provides resource knowledge for
// Microsoft.Kusto/clusters/databases/scripts.
//
// Mirrors azurerm_kusto_database_script. name / database_id are envelope-owned
// (name and the parent database RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_script_resource.go:26-119
//     (schema: ForceNew name/database_id/sas_token/script_content/script_level;
//     Sensitive sas_token/script_content; ExactlyOneOf(url, script_content);
//     RequiredWith url<->sas_token; script_level + principal_permissions_action enums
//     and defaults; continue_on_errors_enabled default false; timeouts)
//   - terraform-provider-azurerm internal/services/kusto/kusto_script_resource.go:149-179
//     (expand: forceUpdateTag, scriptUrl, scriptUrlSasToken, scriptContent,
//     scriptLevel, principalPermissionsAction, continueOnErrors)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/scripts:
//     model_scriptproperties.go:6-15 (properties.* body paths),
//     id_script.go:115-135 (ARM path casing: clusters/databases/scripts),
//     constants.go:14-17 (PrincipalPermissionsAction: Remove/Retain...OnScriptCompletion),
//     constants.go:152-155 (ScriptLevel: Cluster/Database)
//
// Not encoded (deliberate):
//   - url / sas_token / script_content (validation.StringIsNotEmpty): non-empty
//     checks add no declarative constraint beyond presence.
//   - continue_on_errors_enabled maps to properties.continueOnErrors; the "_enabled"
//     TF suffix is stripped in the ARM body (bool default false, DefaultValues below).
//   - provisioningState is server-computed read-only (see ComputedFields).
//   - force_an_update_when_value_changed (properties.forceUpdateTag) is Optional+Computed:
//     AzureRM auto-generates a UUID when unset, so it is listed in DefaultValues with a
//     nil value rather than as a fixed default.
type KustoClusterDatabaseScript struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoClusterDatabaseScript)(nil)

// NewKustoClusterDatabaseScript returns knowledge for the
// clusters/databases/scripts resource.
func NewKustoClusterDatabaseScript() *KustoClusterDatabaseScript {
	return &KustoClusterDatabaseScript{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/databases/scripts",
			ApiVersions:  []string{"2025-02-14"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.scriptUrlSasToken"},
				{PropertyPath: "properties.scriptContent"},
				{PropertyPath: "properties.scriptLevel"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.scriptLevel",
					AllowedValues: []string{"Cluster", "Database"},
					Message:       "must be Cluster or Database",
				},
				{
					PropertyPath:  "properties.principalPermissionsAction",
					AllowedValues: []string{"RemovePermissionOnScriptCompletion", "RetainPermissionOnScriptCompletion"},
					Message:       "must be RemovePermissionOnScriptCompletion or RetainPermissionOnScriptCompletion",
				},
			},
			SensitiveFields: []string{
				"properties.scriptUrlSasToken",
				"properties.scriptContent",
			},
			// ExactlyOneOf(url, script_content) -> scriptUrl / scriptContent.
			ExactlyOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.scriptUrl", "properties.scriptContent"}},
			},
			// url RequiredWith sas_token, and sas_token RequiredWith url (bidirectional).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.scriptUrl", "properties.scriptUrlSasToken"}},
				{Paths: []string{"properties.scriptUrlSasToken", "properties.scriptUrl"}},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.scriptLevel", Value: "Database"},
				{PropertyPath: "properties.principalPermissionsAction", Value: "RetainPermissionOnScriptCompletion"},
				{PropertyPath: "properties.continueOnErrors", Value: false},
				{PropertyPath: "properties.forceUpdateTag"}, // Optional+Computed: UUID generated when unset
			},
		},
	}
}

func init() { azwise.Register(NewKustoClusterDatabaseScript()) }
