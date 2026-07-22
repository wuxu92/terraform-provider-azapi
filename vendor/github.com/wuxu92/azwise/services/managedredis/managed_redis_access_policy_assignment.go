package managedredis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedRedisAccessPolicyAssignment provides resource knowledge for
// Microsoft.Cache/redisEnterprise/databases/accessPolicyAssignments.
//
// Mirrors azurerm_managed_redis_access_policy_assignment. The assignment is always
// created on the "default" database of the parent cluster, using the caller's
// object_id both as the assignment name and as properties.user.objectId.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedredis/managed_redis_access_policy_assignment_resource.go
//   - Arguments(): managed_redis_id (ForceNew, parent cluster ref), object_id
//     (ForceNew, validation.IsUUID -> properties.user.objectId)
//   - Create(): properties.accessPolicyName hardcoded to "default"
//   - Create/Read/Delete timeouts: 30m / 5m / 30m (no Update — every field is ForceNew)
//   - go-azure-sdk resource-manager/redisenterprise/2025-07-01/databases
//     model_accesspolicyassignmentproperties.go:
//     AccessPolicyAssignmentProperties.{AccessPolicyName string, User {ObjectId *string}}
//
// object_id is validated with validation.IsUUID; that generic semantic validator is
// attached in the azapin customizer (validators.UUID on properties.user.objectId),
// not as a declarative StringRule, so it is not repeated here.
type ManagedRedisAccessPolicyAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedRedisAccessPolicyAssignment)(nil)

// NewManagedRedisAccessPolicyAssignment returns knowledge for the
// accessPolicyAssignments resource.
func NewManagedRedisAccessPolicyAssignment() *ManagedRedisAccessPolicyAssignment {
	return &ManagedRedisAccessPolicyAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/redisEnterprise/databases/accessPolicyAssignments",
			ApiVersions:  []string{"2025-07-01"},
			// object_id is the only body-carried argument and is ForceNew; the parent
			// cluster ref and the "default" database are envelope/path.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.user.objectId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Both properties are required for a valid assignment; accessPolicyName is
			// hardcoded to "default" by AzureRM.
			RequiredFields: []string{
				"properties.accessPolicyName",
				"properties.user.objectId",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.accessPolicyName", Value: "default"},
			},
		},
	}
}

func init() { azwise.Register(NewManagedRedisAccessPolicyAssignment()) }
