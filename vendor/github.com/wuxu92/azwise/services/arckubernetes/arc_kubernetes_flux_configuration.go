package arckubernetes

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcKubernetesFluxConfiguration provides resource knowledge for
// Microsoft.KubernetesConfiguration/fluxConfigurations.
//
// Mirrors azurerm_arc_kubernetes_flux_configuration. The flux configuration is a
// scoped resource whose parent is a connected cluster (azurerm's cluster_id) and
// whose ARM name is azurerm's name; both live on the operational envelope.
//
// Source-kind selection (blob_storage / bucket / git_repository) is ExactlyOneOf
// and maps to the sibling ARM objects properties.azureBlob / properties.bucket /
// properties.gitRepository.
//
// Skipped / non-declarative (noted, not dropped silently):
//   - kustomizations is a Set that ARM stores as a map keyed by name; its nested
//     name/path/timeout/sync/retry/recreating/garbage-collection rules target
//     map-element paths and are not expressible as flat rules.
//   - git_repository.reference_type / reference_value are combined by the expander
//     into properties.gitRepository.repositoryRef (branch/commit/tag/semver); no
//     single ARM body field corresponds, so the StringInSlice on reference_type is
//     omitted.
//   - Secrets that AzureRM routes into properties.configurationProtectedSettings
//     (bucket.secret_key_base64, git_repository.https_key_base64,
//     git_repository.ssh_private_key_base64) have no dedicated body field.
//   - Semantic validators (StorageContainerDataPlaneID, LocalAuthReference,
//     IsURLWithHTTPorHTTPS, KubernetesGitRepositoryUrl, StringIsBase64) have no
//     declarative StringRule equivalent and belong in a customizer.
//   - continuous_reconciliation_enabled and bucket.tls_enabled are INVERTED into
//     properties.suspend and properties.bucket.insecure respectively.
//
// Sources:
//   - terraform-provider-azurerm internal/services/arckubernetes/arc_kubernetes_flux_configuration_resource.go
//     (schema L135-510, create L516-593, expanders L765-970; name regex, namespace
//     regex+ForceNew, scope enum+ForceNew+Default namespace; timeouts 30m/5m/30m/30m)
//   - terraform-provider-azurerm internal/services/containers/kubernetes_flux_configuration_resource.go
//     (AKS-scoped variant; schema L160-604; adds git_repository.provider_type
//     StringInSlice ProviderType at L554-556 -> properties.gitRepository.provider)
//   - go-azure-sdk resource-manager/kubernetesconfiguration/2025-04-01/fluxconfiguration
//     FluxConfigurationProperties (namespace/scope/suspend/kustomizations/azureBlob/
//     bucket/gitRepository settable); ScopeType = cluster|namespace;
//     BucketDefinition.Insecure = !tls_enabled
type ArcKubernetesFluxConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcKubernetesFluxConfiguration)(nil)

// NewArcKubernetesFluxConfiguration returns knowledge for the fluxConfigurations resource.
func NewArcKubernetesFluxConfiguration() *ArcKubernetesFluxConfiguration {
	return &ArcKubernetesFluxConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.KubernetesConfiguration/fluxConfigurations",
			ApiVersions:  []string{"2025-04-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.namespace"},
				{PropertyPath: "properties.scope"},
			},
			StringRules: []azwise.StringRule{
				// name (StringMatch).
				{Regex: "^[a-z\\d]([-a-z\\d]{0,28}[a-z\\d])?$", Message: "name must be between 1 and 30 characters, contain only lowercase letters, numbers and hyphens (-), and start and end with a lowercase letter or number"},
				// namespace (StringMatch).
				{PropertyPath: "properties.namespace", Regex: "^[a-z\\d]([-a-z\\d]{0,61}[a-z\\d])?$", Message: "namespace must be between 1 and 63 characters, contain only lowercase letters, numbers and hyphens (-), and start and end with a lowercase letter or number"},
				// scope (StringInSlice -> ScopeType enum).
				{PropertyPath: "properties.scope", AllowedValues: []string{"cluster", "namespace"}},
				// bucket.bucket_name (StringMatch).
				{PropertyPath: "properties.bucket.bucketName", Regex: "^[a-z\\d]([-a-z\\d]{0,61}[a-z\\d])?$", Message: "bucket_name must be between 1 and 63 characters, contain only lowercase letters, numbers and hyphens (-), and start and end with a lowercase letter or number"},
				// git_repository.provider_type (StringInSlice -> ProviderType enum). Unioned
				// from azurerm_kubernetes_flux_configuration; the arc variant omits it. Fires
				// only when git_repository is configured, so it is safe for both bodies.
				{PropertyPath: "properties.gitRepository.provider", AllowedValues: []string{"Azure", "Generic", "GitHub"}},
			},
			IntRules: []azwise.IntRule{
				// blob_storage.sync_interval_in_seconds / timeout_in_seconds (IntBetween).
				{PropertyPath: "properties.azureBlob.syncIntervalInSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(35791394))},
				{PropertyPath: "properties.azureBlob.timeoutInSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(35791394))},
				// bucket.sync_interval_in_seconds / timeout_in_seconds (IntBetween).
				{PropertyPath: "properties.bucket.syncIntervalInSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(35791394))},
				{PropertyPath: "properties.bucket.timeoutInSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(35791394))},
			},
			SensitiveFields: []string{
				"properties.azureBlob.accountKey",
				"properties.azureBlob.sasToken",
				"properties.azureBlob.servicePrincipal.clientCertificate",
				"properties.azureBlob.servicePrincipal.clientCertificatePassword",
				"properties.azureBlob.servicePrincipal.clientSecret",
				"properties.gitRepository.httpsCACert",
				"properties.configurationProtectedSettings",
			},
			DefaultValues: []azwise.DefaultValue{
				// scope Default namespace.
				{PropertyPath: "properties.scope", Value: "namespace"},
				// continuous_reconciliation_enabled Default true -> suspend false (inverted).
				{PropertyPath: "properties.suspend", Value: false},
				// bucket.tls_enabled Default true -> insecure false (inverted).
				{PropertyPath: "properties.bucket.insecure", Value: false},
				// sync/timeout Default 600 for the single-object source blocks.
				{PropertyPath: "properties.azureBlob.syncIntervalInSeconds", Value: float64(600)},
				{PropertyPath: "properties.azureBlob.timeoutInSeconds", Value: float64(600)},
				{PropertyPath: "properties.bucket.syncIntervalInSeconds", Value: float64(600)},
				{PropertyPath: "properties.bucket.timeoutInSeconds", Value: float64(600)},
				{PropertyPath: "properties.gitRepository.syncIntervalInSeconds", Value: float64(600)},
				{PropertyPath: "properties.gitRepository.timeoutInSeconds", Value: float64(600)},
			},
			// Exactly one source kind must be configured.
			ExactlyOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.azureBlob", "properties.bucket", "properties.gitRepository"}},
			},
		},
	}
}

func init() { azwise.Register(NewArcKubernetesFluxConfiguration()) }
