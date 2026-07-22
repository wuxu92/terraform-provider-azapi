// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebPubSub provides resource knowledge for Microsoft.SignalRService/webPubSub
// (Azure Web PubSub).
//
// The ARM ID casing is `webPubSub` (verified from the SDK id_webpubsub.go
// StaticSegment "webPubSub"). Both the signalr and webpubsub SDKs live under the
// Microsoft.SignalRService provider.
//
// MERGED TF resources (both map to the SAME ARM type Microsoft.SignalRService/
// webPubSub — the SDK uses one WebPubSubId + WebPubSubResource for both, keyed by
// the optional top-level `kind` discriminator):
//   - azurerm_web_pubsub          (kind omitted / "WebPubSub")
//   - azurerm_web_pubsub_socketio (kind "SocketIO", adds properties.socketIO)
//
// Only knowledge universal to BOTH bodies is emitted here. Kind-specific items are
// deliberately excluded or scoped so they fire only when their sub-object is
// present:
//   - The top-level `kind` field ("SocketIO") is socketio-only, so it is NOT added
//     as a required field or default.
//   - properties.socketIO.serviceMode is a SocketIO-only sub-object; it is emitted
//     as a StringRule (enum), which validates only when the field is present, but
//     its default ("Default") is NOT injected as a universal DefaultValue (that
//     would corrupt a plain webPubSub body).
//
// Folded / separate resources NOT modelled here:
//   - azurerm_web_pubsub_network_acl mutates the SAME webPubSub body at
//     properties.networkACLs (not a distinct ARM child resource); its ACL rules
//     belong to a sub-object only that resource manages, so networkACLs is left out
//     (and NOT marked computed).
//   - customCertificates / customDomains / hubs / sharedPrivateLinkResources are
//     distinct ARM child resource types with their own knowledge files.
//   - access keys + connection strings are read-only outputs from a separate Keys
//     API.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/web_pubsub_resource.go
//     :42-47 (timeouts), :60-202 (schema), :206-277 (create -> webpubsub.WebPubSubResource)
//   - internal/services/signalr/web_pubsub_socketio_resource.go :73-170 (arguments),
//     :220-264 (CustomizeDiff sku/capacity), :274-338 (create, kind=SocketIO,
//     properties.socketIO.serviceMode)
//   - internal/services/signalr/validate/webpubsub_name.go:13-18 (name regex)
//   - go-azure-sdk resource-manager/webpubsub/2024-03-01/webpubsub:
//     model_webpubsubresource.go, model_webpubsubproperties.go, model_resourcesku.go,
//     model_webpubsubtlssettings.go, model_webpubsubsocketiosettings.go,
//     constants.go (ServiceKind), id_webpubsub.go (Microsoft.SignalRService/webPubSub)
//   - helpers/constants.go (sku name / public network access / socketIO service mode enums)
type WebPubSub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebPubSub)(nil)

// NewWebPubSub returns knowledge for the Azure Web PubSub resource (and its
// SocketIO variant, which share the webPubSub ARM type).
func NewWebPubSub() *WebPubSub {
	return &WebPubSub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/webPubSub",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// sku is Required in both resources (name mandatory). location/name/
			// resource_group are envelope-owned.
			RequiredFields: []string{
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_access(_enabled) Default "Enabled" in both.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// aad_auth_enabled Default true -> disableAadAuth = false.
				{PropertyPath: "properties.disableAadAuth", Value: false},
				// local_auth_enabled Default true -> disableLocalAuth = false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				// tls_client_cert_enabled Default false.
				{PropertyPath: "properties.tls.clientCertEnabled", Value: false},
				// capacity Default 1 in both.
				{PropertyPath: "sku.capacity", Value: float64(1)},
			},
			StringRules: []azwise.StringRule{
				// resource name: validate.WebPubSubName regex (shared by both).
				{
					PropertyPath: "",
					Regex:        "^[a-zA-Z][-a-zA-Z0-9]{1,61}[a-zA-Z0-9]$",
					Message:      "web pubsub name must be 3-63 chars, start with a letter, end with a letter or number, and contain only letters, numbers and hyphens",
				},
				// sku.name: ResourceSku.Name is a free ARM string; AzureRM restricts it
				// to this set in both resources (helpers.PossibleValuesForSkuName).
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Free_F1", "Standard_S1", "Premium_P1", "Premium_P2"},
					Message:       "sku name must be Free_F1, Standard_S1, Premium_P1 or Premium_P2",
				},
				// public_network_access -> Enabled/Disabled.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
				// SocketIO-only sub-object: fires only when properties.socketIO is set,
				// so it is safe to union (helpers.PossibleValuesForSocketIOServiceMode).
				{
					PropertyPath:  "properties.socketIO.serviceMode",
					AllowedValues: []string{"Default", "Serverless"},
					Message:       "socketIO serviceMode must be Default or Serverless",
				},
			},
			// capacity uses validation.IntInSlice over a discrete non-contiguous set
			// {1..10,20,...,1000}; not expressible as an IntRule (Min/Max), so omitted.
			//
			// Read-only: present in WebPubSubProperties (GET) but not user-settable in
			// the create body (or fetched via a separate API). networkACLs is settable
			// (azurerm_web_pubsub_network_acl) so it is NOT listed.
			ComputedFields: []string{
				"properties.externalIP",
				"properties.hostName",
				"properties.publicPort",
				"properties.serverPort",
				"properties.version",
				"properties.provisioningState",
				"properties.privateEndpointConnections",
			},
		},
	}
}

func init() { azwise.Register(NewWebPubSub()) }
