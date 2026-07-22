// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SignalRCustomCertificate provides resource knowledge for
// Microsoft.SignalRService/signalR/customCertificates.
//
// Mirrors azurerm_signalr_service_custom_certificate. The ARM ID casing is
// `signalR/customCertificates` (verified from SDK id_customcertificate.go).
//
// AzureRM exposes a single custom_certificate_id (a Key Vault nested-item ID) that
// it splits into properties.keyVaultBaseUri + properties.keyVaultSecretName
// (+ optional keyVaultSecretVersion). The resource has no Update func, so every
// argument is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/signalr_service_custom_certificate_resource.go
//     :42-74 (schema/attributes), :84-148 (create -> signalr.CustomCertificate),
//     timeouts :86 Create 30m / :152 Read 5m / :179 Delete 30m
//   - go-azure-sdk resource-manager/signalr/2024-03-01/signalr
//     model_customcertificateproperties.go, id_customcertificate.go
type SignalRCustomCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SignalRCustomCertificate)(nil)

// NewSignalRCustomCertificate returns knowledge for the SignalR custom certificate.
func NewSignalRCustomCertificate() *SignalRCustomCertificate {
	return &SignalRCustomCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/signalR/customCertificates",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// No Update func -> the derived body properties are immutable.
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

func init() { azwise.Register(NewSignalRCustomCertificate()) }
