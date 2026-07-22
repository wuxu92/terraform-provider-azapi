package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterNetworkConnection provides resource knowledge for
// Microsoft.DevCenter/networkConnections.
//
// Contributing Terraform resource: azurerm_dev_center_network_connection.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_network_connection_resource.go
//     (schema L55-104, Create body L137-160)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/networkconnections:
//     id_networkconnection.go (type segment "networkConnections"),
//     model_networkproperties.go, constants.go (DomainJoinType).
//
// Key mappings:
//   - domain_join_type → properties.domainJoinType (Required, ForceNew, enum)
//   - subnet_id → properties.subnetId (Required)
//   - domain_name → properties.domainName
//   - domain_password → properties.domainPassword (Sensitive)
//   - domain_username → properties.domainUsername
//   - organization_unit → properties.organizationUnit
//
// Notes:
//   - name/location/resource_group_name/tags are envelope-owned.
//   - domain_name (validate.DevCenterNetworkConnectionDomainName) and domain_username
//     (validate.DevCenterNetworkConnectionDomainUsername) carry resource-specific semantic
//     validators (regex + format checks). These are not expressible as declarative
//     StringRules; port them to azapin customizer validators on properties.domainName /
//     properties.domainUsername when wiring native validation.
type DevCenterNetworkConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterNetworkConnection)(nil)

func NewDevCenterNetworkConnection() *DevCenterNetworkConnection {
	return &DevCenterNetworkConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/networkConnections",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.domainJoinType"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.domainJoinType",
					AllowedValues: []string{"AzureADJoin", "HybridAzureADJoin", "None"},
				},
			},
			SensitiveFields: []string{
				"properties.domainPassword",
			},
			RequiredFields: []string{
				"properties.domainJoinType",
				"properties.subnetId",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterNetworkConnection()) }
