package arckubernetes

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcKubernetesClusterExtension provides resource knowledge for
// Microsoft.KubernetesConfiguration/extensions.
//
// Mirrors azurerm_arc_kubernetes_cluster_extension. The extension is a scoped
// resource whose parent is a connected cluster (azurerm's cluster_id) and whose
// ARM name is azurerm's name; both live on the operational envelope. identity is a
// system-assigned identity on the envelope (Required + ForceNew) and is not
// repeated as a body rule.
//
// Notes:
//   - release_namespace and target_namespace map to distinct ARM paths under
//     properties.scope and are mutually exclusive (AzureRM ConflictsWith).
//   - configuration_settings / configuration_protected_settings are maps; the
//     per-element StringIsNotEmpty validators target map values and are skipped.
//   - autoUpgradeMinorVersion is derived by AzureRM (version == "") and has no
//     schema field, so no rule is emitted for it.
//   - release_train / release_namespace / target_namespace are Optional+Computed
//     (server may default); represented as nil-valued DefaultValues.
//   - plan (Marketplace plan) is a top-level ForceNew block that only the AKS-scoped
//     resource exposes; its name/product/publisher/promotion_code/version map to the
//     plan.* object and are unioned as ForceNew (they only fire when plan is set, so
//     they never affect an arc-connected-cluster body that omits plan).
//
// Sources:
//   - terraform-provider-azurerm internal/services/arckubernetes/arc_kubernetes_cluster_extension_resource.go
//     (schema L72-152, create L163-247; name regex ^[a-zA-Z0-9][a-zA-Z0-9-.]{0,252}$;
//     extension_type/release_train/release_namespace/target_namespace/version ForceNew;
//     configuration_protected_settings Sensitive; timeouts 30m/5m/30m/30m)
//   - terraform-provider-azurerm internal/services/containers/kubernetes_cluster_extension_resource.go
//     (AKS-scoped variant; schema L81-206; adds plan block L127-168 name/product/publisher
//     Required+ForceNew, promotion_code/version Optional+ForceNew)
//   - go-azure-sdk resource-manager/kubernetesconfiguration/2024-11-01/extensions
//     ExtensionProperties (extensionType/releaseTrain/version/configurationSettings/
//     configurationProtectedSettings/scope settable; currentVersion server-computed);
//     Scope -> {cluster.releaseNamespace, namespace.targetNamespace};
//     Extension.Plan -> {name, product, publisher, promotionCode, version}
type ArcKubernetesClusterExtension struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcKubernetesClusterExtension)(nil)

// NewArcKubernetesClusterExtension returns knowledge for the extensions resource.
func NewArcKubernetesClusterExtension() *ArcKubernetesClusterExtension {
	return &ArcKubernetesClusterExtension{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.KubernetesConfiguration/extensions",
			ApiVersions:  []string{"2024-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.extensionType"},
				{PropertyPath: "properties.releaseTrain"},
				{PropertyPath: "properties.scope.cluster.releaseNamespace"},
				{PropertyPath: "properties.scope.namespace.targetNamespace"},
				{PropertyPath: "properties.version"},
				// plan (Marketplace) block, ForceNew. Unioned from azurerm_kubernetes_cluster_extension;
				// only fires when a plan is set, so it never affects the arc variant.
				{PropertyPath: "plan.name"},
				{PropertyPath: "plan.product"},
				{PropertyPath: "plan.publisher"},
				{PropertyPath: "plan.promotionCode"},
				{PropertyPath: "plan.version"},
			},
			StringRules: []azwise.StringRule{
				// name (StringMatch).
				{Regex: "^[a-zA-Z0-9][a-zA-Z0-9-.]{0,252}$", Message: "name must be between 1 and 253 characters and may contain only letters, numbers, periods (.), hyphens (-), and must begin with a letter or number"},
				// extension_type (StringIsNotEmpty).
				{PropertyPath: "properties.extensionType", MinLength: 1, Message: "extension_type must not be empty"},
				// release_train (StringIsNotEmpty).
				{PropertyPath: "properties.releaseTrain", MinLength: 1, Message: "release_train must not be empty"},
				// version (StringIsNotEmpty).
				{PropertyPath: "properties.version", MinLength: 1, Message: "version must not be empty"},
			},
			SensitiveFields: []string{
				"properties.configurationProtectedSettings",
			},
			DefaultValues: []azwise.DefaultValue{
				// Optional+Computed: server supplies a value when omitted.
				{PropertyPath: "properties.releaseTrain", Value: nil},
				{PropertyPath: "properties.scope.cluster.releaseNamespace", Value: nil},
				{PropertyPath: "properties.scope.namespace.targetNamespace", Value: nil},
			},
			ComputedFields: []string{
				"properties.currentVersion",
			},
			// release_namespace <-> target_namespace are mutually exclusive.
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.scope.cluster.releaseNamespace", "properties.scope.namespace.targetNamespace"}},
			},
		},
	}
}

func init() { azwise.Register(NewArcKubernetesClusterExtension()) }
