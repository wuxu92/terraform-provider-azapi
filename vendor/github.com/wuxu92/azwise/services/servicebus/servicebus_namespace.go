// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusNamespace provides resource knowledge for Microsoft.ServiceBus/namespaces.
//
// Contributing Terraform resources:
//   - azurerm_servicebus_namespace
//   - azurerm_servicebus_namespace_customer_managed_key (FOLDED IN — mutates the
//     namespace's properties.encryption body rather than owning a distinct ARM type)
//
// Scoping verified via SDK id parser: NamespaceId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s (namespaces/id_namespace.go L104-106,
// ResourceProviderSegment "Microsoft.ServiceBus", StaticSegment "namespaces" L116-117).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_namespace_resource.go
//     (schema L70-249, CustomizeDiff L251-268, Create body L317-361, encryption expand L611-636)
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_namespace_customer_managed_key_resource.go
//     (Arguments L52-73 — key_vault_key_id / identity_id / infrastructure_encryption_enabled)
//   - terraform-provider-azurerm internal/services/servicebus/validate/namespace_name.go (L12-35)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/namespaces:
//     model_sbnamespaceproperties.go, model_sbsku.go, model_encryption.go,
//     constants.go (SkuName L316-320, TlsVersion L404-408, PublicNetworkAccess L231-235)
//
// Notes:
//   - name/location/resource_group_name/identity are envelope-owned; not emitted as body rules.
//   - sku maps to sku.name (top-level SBSku, not under properties); AzureRM mirrors it into sku.tier.
//   - capacity (sku.capacity, IntInSlice{0,1,2,4,8,16}) and premium_messaging_partitions
//     (properties.premiumMessagingPartitions, IntInSlice{0,1,2,4}) use discrete non-contiguous
//     allowed sets that IntRule (min/max only) cannot express; documented, not emitted.
//   - sku ForceNew is CONDITIONAL (only when migrating to/from Premium — Basic<->Standard is
//     an in-place update; servicebus_namespace_resource.go L258-264). Emitting sku.name as an
//     unconditional ForceNew would falsely force replacement on legal Basic<->Standard changes,
//     so it is intentionally NOT declared here.
//   - customer_managed_key removal is ForceNew (CustomizeDiff L253-256); block presence is not
//     a single body path, so it is documented rather than emitted.
//   - CMK key_vault_key_id expands into properties.encryption.keyVaultProperties[*].keyName /
//     .keyVersion / .keyVaultUri and identity_id into .identity.userAssignedIdentity — array
//     element paths not expressible as single-path rules; documented, not emitted. Only
//     infrastructure_encryption_enabled maps to a scalar (requireInfrastructureEncryption).
//   - default_primary/secondary_connection_string and default_primary/secondary_key are
//     data-plane (ListKeys) computed values, not settable body properties; omitted.
type ServiceBusNamespace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusNamespace)(nil)

// NewServiceBusNamespace returns knowledge for the namespaces resource.
func NewServiceBusNamespace() *ServiceBusNamespace {
	return &ServiceBusNamespace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.premiumMessagingPartitions"},
				// customer_managed_key.infrastructure_encryption_enabled (ForceNew: true).
				{PropertyPath: "properties.encryption.requireInfrastructureEncryption"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        "^[a-zA-Z][-a-zA-Z0-9]{4,48}[a-zA-Z0-9]$",
					Message:      "namespace name must be 6-50 characters, start with a letter, end with a letter or number, contain only letters, numbers and hyphens, and cannot end with -, -sb or -mgmt",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Standard", "Premium"},
					Message:       "sku must be one of Basic, Standard, Premium",
				},
				{
					PropertyPath:  "properties.minimumTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2"},
					Message:       "minimum_tls_version must be one of 1.0, 1.1, 1.2",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled", "SecuredByPerimeter"},
					Message:       "publicNetworkAccess must be one of Enabled, Disabled, SecuredByPerimeter",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// local_auth_enabled default true => disableLocalAuth false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				// public_network_access_enabled default true => Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.minimumTlsVersion", Value: "1.2"},
			},
			RequiredFields: []string{
				"sku.name",
			},
			ComputedFields: []string{
				"properties.metricId",
				"properties.serviceBusEndpoint",
				"properties.provisioningState",
				"properties.status",
				"properties.createdAt",
				"properties.updatedAt",
				"properties.privateEndpointConnections",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusNamespace()) }
