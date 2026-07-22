package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHub provides resource knowledge for Microsoft.Devices/IotHubs.
//
// Contributing Terraform resources (all mutate the single IotHub body):
//   - azurerm_iothub                      (the hub itself)
//   - azurerm_iothub_shared_access_policy (properties.authorizationPolicies[*])
//   - azurerm_iothub_endpoint_cosmosdb_account,
//     azurerm_iothub_endpoint_eventhub,
//     azurerm_iothub_endpoint_servicebus_queue,
//     azurerm_iothub_endpoint_servicebus_topic,
//     azurerm_iothub_endpoint_storage_container
//                                          (properties.routing.endpoints.*[*])
//   - azurerm_iothub_enrichment            (properties.routing.enrichments[*])
//   - azurerm_iothub_route                 (properties.routing.routes[*])
//   - azurerm_iothub_fallback_route        (properties.routing.fallbackRoute)
//   - azurerm_iothub_file_upload           (properties.storageEndpoints.$default,
//                                            properties.messagingEndpoints.fileNotifications)
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_resource.go
//     (schema L96-643, Create body L647-788, expand L1138-1439/L1848-1872)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_name.go
//     (IoTHubName L10-19), iothub_ip_rule_name.go (IoTHubIpRuleName L11-19),
//     iot_hub_endpoint_name.go (IoTHubEndpointName L8-24)
//   - go-azure-sdk (kermit) sdk/iothub/2022-04-30-preview/iothub:
//     models.go IotHubProperties L1057-1108, IotHubSkuInfo L1699-1706,
//     EventHubProperties L621, CloudToDeviceProperties L186-192,
//     FeedbackProperties L686-693, NetworkRuleSetProperties L1970-1977,
//     StorageEndpointProperties L2812-2823; enums.go (IotHubSku L218-232,
//     PublicNetworkAccess L346-350, DefaultAction L103-107)
//
// Notes:
//   - name/location/resource_group_name are envelope fields; not emitted as body rules.
//   - sku is a top-level envelope member of IotHubDescription (sku.name/sku.capacity),
//     NOT under properties.
//   - Folded-into-parent sub-resources are NOT distinct ARM types and are documented
//     rather than emitted as rules:
//       * routing config (routes, enrichments, fallbackRoute, endpoints) lives under
//         properties.routing.* — routes[*]/enrichments[*] and the five endpoint arrays
//         (endpoints.eventHubs[*], endpoints.serviceBusQueues[*],
//         endpoints.serviceBusTopics[*], endpoints.storageContainers[*], and the
//         2022-04-30-preview endpoints.cosmosDBSqlCollections[*]) are array-element
//         paths, skipped.
//       * shared_access_policy maps to properties.authorizationPolicies[*] (array element,
//         also AzureRM-computed) — skipped; it IS present in the Create model so it is NOT
//         listed as a ComputedField (StripComputedFields must not drop user values).
//       * file_upload maps to the fixed-key maps properties.storageEndpoints.$default.*
//         and properties.messagingEndpoints.fileNotifications.* — only the sensitive
//         connection string is emitted; ISO8601-duration fields are left to the server.
//   - event_hub_events_endpoint/namespace/path and event_hub_operations_endpoint/path are
//     read-only projections of properties.eventHubEndpoints.events.* (a settable sub-object),
//     so they are not listed as ComputedFields.
//   - public_network_access_enabled (bool) maps to the enum properties.publicNetworkAccess
//     (Enabled/Disabled); local_authentication_enabled (bool, default true) maps to the
//     inverted properties.disableLocalAuth (default false).
//   - endpoint.encoding is ForceNew in AzureRM but is an array-element path
//     (properties.routing.endpoints.storageContainers[*].encoding) — not emitted.
type IotHub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHub)(nil)

// NewIotHub returns knowledge for the Microsoft.Devices/IotHubs resource.
func NewIotHub() *IotHub {
	return &IotHub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Devices/IotHubs",
			ApiVersions:  []string{"2022-04-30-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.minTlsVersion"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-zA-Z-]{1,}$`,
					Message:      "IoT Hub name may only contain alphanumeric characters and dashes",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"B1", "B2", "B3", "F1", "S1", "S2", "S3"},
					Message:       "sku name must be one of B1, B2, B3, F1, S1, S2, S3",
				},
				{
					PropertyPath:  "properties.minTlsVersion",
					AllowedValues: []string{"1.2"},
					Message:       "min_tls_version must be 1.2",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.networkRuleSets.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "network_rule_set default_action must be Allow or Deny",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "sku.capacity",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(200)),
					Message:      "sku capacity must be between 1 and 200",
				},
				{
					PropertyPath: "properties.eventHubEndpoints.events.partitionCount",
					MinValue:     azwise.Ptr(int64(2)),
					MaxValue:     azwise.Ptr(int64(128)),
					Message:      "event_hub_partition_count must be between 2 and 128",
				},
				{
					PropertyPath: "properties.eventHubEndpoints.events.retentionTimeInDays",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(7)),
					Message:      "event_hub_retention_in_days must be between 1 and 7",
				},
				{
					PropertyPath: "properties.cloudToDevice.maxDeliveryCount",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
					Message:      "cloud_to_device max_delivery_count must be between 1 and 100",
				},
				{
					PropertyPath: "properties.cloudToDevice.feedback.maxDeliveryCount",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
					Message:      "cloud_to_device feedback max_delivery_count must be between 1 and 100",
				},
			},
			SensitiveFields: []string{
				"properties.storageEndpoints.$default.connectionString",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.state",
				"properties.hostName",
				"properties.locations",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				{PropertyPath: "properties.minTlsVersion", Value: "1.2"},
				{PropertyPath: "properties.eventHubEndpoints.events.partitionCount", Value: 4},
				{PropertyPath: "properties.eventHubEndpoints.events.retentionTimeInDays", Value: 1},
				{PropertyPath: "properties.networkRuleSets.defaultAction", Value: "Deny"},
			},
			RequiredFields: []string{
				"sku.name",
				"sku.capacity",
			},
		},
	}
}

func init() { azwise.Register(NewIotHub()) }
