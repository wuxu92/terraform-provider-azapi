// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebPubSubSharedPrivateLinkResource provides resource knowledge for
// Microsoft.SignalRService/webPubSub/sharedPrivateLinkResources.
//
// Mirrors azurerm_web_pubsub_shared_private_link_resource. The ARM ID casing is
// `webPubSub/sharedPrivateLinkResources` (verified from SDK
// id_sharedprivatelinkresource.go).
//
// AzureRM field mapping: subresource_name -> properties.groupId,
// target_resource_id -> properties.privateLinkResourceId (both ForceNew),
// request_message -> properties.requestMessage (mutable), status is read-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/web_pubsub_shared_private_link_resource_resource.go
//     :24-78 (schema/timeouts), :81-131 (create -> webpubsub.SharedPrivateLinkResource)
//   - go-azure-sdk resource-manager/webpubsub/2024-03-01/webpubsub
//     model_sharedprivatelinkresourceproperties.go, id_sharedprivatelinkresource.go
type WebPubSubSharedPrivateLinkResource struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebPubSubSharedPrivateLinkResource)(nil)

// NewWebPubSubSharedPrivateLinkResource returns knowledge for the Web PubSub shared
// private link resource.
func NewWebPubSubSharedPrivateLinkResource() *WebPubSubSharedPrivateLinkResource {
	return &WebPubSubSharedPrivateLinkResource{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/webPubSub/sharedPrivateLinkResources",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.groupId"},
				{PropertyPath: "properties.privateLinkResourceId"},
			},
			RequiredFields: []string{
				"properties.groupId",
				"properties.privateLinkResourceId",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.status",
			},
		},
	}
}

func init() { azwise.Register(NewWebPubSubSharedPrivateLinkResource()) }
