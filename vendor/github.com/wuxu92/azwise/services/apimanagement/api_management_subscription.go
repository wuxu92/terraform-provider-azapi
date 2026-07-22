package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementSubscription provides resource knowledge for
// Microsoft.ApiManagement/service/subscriptions.
//
// Mirrors azurerm_api_management_subscription. The AzureRM `product_id`/`api_id`
// convenience fields (mutually exclusive) both expand into ARM
// properties.scope; `user_id` maps to properties.ownerId. All three are
// ForceNew. primary_key/secondary_key are sensitive.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_subscription_resource.go
//     schema (lines 50-126); Create body (lines 168-196); Timeouts 30m/5m/30m/30m.
//   - Microsoft.ApiManagement/service/subscriptions@2022-08-01
//     subscription.SubscriptionCreateParameterProperties: displayName (required),
//     scope (required), state (enum), allowTracing, ownerId, primaryKey, secondaryKey.
type ApiManagementSubscription struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementSubscription)(nil)

// NewApiManagementSubscription returns knowledge for the API Management subscription resource.
func NewApiManagementSubscription() *ApiManagementSubscription {
	return &ApiManagementSubscription{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/subscriptions",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				// user_id -> ownerId (ForceNew).
				{PropertyPath: "properties.ownerId"},
				// product_id / api_id both expand to scope (both ForceNew).
				{PropertyPath: "properties.scope"},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.scope",
			},
			SensitiveFields: []string{
				"properties.primaryKey",
				"properties.secondaryKey",
			},
			DefaultValues: []azwise.DefaultValue{
				// state default "submitted" (schema line 96).
				{PropertyPath: "properties.state", Value: "submitted"},
				// allow_tracing default true (schema line 124).
				{PropertyPath: "properties.allowTracing", Value: true},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"active", "cancelled", "expired", "rejected", "submitted", "suspended"},
					Message:       "state must be one of active, cancelled, expired, rejected, submitted, suspended",
				},
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementSubscription()) }
