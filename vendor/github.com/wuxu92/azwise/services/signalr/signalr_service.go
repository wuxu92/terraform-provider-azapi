// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package signalr

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SignalRService provides resource knowledge for Microsoft.SignalRService/signalR
// (Azure SignalR Service).
//
// Mirrors azurerm_signalr_service. The ARM ID casing is `signalR` (verified from
// the SDK id_signalr.go StaticSegment "signalR").
//
// Folded / separate resources intentionally NOT modelled here:
//   - azurerm_signalr_service_network_acl mutates the SAME signalR body at
//     properties.networkACLs (it is not a distinct ARM child resource — it just
//     PATCHes the parent). Its default_action / public_network / private_endpoint
//     ACL rules therefore belong to a properties.networkACLs sub-object that only
//     that resource manages; folding it in as universal knowledge would corrupt
//     plain azurerm_signalr_service bodies, so networkACLs is left out (and NOT
//     marked computed).
//   - customCertificates / customDomains / sharedPrivateLinkResources are distinct
//     ARM child resource types with their own knowledge files.
//   - primary/secondary access keys + connection strings are read-only outputs
//     retrieved from a separate Keys API, not part of the signalR body.
//
// Notes on fields that cannot be expressed declaratively:
//   - service_mode (Serverless/Classic/Default), connectivity_logs_enabled,
//     messaging_logs_enabled, http_request_logs_enabled and live_trace_enabled are
//     lowered by AzureRM into the properties.features[] array (FeatureFlags) and
//     properties.resourceLogConfiguration — array-element / composite paths that
//     azwise cannot target, so no StringRule/DefaultValue is emitted for them.
//   - upstream_endpoint.user_assigned_identity_id uses validation.IsUUID; it lives
//     inside properties.upstream.templates[] (an array element) so it routes to an
//     azapin customizer validator (validators.UUID), not a declarative rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/signalr/signalr_service_resource.go
//     :42-63 (timeouts), :65-179 (create -> signalr.SignalRResource),
//     :825-1078 (schema)
//   - go-azure-sdk resource-manager/signalr/2024-03-01/signalr:
//     model_signalrresource.go, model_signalrproperties.go, model_resourcesku.go,
//     model_serverlesssettings.go, model_signalrtlssettings.go,
//     id_signalr.go (Microsoft.SignalRService/signalR)
//   - helpers/constants.go:34-49 (sku name enum, not in REST spec)
type SignalRService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SignalRService)(nil)

// NewSignalRService returns knowledge for the Azure SignalR Service resource.
func NewSignalRService() *SignalRService {
	return &SignalRService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SignalRService/signalR",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// sku (name + capacity) is Required. location/name/resource_group are
			// envelope-owned and excluded.
			RequiredFields: []string{
				"sku.name",
				"sku.capacity",
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_access_enabled Default true -> "Enabled".
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// aad_auth_enabled Default true -> disableAadAuth = !true = false.
				{PropertyPath: "properties.disableAadAuth", Value: false},
				// local_auth_enabled Default true -> disableLocalAuth = !true = false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				// tls_client_cert_enabled Default false.
				{PropertyPath: "properties.tls.clientCertEnabled", Value: false},
				// serverless_connection_timeout_in_seconds Default 30.
				{PropertyPath: "properties.serverless.connectionTimeoutInSeconds", Value: float64(30)},
			},
			StringRules: []azwise.StringRule{
				// sku.name: ResourceSku.Name is a free ARM string; AzureRM restricts it
				// to this set (helpers.PossibleValuesForSkuName; enum absent from the
				// REST spec).
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Free_F1", "Standard_S1", "Premium_P1", "Premium_P2"},
					Message:       "sku name must be Free_F1, Standard_S1, Premium_P1 or Premium_P2",
				},
				// public_network_access_enabled -> Enabled/Disabled.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
			},
			// capacity uses validation.IntInSlice over a discrete set
			// {1..10,20,30,...,1000}; this is not a contiguous range so it cannot be
			// expressed as an IntRule (Min/Max) and is intentionally omitted.
			//
			// Read-only: present in SignalRProperties (GET) but not user-settable in the
			// create body (or retrieved via a separate API). networkACLs is settable
			// (managed by azurerm_signalr_service_network_acl) so it is NOT listed here.
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

func init() { azwise.Register(NewSignalRService()) }
