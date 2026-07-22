package redis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RedisLinkedServer provides resource knowledge for Microsoft.Cache/Redis/linkedServers.
//
// Mirrors azurerm_redis_linked_server. The resource has no Update, so every body
// argument is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/redis/redis_linked_server_resource.go:26-125
//   - Timeouts (:36-40): Create 60m / Read 5m / Delete 60m (no Update)
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     linked_redis_cache_id (linkedserver.ValidateRediID -> properties.linkedRedisCacheId,
//     resource-id validator).
//   - go-azure-sdk resource-manager/redis/2024-11-01/linkedserver
//     model_redislinkedservercreateparameters.go / model_redislinkedservercreateproperties.go /
//     model_redislinkedserverproperties.go / constants.go
type RedisLinkedServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RedisLinkedServer)(nil)

// NewRedisLinkedServer returns knowledge for the Redis linked server resource.
func NewRedisLinkedServer() *RedisLinkedServer {
	return &RedisLinkedServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/Redis/linkedServers",
			ApiVersions:  []string{"2024-11-01"},
			ForceNew: []azwise.ForceNewRule{
				// No Update func: linked_redis_cache_id, linked_redis_cache_location and
				// server_role are all ForceNew.
				{PropertyPath: "properties.linkedRedisCacheId"},
				{PropertyPath: "properties.linkedRedisCacheLocation"},
				{PropertyPath: "properties.serverRole"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 60 * time.Minute,
			},
			RequiredFields: []string{
				"properties.linkedRedisCacheId",
				"properties.linkedRedisCacheLocation",
				"properties.serverRole",
			},
			StringRules: []azwise.StringRule{
				// server_role: full ARM ReplicationRole enum.
				{
					PropertyPath:  "properties.serverRole",
					AllowedValues: []string{"Primary", "Secondary"},
					Message:       "serverRole must be Primary or Secondary",
				},
			},
			// provisioningState is present in RedisLinkedServerProperties (GET) but
			// absent from RedisLinkedServerCreateProperties (create model).
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewRedisLinkedServer()) }
