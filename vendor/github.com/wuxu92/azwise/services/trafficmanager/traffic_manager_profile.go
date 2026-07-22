package trafficmanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// TrafficManagerProfile provides resource knowledge for
// Microsoft.Network/trafficManagerProfiles.
//
// Mirrors azurerm_traffic_manager_profile. name and resource_group_name live on
// the envelope; the ARM body is created at location "global".
//
// Sources:
//   - terraform-provider-azurerm internal/services/trafficmanager/traffic_manager_profile_resource.go
//     (schema lines 53-201, Create body lines 226-247, expand funcs, CRUD timeouts 46-51)
//   - go-azure-sdk resource-manager/trafficmanager/2022-04-01/profiles:
//     id_trafficmanagerprofile.go (ARM type casing "trafficManagerProfiles" under
//     "Microsoft.Network"), model_profileproperties.go, model_dnsconfig.go,
//     model_monitorconfig.go, constants.go (TrafficRoutingMethod, MonitorProtocol,
//     ProfileStatus)
//
// Not encoded (deliberate):
//   - monitor_config.interval_in_seconds is IntInSlice([10,30]) — not a min/max
//     range, so it cannot be expressed as an IntRule; the default (30) is recorded.
//   - monitor_config.expected_status_code_ranges (StatusCodeRange validator) and
//     monitor_config.custom_header are array-element paths (properties.monitorConfig.*[*])
//     which azwise cannot resolve — skipped.
type TrafficManagerProfile struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*TrafficManagerProfile)(nil)

// NewTrafficManagerProfile returns knowledge for the trafficManagerProfiles resource.
func NewTrafficManagerProfile() *TrafficManagerProfile {
	return &TrafficManagerProfile{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/trafficManagerProfiles",
			ApiVersions:  []string{"2022-04-01"},
			// dns_config.relative_name is ForceNew (schema line 84); name is envelope.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.dnsConfig.relativeName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.trafficRoutingMethod",
					AllowedValues: []string{"Geographic", "MultiValue", "Performance", "Priority", "Subnet", "Weighted"},
					Message:       "traffic_routing_method must be one of Geographic, MultiValue, Performance, Priority, Subnet, Weighted",
				},
				{
					PropertyPath:  "properties.monitorConfig.protocol",
					AllowedValues: []string{"HTTP", "HTTPS", "TCP"},
					Message:       "monitor_config.protocol must be one of HTTP, HTTPS, TCP",
				},
				{
					PropertyPath:  "properties.profileStatus",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "profile_status must be Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.dnsConfig.ttl",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(2147483647)),
					Message:      "dns_config.ttl must be between 0 and 2147483647",
				},
				{
					PropertyPath: "properties.monitorConfig.port",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(65535)),
					Message:      "monitor_config.port must be between 1 and 65535",
				},
				{
					PropertyPath: "properties.monitorConfig.timeoutInSeconds",
					MinValue:     azwise.Ptr(int64(5)),
					MaxValue:     azwise.Ptr(int64(10)),
					Message:      "monitor_config.timeout_in_seconds must be between 5 and 10",
				},
				{
					PropertyPath: "properties.monitorConfig.toleratedNumberOfFailures",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(9)),
					Message:      "monitor_config.tolerated_number_of_failures must be between 0 and 9",
				},
				{
					PropertyPath: "properties.maxReturn",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(8)),
					Message:      "max_return must be between 1 and 8",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.profileStatus", Value: "Enabled"},
				{PropertyPath: "properties.monitorConfig.intervalInSeconds", Value: int64(30)},
				{PropertyPath: "properties.monitorConfig.timeoutInSeconds", Value: int64(10)},
				{PropertyPath: "properties.monitorConfig.toleratedNumberOfFailures", Value: int64(3)},
			},
			// traffic_routing_method, dns_config (relative_name + ttl) and monitor_config
			// (protocol + port) are Required.
			RequiredFields: []string{
				"properties.trafficRoutingMethod",
				"properties.dnsConfig.relativeName",
				"properties.dnsConfig.ttl",
				"properties.monitorConfig.protocol",
				"properties.monitorConfig.port",
			},
			// fqdn is populated by Azure (schema line 195-197).
			ComputedFields: []string{
				"properties.dnsConfig.fqdn",
			},
		},
	}
}

func init() { azwise.Register(NewTrafficManagerProfile()) }
