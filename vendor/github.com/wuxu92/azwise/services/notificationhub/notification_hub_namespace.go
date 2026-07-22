package notificationhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NotificationHubNamespace provides resource knowledge for
// Microsoft.NotificationHubs/namespaces.
//
// Contributing Terraform resource: azurerm_notification_hub_namespace.
//
// Sources:
//   - terraform-provider-azurerm internal/services/notificationhub/notification_hub_namespace_resource.go
//     (schema L29-114, Create body L142-158)
//   - go-azure-sdk resource-manager/notificationhubs/2023-09-01/namespaces:
//     model_namespaceresource.go, model_namespaceproperties.go, model_sku.go, constants.go
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not emitted as body rules.
//   - zone_redundancy_enabled (bool) expands to properties.zoneRedundancy enum
//     (Disabled/Enabled); the ForceNew + default are expressed on the ARM path.
//   - replication_region uses case-insensitive validation + location.Normalize; the
//     full SDK ReplicationRegion set is used since AzAPI sends raw ARM values.
type NotificationHubNamespace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NotificationHubNamespace)(nil)

// NewNotificationHubNamespace returns knowledge for the namespaces resource.
func NewNotificationHubNamespace() *NotificationHubNamespace {
	return &NotificationHubNamespace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NotificationHubs/namespaces",
			ApiVersions:  []string{"2023-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.enabled"},
				{PropertyPath: "properties.namespaceType"},
				{PropertyPath: "properties.zoneRedundancy"},
				{PropertyPath: "properties.replicationRegion"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Free", "Standard"},
					Message:       "sku_name must be Basic, Free or Standard",
				},
				{
					PropertyPath:  "properties.namespaceType",
					AllowedValues: []string{"Messaging", "NotificationHub"},
					Message:       "namespace_type must be Messaging or NotificationHub",
				},
				{
					PropertyPath: "properties.replicationRegion",
					AllowedValues: []string{
						"AustraliaEast", "BrazilSouth", "Default", "None",
						"NorthEurope", "SouthAfricaNorth", "SouthEastAsia", "WestUs2",
					},
					Message: "replication_region must be one of the supported ARM ReplicationRegion values",
				},
			},
			ComputedFields: []string{
				"properties.metricId",
				"properties.serviceBusEndpoint",
				"properties.provisioningState",
				"properties.status",
				"properties.createdAt",
				"properties.updatedAt",
				"properties.region",
				"properties.subscriptionId",
				"properties.scaleUnit",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enabled", Value: true},
				{PropertyPath: "properties.zoneRedundancy", Value: "Disabled"},
				{PropertyPath: "properties.replicationRegion", Value: "Default"},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.namespaceType",
			},
		},
	}
}

func init() { azwise.Register(NewNotificationHubNamespace()) }
