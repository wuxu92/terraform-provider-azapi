package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerAppEnvironment provides resource knowledge for
// Microsoft.App/managedEnvironments.
//
// Contributing Terraform resources (both mutate this one ARM type):
//   - azurerm_container_app_environment (the environment itself)
//   - azurerm_container_app_environment_custom_domain (sets
//     properties.customDomainConfiguration on the same body; no distinct ARM type)
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_environment_resource.go
//     schema (Arguments 82-201, Attributes 203-241) + Create (243-345): dapr connection string /
//     infrastructure rg / subnet / internal / zone-redundancy ForceNew, public_network_access enum,
//     mtls & zone-redundant defaults; timeouts 30m/5m/30m/30m.
//   - container_app_environment_custom_domain_resource.go Create (83-151): customDomainConfiguration
//     (dnsSuffix/certificateValue/certificatePassword) set on the environment body.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/managedenvironments
//     ManagedEnvironmentProperties, VnetConfiguration, Mtls, AppLogsConfiguration,
//     CustomDomainConfiguration, constants.go PublicNetworkAccess enum.
type ContainerAppEnvironment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerAppEnvironment)(nil)

func NewContainerAppEnvironment() *ContainerAppEnvironment {
	return &ContainerAppEnvironment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/managedEnvironments",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.daprAIConnectionString"},
				{PropertyPath: "properties.infrastructureResourceGroup"},
				{PropertyPath: "properties.vnetConfiguration.infrastructureSubnetId"},
				{PropertyPath: "properties.vnetConfiguration.internal"},
				{PropertyPath: "properties.zoneRedundant"},
			},
			StringRules: []azwise.StringRule{
				// public_network_access -> PublicNetworkAccess enum.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "public network access must be Disabled or Enabled",
				},
			},
			// dapr_application_insights_connection_string and the custom-domain certificate
			// password are Sensitive.
			SensitiveFields: []string{
				"properties.daprAIConnectionString",
				"properties.customDomainConfiguration.certificatePassword",
			},
			// Read-only properties populated by Azure in the GET response.
			ComputedFields: []string{
				"properties.defaultDomain",
				"properties.staticIp",
				"properties.vnetConfiguration.dockerBridgeCidr",
				"properties.vnetConfiguration.platformReservedCidr",
				"properties.vnetConfiguration.platformReservedDnsIP",
				"properties.customDomainConfiguration.customDomainVerificationId",
				"properties.provisioningState",
			},
			// AzureRM schema defaults; Azure decides public_network_access when omitted.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.zoneRedundant", Value: false},
				{PropertyPath: "properties.vnetConfiguration.internal", Value: false},
				{PropertyPath: "properties.peerAuthentication.mtls.enabled", Value: false},
				{PropertyPath: "properties.publicNetworkAccess"},
			},
		},
	}
}

func init() { azwise.Register(NewContainerAppEnvironment()) }
