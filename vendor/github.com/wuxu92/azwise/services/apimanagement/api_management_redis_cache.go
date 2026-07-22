package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementRedisCache provides resource knowledge for
// Microsoft.ApiManagement/service/caches.
//
// Mirrors azurerm_api_management_redis_cache. connection_string is sensitive.
// redis_cache_id (the target Azure Cache for Redis resource ID) is prefixed with
// the Resource Manager endpoint by AzureRM before being sent as
// properties.resourceId; AzAPI users supply the full ARM value directly.
// cache_location maps to properties.useFromLocation and defaults to "default".
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_redis_cache_resource.go
//     schema (lines 45-86), CreateUpdate body (lines 122-135), timeouts (30m/5m/30m/30m)
//   - go-azure-sdk apimanagement/2022-08-01/cache CacheContractProperties
type ApiManagementRedisCache struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementRedisCache)(nil)

func NewApiManagementRedisCache() *ApiManagementRedisCache {
	return &ApiManagementRedisCache{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/caches",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // name (ApiManagementChildName, ForceNew)
			},
			RequiredFields: []string{
				"properties.connectionString",
				"properties.useFromLocation",
			},
			// cache_location defaults to "default" (location.Normalize).
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.useFromLocation", Value: "default"},
			},
			SensitiveFields: []string{
				"properties.connectionString", // Sensitive: true in AzureRM schema
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── ApiManagementChildName
				{
					Regex:   `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					Message: "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length",
				},
				// ── connection_string → properties.connectionString ── StringIsNotEmpty
				{PropertyPath: "properties.connectionString", MinLength: 1, Message: "connection_string must not be empty"},
				// ── description → properties.description ── StringIsNotEmpty
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementRedisCache()) }
