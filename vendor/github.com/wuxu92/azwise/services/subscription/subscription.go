package subscription

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SubscriptionAlias provides resource knowledge for Microsoft.Subscription/aliases.
//
// Mirrors azurerm_subscription, which creates a subscription Alias
// (Microsoft.Subscription/aliases) via a PutAliasRequest body — a real ARM resource,
// not a pure operation. The alias is a tenant-level resource (its ID has no
// subscription/resourceGroup segments).
//
// Sources:
//   - internal/services/subscription/subscription_resource.go
//     (schema 54-121: subscription_name Required SubscriptionName (1-64, no <>;|);
//     alias Optional+Computed ForceNew StringIsNotEmpty (resource name);
//     billing_scope_id Optional ExactlyOneOf(subscription_id) billing-scope validators;
//     workload Optional ForceNew StringInSlice default Production; subscription_id
//     Optional+Computed ForceNew IsUUID ExactlyOneOf(billing_scope_id);
//     tenant_id Computed; timeouts Create/Update/Delete 30m Read 5m;
//     create 163-217 → PutAliasRequest{Properties{Workload, SubscriptionId,
//     DisplayName, BillingScope}}).
//   - internal/services/subscription/validate/subscription_name.go (len 1-64, excludes <>;|).
//   - go-azure-sdk resource-manager/subscription/2021-10-01/subscriptions:
//     model_putaliasrequestproperties.go (billingScope/displayName/subscriptionId/
//     workload/resellerId/additionalProperties settable),
//     model_subscriptionaliasresponseproperties.go (acceptOwnershipState/
//     acceptOwnershipUrl/createdTime/managementGroupId/provisioningState/
//     subscriptionOwnerId read-only), constants.go (Workload DevTest/Production),
//     id_alias.go (segment casing "Microsoft.Subscription"/"aliases").
type SubscriptionAlias struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SubscriptionAlias)(nil)

// NewSubscriptionAlias returns knowledge for the aliases resource.
func NewSubscriptionAlias() *SubscriptionAlias {
	return &SubscriptionAlias{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Subscription/aliases",
			ApiVersions:  []string{"2021-10-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.workload"},
				{PropertyPath: "properties.subscriptionId"},
			},
			StringRules: []azwise.StringRule{
				{
					// alias: StringIsNotEmpty (resource name).
					MinLength: 1,
					Message:   "must not be empty",
				},
				{
					// subscription_name → properties.displayName: 1-64 chars, no <>;| .
					PropertyPath: "properties.displayName",
					Regex:        `^[^<>;|]{1,64}$`,
					MinLength:    1,
					MaxLength:    64,
					Message:      "must be 1-64 characters and cannot contain the characters <, >, ; or |",
				},
				{
					// workload: StringInSlice; full SDK enum set.
					PropertyPath:  "properties.workload",
					AllowedValues: []string{"DevTest", "Production"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// workload Default "Production".
				{PropertyPath: "properties.workload", Value: "Production"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.acceptOwnershipState",
				"properties.acceptOwnershipUrl",
				"properties.createdTime",
				"properties.managementGroupId",
				"properties.subscriptionOwnerId",
			},
			// NOTE: subscription_name → properties.displayName is Required in AzureRM but
			// is only sent to the ARM body when creating a brand-new subscription (not when
			// adopting an existing one via subscription_id), so it is not listed as a
			// universal RequiredField.
			// NOTE: billing_scope_id ExactlyOneOf subscription_id (config-level constraint
			// mapping to properties.billingScope / properties.subscriptionId). billing_scope_id
			// uses billing-scope-ID validators (MicrosoftCustomerAccount/Enrollment/
			// MicrosoftPartnerAccount) and subscription_id uses validation.IsUUID — both are
			// semantic checks; attach a UUID validator to properties.subscriptionId and a
			// billing-scope validator to properties.billingScope in the resource customizer.
			// NOTE: tenant_id (Computed) and tags are not part of the alias PUT body — tags
			// are applied via a separate Tags API call at the subscription scope — so neither
			// is represented as an alias body property here.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSubscriptionAlias()) }
