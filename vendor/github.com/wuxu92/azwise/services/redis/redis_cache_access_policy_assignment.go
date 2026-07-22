package redis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RedisCacheAccessPolicyAssignment provides resource knowledge for
// Microsoft.Cache/Redis/accessPolicyAssignments.
//
// Mirrors azurerm_redis_cache_access_policy_assignment. This is the CLASSIC Redis
// access policy assignment and is a DISTINCT ARM type from
// Microsoft.Cache/redisEnterprise/databases/accessPolicyAssignments (registered in
// services/managedredis) — no collision. The resource has no Update, so every body
// argument is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/redis/redis_cache_access_policy_assignment_resource.go:31-192
//   - Timeouts (Create/Read/Delete :82,132,170): 5m each (no Update)
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     redis_cache_id (rediscacheaccesspolicyassignments.ValidateRediID -> parent cluster
//     ref, resource-id validator).
//   - go-azure-sdk resource-manager/redis/2024-11-01/rediscacheaccesspolicyassignments
//     model_rediscacheaccesspolicyassignment.go / model_rediscacheaccesspolicyassignmentproperties.go
type RedisCacheAccessPolicyAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RedisCacheAccessPolicyAssignment)(nil)

// NewRedisCacheAccessPolicyAssignment returns knowledge for the Redis cache access
// policy assignment resource.
func NewRedisCacheAccessPolicyAssignment() *RedisCacheAccessPolicyAssignment {
	return &RedisCacheAccessPolicyAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/Redis/accessPolicyAssignments",
			ApiVersions:  []string{"2024-11-01"},
			ForceNew: []azwise.ForceNewRule{
				// No Update func: every body argument is ForceNew.
				{PropertyPath: "properties.accessPolicyName"},
				{PropertyPath: "properties.objectId"},
				{PropertyPath: "properties.objectIdAlias"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 5 * time.Minute,
			},
			RequiredFields: []string{
				"properties.accessPolicyName",
				"properties.objectId",
				"properties.objectIdAlias",
			},
		},
	}
}

func init() { azwise.Register(NewRedisCacheAccessPolicyAssignment()) }
