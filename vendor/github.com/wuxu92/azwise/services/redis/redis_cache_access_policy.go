package redis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RedisCacheAccessPolicy provides resource knowledge for
// Microsoft.Cache/Redis/accessPolicies.
//
// Mirrors azurerm_redis_cache_access_policy. AzureRM always creates a custom
// policy: properties.type is hardcoded to "Custom".
//
// Sources:
//   - terraform-provider-azurerm internal/services/redis/redis_cache_access_policy_resource.go:28-113
//   - Timeouts (Create/Read/Update/Delete :66,117,163,199): 5m each
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     redis_cache_id (rediscacheaccesspolicies.ValidateRediID -> parent cluster ref,
//     resource-id validator).
//   - go-azure-sdk resource-manager/redis/2024-11-01/rediscacheaccesspolicies
//     model_rediscacheaccesspolicy.go / model_rediscacheaccesspolicyproperties.go / constants.go
type RedisCacheAccessPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RedisCacheAccessPolicy)(nil)

// NewRedisCacheAccessPolicy returns knowledge for the Redis cache access policy resource.
func NewRedisCacheAccessPolicy() *RedisCacheAccessPolicy {
	return &RedisCacheAccessPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/Redis/accessPolicies",
			ApiVersions:  []string{"2024-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Update: 5 * time.Minute,
				Delete: 5 * time.Minute,
			},
			// permissions is Required. name / redis_cache_id are envelope/path.
			RequiredFields: []string{
				"properties.permissions",
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM hardcodes properties.type to "Custom" on create.
				{PropertyPath: "properties.type", Value: "Custom"},
			},
			StringRules: []azwise.StringRule{
				// type: full ARM AccessPolicyType enum (AzureRM only sends Custom, but the
				// ARM type accepts BuiltIn and AzAPI sends raw values).
				{
					PropertyPath:  "properties.type",
					AllowedValues: []string{"BuiltIn", "Custom"},
					Message:       "access policy type must be BuiltIn or Custom",
				},
			},
		},
	}
}

func init() { azwise.Register(NewRedisCacheAccessPolicy()) }
