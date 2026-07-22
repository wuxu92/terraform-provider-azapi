package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerApp provides resource knowledge for Microsoft.App/containerApps.
//
// Mirrors azurerm_container_app.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_resource.go
//     schema (Arguments 69-124, Attributes) + Create (159-241): name/environment ForceNew,
//     revision_mode enum, max_inactive_revisions range, managedEnvironmentId mapping;
//     timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/containerapps
//     ContainerAppProperties (configuration, managedEnvironmentId, template, workloadProfileName,
//     read-only customDomainVerificationId/outboundIpAddresses/latestRevision*), Configuration
//     (activeRevisionsMode, maxInactiveRevisions), constants.go ActiveRevisionsMode enum.
//
// NOTE: the container app's ingress custom domains are managed by the separate
// azurerm_container_app_custom_domain resource, which mutates this same
// Microsoft.App/containerApps body (properties.configuration.ingress.customDomains)
// and has no distinct ARM resource type.
//
// TODO: the template/ingress/dapr/registry/secret nested blocks are extensive; only
// the top-level scalar constraints are captured declaratively here.
type ContainerApp struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerApp)(nil)

func NewContainerApp() *ContainerApp {
	return &ContainerApp{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/containerApps",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// container_app_environment_id is ForceNew and maps to managedEnvironmentId
			// in the create body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.managedEnvironmentId"},
			},
			RequiredFields: []string{
				"properties.managedEnvironmentId",
				"properties.configuration.activeRevisionsMode",
				"properties.template",
			},
			StringRules: []azwise.StringRule{
				// revision_mode -> ActiveRevisionsMode enum.
				{
					PropertyPath:  "properties.configuration.activeRevisionsMode",
					AllowedValues: []string{"Multiple", "Single"},
					Message:       "revision mode must be Multiple or Single",
				},
			},
			IntRules: []azwise.IntRule{
				// max_inactive_revisions -> validation.IntBetween(0, 100).
				{
					PropertyPath: "properties.configuration.maxInactiveRevisions",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(100)),
					Message:      "max inactive revisions must be between 0 and 100",
				},
			},
			// Read-only properties populated by Azure in the GET response.
			ComputedFields: []string{
				"properties.customDomainVerificationId",
				"properties.outboundIpAddresses",
				"properties.latestRevisionName",
				"properties.latestRevisionFqdn",
				"properties.latestReadyRevisionName",
				"properties.eventStreamEndpoint",
				"properties.provisioningState",
				"properties.runningStatus",
			},
		},
	}
}

func init() { azwise.Register(NewContainerApp()) }
