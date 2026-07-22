// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebPubSubCustomCertificate provides resource knowledge for
// Microsoft.SignalRService/webPubSub/customCertificates.
//
// Mirrors azurerm_web_pubsub_custom_certificate. The ARM ID casing is
// `webPubSub/customCertificates` (verified from SDK id_customcertificate.go).
//
// AzureRM splits custom_certificate_id (a Key Vault nested-item ID) into
// properties.keyVaultBaseUri + properties.keyVaultSecretName (+ optional version).
// The resource has no Update func, so every body property is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/web_pubsub_custom_certificate_resource.go
//     :45-77 (schema/attributes), :87-147 (create -> webpubsub.CustomCertificate),
//     timeouts :89 Create 30m / :151 Read 5m / :192 Delete 30m
//   - go-azure-sdk resource-manager/webpubsub/2024-03-01/webpubsub
//     model_customcertificateproperties.go, id_customcertificate.go
type WebPubSubCustomCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebPubSubCustomCertificate)(nil)

// NewWebPubSubCustomCertificate returns knowledge for the Web PubSub custom certificate.
func NewWebPubSubCustomCertificate() *WebPubSubCustomCertificate {
	return &WebPubSubCustomCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/webPubSub/customCertificates",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.keyVaultBaseUri"},
				{PropertyPath: "properties.keyVaultSecretName"},
				{PropertyPath: "properties.keyVaultSecretVersion"},
			},
			RequiredFields: []string{
				"properties.keyVaultBaseUri",
				"properties.keyVaultSecretName",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewWebPubSubCustomCertificate()) }
