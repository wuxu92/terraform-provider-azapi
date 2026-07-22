// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusNamespaceDisasterRecoveryConfig provides resource knowledge for
// Microsoft.ServiceBus/namespaces/disasterRecoveryConfigs.
//
// Contributing Terraform resource: azurerm_servicebus_namespace_disaster_recovery_config.
//
// Scoping verified via SDK id parser: DisasterRecoveryConfigId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/disasterRecoveryConfigs/%s
// (disasterrecoveryconfigs/id_disasterrecoveryconfig.go L110-111).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_namespace_disaster_recovery_config_resource.go
//     (schema L46-94, Create body L125-129)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/disasterrecoveryconfigs:
//     model_armdisasterrecoveryproperties.go (PartnerNamespace, AlternateName)
//
// Notes:
//   - name/primary_namespace_id are envelope / parent-reference fields; not emitted as body rules.
//   - partner_namespace_id maps to properties.partnerNamespace and is Required in AzureRM.
//   - alias_authorization_rule_id is used only to look up connection-string keys during Read
//     (ListKeys, L213-222); it is not written into the DR config body — omitted.
//   - primary/secondary_connection_string_alias and default_primary/secondary_key are
//     data-plane (ListKeys) computed values, not settable body properties; omitted.
//   - name has no ValidateFunc in AzureRM (schema L47-51); no name StringRule emitted.
type ServiceBusNamespaceDisasterRecoveryConfig struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusNamespaceDisasterRecoveryConfig)(nil)

// NewServiceBusNamespaceDisasterRecoveryConfig returns knowledge for the
// namespaces/disasterRecoveryConfigs resource.
func NewServiceBusNamespaceDisasterRecoveryConfig() *ServiceBusNamespaceDisasterRecoveryConfig {
	return &ServiceBusNamespaceDisasterRecoveryConfig{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/disasterRecoveryConfigs",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.partnerNamespace",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.role",
				"properties.pendingReplicationOperationsCount",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusNamespaceDisasterRecoveryConfig()) }
