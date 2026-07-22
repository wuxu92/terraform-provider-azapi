package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerAppEnvironmentDaprComponent provides resource knowledge for
// Microsoft.App/managedEnvironments/daprComponents.
//
// Mirrors azurerm_container_app_environment_dapr_component.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_environment_dapr_component_resource.go
//     schema (Arguments 49-111) + Create (117-174): name/environment/component_type ForceNew,
//     init_timeout & ignore_errors defaults, version required; timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/daprcomponents
//     DaprComponentProperties (componentType/version/initTimeout/ignoreErrors/metadata/scopes/secrets,
//     read-only provisioningState/deploymentErrors).
//
// NOTE: init_timeout uses the custom validate.InitTimeout (ISO8601 duration) validator; it is
// not a declarative enum/regex/range and is not expressed as a StringRule here.
type ContainerAppEnvironmentDaprComponent struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerAppEnvironmentDaprComponent)(nil)

func NewContainerAppEnvironmentDaprComponent() *ContainerAppEnvironmentDaprComponent {
	return &ContainerAppEnvironmentDaprComponent{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/managedEnvironments/daprComponents",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.componentType"},
			},
			RequiredFields: []string{
				"properties.componentType",
				"properties.version",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.deploymentErrors",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.initTimeout", Value: "5s"},
				{PropertyPath: "properties.ignoreErrors", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewContainerAppEnvironmentDaprComponent()) }
