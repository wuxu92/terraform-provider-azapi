// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SignalRCustomDomain provides resource knowledge for
// Microsoft.SignalRService/signalR/customDomains.
//
// Mirrors azurerm_signalr_service_custom_domain. The ARM ID casing is
// `signalR/customDomains` (verified from SDK id_customdomain.go).
//
// The resource has no Update func, so every body property is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/signalr_service_custom_domain_resource.go
//     :41-71 (schema), :85-139 (create -> signalr.CustomDomain),
//     timeouts :87 Create 30m / :143 Read 5m / :166 Delete 30m
//   - go-azure-sdk resource-manager/signalr/2024-03-01/signalr
//     model_customdomainproperties.go, model_resourcereference.go, id_customdomain.go
type SignalRCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SignalRCustomDomain)(nil)

// NewSignalRCustomDomain returns knowledge for the SignalR custom domain.
func NewSignalRCustomDomain() *SignalRCustomDomain {
	return &SignalRCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/signalR/customDomains",
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

func init() { azwise.Register(NewSignalRCustomDomain()) }
