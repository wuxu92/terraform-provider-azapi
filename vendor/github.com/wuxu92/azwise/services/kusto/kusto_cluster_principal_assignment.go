package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoClusterPrincipalAssignment provides resource knowledge for
// Microsoft.Kusto/clusters/principalAssignments.
//
// Mirrors azurerm_kusto_cluster_principal_assignment. name / cluster_name /
// resource_group_name are envelope-owned; name is ForceNew (RequiresReplace). The
// whole resource is create/delete only (no Update) and every input is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_principal_assignment_resource.go:24-102
//     (schema: ForceNew name/cluster_name/tenant_id/principal_id/principal_type/role;
//     principal_type + role enums; timeouts; no Update)
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_principal_assignment_resource.go:129-136
//     (expand: tenantId, principalId, principalType, role)
//   - terraform-provider-azurerm internal/services/kusto/validate/cluster_principal_assignment_name.go:11-27
//     (ClusterPrincipalAssignmentName: charclass regex + max length 260)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/clusterprincipalassignments:
//     model_clusterprincipalassignment.go (envelope: name, properties),
//     model_clusterprincipalproperties.go:6-15 (properties.* body paths),
//     id_principalassignment.go:110-127 (ARM path casing: clusters/principalAssignments),
//     constants.go:14-18 (ClusterPrincipalRole: AllDatabasesAdmin/Monitor/Viewer),
//     constants.go:96-100 (PrincipalType: App/Group/User)
//
// Not encoded (deliberate):
//   - tenant_id / principal_id (validation.StringIsNotEmpty): non-empty checks add
//     no declarative constraint beyond presence.
//   - aadObjectId / principalName / provisioningState / tenantName are server-computed
//     read-only (see ComputedFields).
type KustoClusterPrincipalAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoClusterPrincipalAssignment)(nil)

// NewKustoClusterPrincipalAssignment returns knowledge for the
// clusters/principalAssignments resource.
func NewKustoClusterPrincipalAssignment() *KustoClusterPrincipalAssignment {
	return &KustoClusterPrincipalAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters/principalAssignments",
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
				// Resource name (PropertyPath == ""): ClusterPrincipalAssignmentName.
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
					AllowedValues: []string{"AllDatabasesAdmin", "AllDatabasesMonitor", "AllDatabasesViewer"},
					Message:       "must be AllDatabasesAdmin, AllDatabasesMonitor or AllDatabasesViewer",
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

func init() { azwise.Register(NewKustoClusterPrincipalAssignment()) }
