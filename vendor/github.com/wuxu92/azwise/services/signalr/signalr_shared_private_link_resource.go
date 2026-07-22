// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SignalRSharedPrivateLinkResource provides resource knowledge for
// Microsoft.SignalRService/signalR/sharedPrivateLinkResources.
//
// Mirrors azurerm_signalr_shared_private_link_resource. The ARM ID casing is
// `signalR/sharedPrivateLinkResources` (verified from SDK
// id_sharedprivatelinkresource.go).
//
// AzureRM field mapping: sub_resource_name -> properties.groupId,
// target_resource_id -> properties.privateLinkResourceId (both ForceNew),
// request_message -> properties.requestMessage (mutable), status is read-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/signalr_shared_private_link_resource_resource.go
//     :24-83 (schema/timeouts), :86-137 (create -> signalr.SharedPrivateLinkResource)
//   - go-azure-sdk resource-manager/signalr/2024-03-01/signalr
//     model_sharedprivatelinkresourceproperties.go, id_sharedprivatelinkresource.go
type SignalRSharedPrivateLinkResource struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SignalRSharedPrivateLinkResource)(nil)

// NewSignalRSharedPrivateLinkResource returns knowledge for the SignalR shared
// private link resource.
func NewSignalRSharedPrivateLinkResource() *SignalRSharedPrivateLinkResource {
	return &SignalRSharedPrivateLinkResource{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/signalR/sharedPrivateLinkResources",
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

func init() { azwise.Register(NewSignalRSharedPrivateLinkResource()) }
