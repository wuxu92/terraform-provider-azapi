package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistryWebhook provides resource knowledge for
// Microsoft.ContainerRegistry/registries/webhooks.
//
// Contributing Terraform resource: azurerm_container_registry_webhook.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_webhook_resource.go
//     (schema L51-117, expand L270-311)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_webhook_name.go
//   - go-azure-sdk resource-manager/containerregistry/2025-11-01/webhooks:
//     model_webhookpropertiescreateparameters.go, model_webhookproperties.go,
//     constants.go (WebhookAction, WebhookStatus).
//
// Notes:
//   - service_uri (Required) maps to properties.serviceUri.
//   - custom_headers maps to properties.customHeaders (map).
//   - actions (Required, MinItems 1) maps to properties.actions[*] (enum).
type ContainerRegistryWebhook struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistryWebhook)(nil)

// NewContainerRegistryWebhook returns knowledge for the webhooks resource.
func NewContainerRegistryWebhook() *ContainerRegistryWebhook {
	return &ContainerRegistryWebhook{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/webhooks",
			ApiVersions:  []string{"2025-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryWebhookName
				{
					Regex:     `^[a-zA-Z0-9]{5,50}$`,
					MinLength: 5,
					MaxLength: 50,
					Message:   "alpha numeric characters only, 5-50 chars",
				},
				// ── status → properties.status
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"enabled", "disabled"},
					Message:       "must be one of enabled or disabled",
				},
				// ── actions[*] → properties.actions
				{
					PropertyPath:  "properties.actions[*]",
					AllowedValues: []string{"chart_delete", "chart_push", "delete", "push", "quarantine"},
					Message:       "each action must be one of chart_delete, chart_push, delete, push or quarantine",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// status Default enabled.
				{PropertyPath: "properties.status", Value: "enabled"},
			},
			// service_uri and actions are Required for creation.
			RequiredFields: []string{"properties.serviceUri", "properties.actions"},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistryWebhook()) }
