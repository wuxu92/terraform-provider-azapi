// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package purview

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Account provides resource knowledge for Microsoft.Purview/accounts.
//
// Mirrors azurerm_purview_account. public_network_enabled (bool) maps onto the
// properties.publicNetworkAccess Enabled/Disabled enum; managed_event_hub_enabled
// (bool) maps onto properties.managedEventHubState Enabled/Disabled.
//
// Sources:
//   - terraform-provider-azurerm internal/services/purview/purview_account_resource.go
//     Schema() lines 49-140, Create() lines 144-199, Read() lines 201-279
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/purview/2021-12-01/account
//     model_accountproperties.go (managedEventHubState, managedResourceGroupName,
//     managedResources, endpoints, publicNetworkAccess, provisioningState, ...) /
//     constants.go (PublicNetworkAccess & ManagedEventHubState: Enabled/Disabled/NotSpecified)
//   - id_account.go: /providers/Microsoft.Purview/accounts/{accountName}
//
// atlas_kafka_endpoint_* connection strings are surfaced via ListKeys (not the
// account body) and are read-only, so they are neither SensitiveFields nor
// ComputedFields body entries.
type Account struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Account)(nil)

// NewAccount returns knowledge for the Purview accounts resource.
func NewAccount() *Account {
	return &Account{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Purview/accounts",
			ApiVersions:  []string{"2021-12-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name/resource_group are envelope.
				{PropertyPath: "location"},
				// managed_resource_group_name is ForceNew.
				{PropertyPath: "properties.managedResourceGroupName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_enabled Default true -> publicNetworkAccess Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// managed_event_hub_enabled Default true -> managedEventHubState Enabled.
				{PropertyPath: "properties.managedEventHubState", Value: "Enabled"},
			},
			StringRules: []azwise.StringRule{
				// name: validation.StringMatch (3-63 chars, alphanumeric + hyphen,
				// first/last must be alphanumeric).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][-a-zA-Z0-9]{1,61}[a-zA-Z0-9]$`,
					Message:      "name must be 3-63 characters of letters, numbers and hyphens, starting and ending with a letter or number",
				},
				// public_network_enabled enum. Full SDK PublicNetworkAccess set.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled", "NotSpecified"},
					Message:       "publicNetworkAccess must be one of Enabled, Disabled, or NotSpecified",
				},
				// managed_event_hub_enabled enum. Full SDK ManagedEventHubState set.
				{
					PropertyPath:  "properties.managedEventHubState",
					AllowedValues: []string{"Enabled", "Disabled", "NotSpecified"},
					Message:       "managedEventHubState must be one of Enabled, Disabled, or NotSpecified",
				},
			},
			// Server-computed, read-only body properties surfaced by Read() but never
			// written by Create/Update (endpoints, managed resources, provisioning &
			// account status, audit metadata, cloud connectors, private endpoints).
			ComputedFields: []string{
				"properties.managedResources",
				"properties.endpoints",
				"properties.friendlyName",
				"properties.provisioningState",
				"properties.accountStatus",
				"properties.createdAt",
				"properties.createdBy",
				"properties.createdByObjectId",
				"properties.privateEndpointConnections",
				"properties.cloudConnectors",
			},
			// managed_resource_group_name uses resourcegroups.ValidateName — a semantic
			// name check better expressed as a customizer validator than a regex here.
		},
	}
}

func init() { azwise.Register(NewAccount()) }
