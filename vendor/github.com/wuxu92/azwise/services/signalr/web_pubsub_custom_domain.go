// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebPubSubCustomDomain provides resource knowledge for
// Microsoft.SignalRService/webPubSub/customDomains.
//
// Mirrors azurerm_web_pubsub_custom_domain. The ARM ID casing is
// `webPubSub/customDomains` (verified from SDK id_customdomain.go).
//
// The resource has no Update func, so every body property is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/web_pubsub_custom_domain_resource.go
//     :43-73 (schema), :87-141 (create -> webpubsub.CustomDomain),
//     timeouts :89 Create 30m / :145 Read 5m / :168 Delete 30m
//   - go-azure-sdk resource-manager/webpubsub/2024-03-01/webpubsub
//     model_customdomainproperties.go, model_resourcereference.go, id_customdomain.go
type WebPubSubCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebPubSubCustomDomain)(nil)

// NewWebPubSubCustomDomain returns knowledge for the Web PubSub custom domain.
func NewWebPubSubCustomDomain() *WebPubSubCustomDomain {
	return &WebPubSubCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/webPubSub/customDomains",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.domainName"},
				{PropertyPath: "properties.customCertificate.id"},
			},
			RequiredFields: []string{
				"properties.domainName",
				"properties.customCertificate.id",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewWebPubSubCustomDomain()) }
