package redhatopenshift

import (
	"time"

	"github.com/wuxu92/azwise"
)

// OpenShiftCluster provides resource knowledge for Microsoft.RedHatOpenShift/openShiftClusters.
//
// Mirrors azurerm_redhat_openshift_cluster. name / location / resource_group_name live on the
// operational envelope; name and location are still surfaced here as ForceNew because AzureRM
// marks them RequiresReplace and users benefit from the plan-time guardrail. ARM type casing
// ("Microsoft.RedHatOpenShift" / "openShiftClusters") verified against the SDK id parser.
//
// Sources:
//   - terraform-provider-azurerm internal/services/redhatopenshift/redhat_openshift_cluster_resource.go:95-357
//     (Arguments schema: ForceNew, Required, validators, defaults)
//   - terraform-provider-azurerm internal/services/redhatopenshift/redhat_openshift_cluster_resource.go:359-366
//     (Attributes: console_url computed)
//   - terraform-provider-azurerm internal/services/redhatopenshift/redhat_openshift_cluster_resource.go:380-551
//     (Create/Read/Delete timeouts: 90m/5m/90m; create mapping to OpenShiftClusterProperties)
//   - terraform-provider-azurerm internal/services/redhatopenshift/redhat_openshift_cluster_resource.go:553-703
//     (expand: managed_resource_group_name->resourceGroupId, fips_enabled->fipsValidatedModules,
//     preconfigured_network_security_group_enabled->preconfiguredNSG, encryption_at_host_enabled->encryptionAtHost)
//   - go-azure-sdk resource-manager/redhatopenshift/2025-07-25/openshiftclusters:
//     id_openshiftcluster.go:109-120 (segment casing),
//     model_openshiftcluster.go:11-20 (envelope: location, tags, identity),
//     model_openshiftclusterproperties.go:6-18 (properties.* body paths),
//     model_clusterprofile.go:6-13, model_networkprofile.go:6-12, model_masterprofile.go:6-11,
//     model_workerprofile.go:6-14, model_apiserverprofile.go:6-10, model_ingressprofile.go:6-10,
//     model_consoleprofile.go:6-8, model_serviceprincipalprofile.go:6-9,
//     constants.go:12-17 (EncryptionAtHost), 53-58 (FipsValidatedModules), 94-99 (OutboundType),
//     135-140 (PreconfiguredNSG), 232-237 (Visibility)
//
// Not encoded (deliberate):
//   - worker_profile.{vm_size,disk_size_gb,node_count,subnet_id,encryption_at_host_enabled,
//     disk_encryption_set_id} map into properties.workerProfiles[*].* — an array element.
//     azwise cannot lower or resolve a path through an array element, so its ForceNew,
//     Required, IntAtLeast(128)/IntBetween(3,60) and EncryptionAtHost enum are documented,
//     not emitted.
//   - ingress_profile.visibility maps into properties.ingressProfiles[*].visibility — also an
//     array element; its Visibility enum and ForceNew/Required are not emitted. ip / name are
//     server-assigned read-only fields on that same array element.
//   - service_principal.client_id (validation.IsUUID) is a generic semantic validator; it is a
//     scalar body field (properties.servicePrincipalProfile.clientId) but the UUID check is
//     semantic, not declarative — left to a customizer/hook.
type OpenShiftCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*OpenShiftCluster)(nil)

// NewOpenShiftCluster returns knowledge for the Red Hat OpenShift cluster resource.
func NewOpenShiftCluster() *OpenShiftCluster {
	return &OpenShiftCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RedHatOpenShift/openShiftClusters",
			ApiVersions:  []string{"2025-07-25"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// cluster_profile block (ForceNew) leaves:
				{PropertyPath: "properties.clusterProfile.domain"},
				{PropertyPath: "properties.clusterProfile.version"},
				{PropertyPath: "properties.clusterProfile.fipsValidatedModules"},
				{PropertyPath: "properties.clusterProfile.pullSecret"},
				{PropertyPath: "properties.clusterProfile.resourceGroupId"},
				// network_profile block (ForceNew) leaves:
				{PropertyPath: "properties.networkProfile.podCidr"},
				{PropertyPath: "properties.networkProfile.serviceCidr"},
				{PropertyPath: "properties.networkProfile.outboundType"},
				{PropertyPath: "properties.networkProfile.preconfiguredNSG"},
				// master_profile block (ForceNew) leaves:
				{PropertyPath: "properties.masterProfile.subnetId"},
				{PropertyPath: "properties.masterProfile.vmSize"},
				{PropertyPath: "properties.masterProfile.encryptionAtHost"},
				{PropertyPath: "properties.masterProfile.diskEncryptionSetId"},
				// api_server_profile block (ForceNew) leaf:
				{PropertyPath: "properties.apiserverProfile.visibility"},
			},
			RequiredFields: []string{
				"properties.clusterProfile.domain",
				"properties.clusterProfile.version",
				"properties.servicePrincipalProfile.clientId",
				"properties.servicePrincipalProfile.clientSecret",
				"properties.networkProfile.podCidr",
				"properties.networkProfile.serviceCidr",
				"properties.masterProfile.subnetId",
				"properties.masterProfile.vmSize",
				"properties.apiserverProfile.visibility",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.clusterProfile.fipsValidatedModules",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.networkProfile.outboundType",
					AllowedValues: []string{"Loadbalancer", "UserDefinedRouting"},
					Message:       "must be Loadbalancer or UserDefinedRouting",
				},
				{
					PropertyPath:  "properties.networkProfile.preconfiguredNSG",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.masterProfile.encryptionAtHost",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.apiserverProfile.visibility",
					AllowedValues: []string{"Private", "Public"},
					Message:       "must be Private or Public",
				},
			},
			SensitiveFields: []string{
				"properties.clusterProfile.pullSecret",
				"properties.servicePrincipalProfile.clientSecret",
			},
			ComputedFields: []string{
				// Server-assigned read-only fields (AzureRM Computed-only).
				"properties.clusterProfile.oidcIssuer",
				"properties.apiserverProfile.ip",
				"properties.apiserverProfile.url",
				"properties.consoleProfile.url",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.clusterProfile.fipsValidatedModules", Value: "Disabled"},
				{PropertyPath: "properties.networkProfile.outboundType", Value: "Loadbalancer"},
				{PropertyPath: "properties.networkProfile.preconfiguredNSG", Value: "Disabled"},
				{PropertyPath: "properties.masterProfile.encryptionAtHost", Value: "Disabled"},
			},
		},
	}
}

func init() { azwise.Register(NewOpenShiftCluster()) }
