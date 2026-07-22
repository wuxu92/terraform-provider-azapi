// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebPubSubHub provides resource knowledge for
// Microsoft.SignalRService/webPubSub/hubs.
//
// Mirrors azurerm_web_pubsub_hub. The ARM ID casing is `webPubSub/hubs` (verified
// from SDK id_hub.go).
//
// AzureRM mapping: anonymous_connections_enabled -> properties.anonymousConnectPolicy
// ("Allow"/"Deny", default "Deny"); event_handler -> properties.eventHandlers[];
// event_listener -> properties.eventListeners[]. The hub itself is mutable (has an
// Update path via CreateOrUpdate) so only the envelope name/web_pubsub_id are
// ForceNew.
//
// Not expressible declaratively (array-element / composite paths, routed to an
// azapin customizer if needed):
//   - event_handler.system_events (StringInSlice connect/connected/disconnected)
//     lives in properties.eventHandlers[].systemEvents[].
//   - event_listener.system_event_name_filter (StringInSlice connected/disconnected)
//     and its eventhub name validators live in properties.eventListeners[] elements.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/web_pubsub_hub_resource.go
//     :44-49 (timeouts), :51-155 (schema), :159-218 (create -> webpubsub.WebPubSubHub)
//   - internal/services/signalr/validate/webpubsub_name.go:20-25 (hub name regex)
//   - go-azure-sdk resource-manager/webpubsub/2024-03-01/webpubsub
//     model_webpubsubhubproperties.go, id_hub.go
type WebPubSubHub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebPubSubHub)(nil)

// NewWebPubSubHub returns knowledge for the Web PubSub hub resource.
func NewWebPubSubHub() *WebPubSubHub {
	return &WebPubSubHub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/webPubSub/hubs",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				// anonymous_connections_enabled Default false -> anonymousConnectPolicy "Deny".
				{PropertyPath: "properties.anonymousConnectPolicy", Value: "Deny"},
			},
			StringRules: []azwise.StringRule{
				// resource name: validate.WebPubSubHubName regex.
				{
					PropertyPath: "",
					Regex:        "^[A-Za-z][A-Za-z0-9_`,.\\[\\]]{0,127}$",
					Message:      "web pubsub hub name must be 1-128 chars, start with a letter, and contain only letters, numbers and the special characters `,` `_` `.` `[` `]`",
				},
				// anonymousConnectPolicy is a free ARM string; AzureRM sets Allow/Deny.
				{
					PropertyPath:  "properties.anonymousConnectPolicy",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "anonymousConnectPolicy must be Allow or Deny",
				},
			},
		},
	}
}

func init() { azwise.Register(NewWebPubSubHub()) }
