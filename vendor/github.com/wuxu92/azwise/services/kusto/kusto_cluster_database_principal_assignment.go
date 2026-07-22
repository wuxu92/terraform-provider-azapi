package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoClusterDatabasePrincipalAssignment provides resource knowledge for
// Microsoft.Kusto/clusters/databases/principalAssignments.
//
// Mirrors azurerm_kusto_database_principal_assignment. name / cluster_name /
// database_name / resource_group_name are envelope-owned; name is ForceNew
// (RequiresReplace). Create/delete only (no Update); every input is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_database_principal_assignment_resource.go:24-109
//     (schema: ForceNew name/cluster_name/database_name/tenant_id/principal_id/
//     principal_type/role; principal_type + role enums; timeouts; no Update)
//   - terraform-provider-azurerm internal/services/kusto/kusto_database_principal_assignment_resource.go:132-139
//     (expand: tenantId, principalId, principalType, role)
//   - terraform-provider-azurerm internal/services/kusto/validate/name.go:79-95
//     (DatabasePrincipalAssignmentName: charclass regex + max length 260)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/databaseprincipalassignments:
//     model_databaseprincipalproperties.go:6-15 (properties.* body paths),
//     id_databaseprincipalassignment.go:115-135 (ARM path casing:
//     clusters/databases/principalAssignments),
//     constants.go:52-59 (DatabasePrincipalRole: Admin/Ingestor/Monitor/
//     UnrestrictedViewer/User/Viewer),
//     constants.go:105-109 (PrincipalType: App/Group/User)
//
// Not encoded (deliberate):
//   - tenant_id / principal_id (validation.StringIsNotEmpty): non-empty checks add
//     no declarative constraint beyond presence.
//   - aadObjectId / principalName / provisioningState / tenantName are server-computed
//     read-only (see ComputedFields).
type KustoClusterDatabasePrincipalAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoClusterDatabasePrincipalAssignment)(nil)

// NewKustoClusterDatabasePrincipalAssignment returns knowledge for the
// clusters/databases/principalAssignments resource.
func NewKustoClusterDatabasePrincipalAssignment() *KustoClusterDatabasePrincipalAssignment {
	return &KustoClusterDatabasePrincipalAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/databases/principalAssignments",
			ApiVersions:  []string{"2025-02-14"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.tenantId"},
				{PropertyPath: "properties.principalId"},
				{PropertyPath: "properties.principalType"},
				{PropertyPath: "properties.role"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name (PropertyPath == ""): DatabasePrincipalAssignmentName.
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9\s.-]+$`,
					MaxLength:    260,
					Message:      "must only contain alphanumeric characters, whitespaces, dashes and dots, and be at most 260 characters",
				},
				{
					PropertyPath:  "properties.principalType",
					AllowedValues: []string{"App", "Group", "User"},
					Message:       "must be App, Group or User",
				},
				{
					PropertyPath:  "properties.role",
					AllowedValues: []string{"Admin", "Ingestor", "Monitor", "UnrestrictedViewer", "User", "Viewer"},
					Message:       "must be Admin, Ingestor, Monitor, UnrestrictedViewer, User or Viewer",
				},
			},
			ComputedFields: []string{
				"properties.aadObjectId",
				"properties.principalName",
				"properties.provisioningState",
				"properties.tenantName",
			},
			RequiredFields: []string{
				"properties.principalId",
				"properties.principalType",
				"properties.role",
			},
		},
	}
}

func init() { azwise.Register(NewKustoClusterDatabasePrincipalAssignment()) }
