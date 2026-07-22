package nginx

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NginxDeployment provides resource knowledge for Nginx.NginxPlus/nginxDeployments.
//
// Mirrors azurerm_nginx_deployment.
//
// ARM type casing verified against go-azure-sdk nginxdeployment/id_nginxdeployment.go Segments:
// "Nginx.NginxPlus" / "nginxDeployments".
//
// Sources:
//   - terraform-provider-azurerm internal/services/nginx/nginx_deployment_resource.go
//     Arguments (134-331: name StringIsNotEmpty ForceNew; sku StringIsNotEmpty Required; identity
//     SystemAssignedUserAssignedOptional; location; capacity IntPositive ConflictsWith
//     auto_scale_profile; auto_scale_profile ConflictsWith capacity with name/min_capacity/
//     max_capacity IntPositive; email StringIsNotEmpty; frontend_public MaxItems 1 ConflictsWith
//     frontend_private; frontend_private ConflictsWith frontend_public with ip_address/
//     allocation_method(Dynamic/Static)/subnet_id; network_interface subnet_id; automatic_upgrade_channel
//     enum stable/preview Default stable; web_application_firewall MaxItems 1 with
//     activation_state_enabled + computed status), Attributes (333-349: nginx_version/ip_address/
//     dataplane_api_endpoint Computed), expandCreateForNginxDeployment (398-493), timeouts Create 30m /
//     Read 5m / Update 30m / Delete 30m.
//   - go-azure-sdk resource-manager/nginx/2024-11-01-preview/nginxdeployment
//     NginxDeployment (sku.name, identity, location, tags), NginxDeploymentProperties
//     (autoUpgradeProfile/networkProfile/nginxAppProtect/scalingProperties/userProfile settable;
//     dataplaneApiEndpoint/ipAddress/nginxVersion/provisioningState read-only), ScaleProfileCapacity
//     (min/max int64), NginxPrivateIPAddress (privateIPAddress/privateIPAllocationMethod/subnetId),
//     constants.go ActivationState [Disabled, Enabled] / NginxPrivateIPAllocationMethod [Dynamic, Static].
//
// Note: several constraints live at array-element paths and are documented but not emitted as rules:
// auto_scale_profile name/min_capacity/max_capacity (properties.scalingProperties.autoScaleSettings.
// profiles[*]), frontend_private allocation_method enum (properties.networkProfile.
// frontEndIPConfiguration.privateIPAddresses[*].privateIPAllocationMethod). The web_application_firewall
// status subtree is fully computed (server-populated). The scaling requirement (non-basic SKU requires
// capacity or auto_scale_profile; basic SKU forbids them) is a CustomizeDiff value-based rule that
// cannot be expressed declaratively. managed_resource_group has no ARM equivalent (removed by the API).
type NginxDeployment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NginxDeployment)(nil)

// NewNginxDeployment returns knowledge for the Nginx nginxDeployments resource.
func NewNginxDeployment() *NginxDeployment {
	return &NginxDeployment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Nginx.NginxPlus/nginxDeployments",
			ApiVersions:  []string{"2024-11-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			RequiredFields: []string{
				"sku.name",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "name", MinLength: 1, Message: "name must not be empty"},
				{PropertyPath: "sku.name", MinLength: 1, Message: "sku must not be empty"},
				{PropertyPath: "properties.userProfile.preferredEmail", MinLength: 1, Message: "email must not be empty"},
				{
					PropertyPath:  "properties.autoUpgradeProfile.upgradeChannel",
					AllowedValues: []string{"stable", "preview"},
				},
			},
			IntRules: []azwise.IntRule{
				// capacity → validation.IntPositive.
				{
					PropertyPath: "properties.scalingProperties.capacity",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "capacity must be a positive integer",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoUpgradeProfile.upgradeChannel", Value: "stable"},
			},
			// capacity ConflictsWith auto_scale_profile (schema lines 161, 168).
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.scalingProperties.capacity", "properties.scalingProperties.autoScaleSettings.profiles"},
					Message: "`capacity` conflicts with `auto_scale_profile`",
				},
				{
					Paths:   []string{"properties.networkProfile.frontEndIPConfiguration.publicIPAddresses", "properties.networkProfile.frontEndIPConfiguration.privateIPAddresses"},
					Message: "`frontend_public` conflicts with `frontend_private`",
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.nginxVersion",
				"properties.ipAddress",
				"properties.dataplaneApiEndpoint",
				"properties.provisioningState",
				"properties.nginxAppProtect.webApplicationFirewallStatus",
			},
		},
	}
}

func init() { azwise.Register(NewNginxDeployment()) }
