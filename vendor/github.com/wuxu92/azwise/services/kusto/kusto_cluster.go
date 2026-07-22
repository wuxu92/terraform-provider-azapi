package kusto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KustoCluster provides resource knowledge for Microsoft.Kusto/clusters.
//
// Mirrors azurerm_kusto_cluster. name / resource_group_name / location live on the
// operational envelope; name and location are still surfaced here as ForceNew because
// AzureRM marks them RequiresReplace and users benefit from the plan-time guardrail.
//
// Sources:
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_resource.go:34-332
//     (resourceKustoCluster schema: ForceNew, Required, validators, defaults, timeouts)
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_resource.go:334-456
//     (create mapping to clusters.Cluster / clusters.ClusterProperties / clusters.AzureSku)
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_resource.go:653-730
//     (read/flatten mapping: uri, data_ingestion_uri, publicNetworkAccess, restrictOutbound)
//   - terraform-provider-azurerm internal/services/kusto/kusto_cluster_resource.go:788-831
//     (expandKustoClusterSku tier derivation; expandKustoClusterVNET ForceNew subfields)
//   - terraform-provider-azurerm internal/services/kusto/validate/name.go:47-59
//     (ClusterName regex + 4..22 length)
//   - go-azure-sdk resource-manager/kusto/2024-04-13/clusters:
//     model_cluster.go:12-24 (envelope: location, sku, zones, identity),
//     model_clusterproperties.go:6-34 (properties.* body paths),
//     model_azuresku.go:6-10 (sku.name/sku.tier/sku.capacity),
//     model_optimizedautoscale.go:6-11 (optimizedAutoscale.minimum/maximum),
//     model_virtualnetworkconfiguration.go:6-11 (deprecated VNET subfields),
//     constants.go:56-192 (AzureSkuName), 283-295 (AzureSkuTier), 515-527 (EngineType),
//     391-404 (ClusterNetworkAccessFlag), 791-844 (PublicIPType, PublicNetworkAccess)
//
// Not encoded (deliberate):
//   - trusted_external_tenants (properties.trustedExternalTenants) and its per-element
//     validator (IsUUID | "" | "*"), allowed_ip_ranges (properties.allowedIpRangeList)
//     and allowed_fqdns (properties.allowedFqdnList) with per-element StringIsNotEmpty:
//     these are array-element constraints. azwise/azapin cannot lower or resolve a path
//     through an array element, so they are documented, not emitted.
//   - language_extension / language_extensions ConflictsWith (kusto_cluster_resource.go:234,255):
//     both are 4.x-only deprecated representations of the SAME ARM field
//     (properties.languageExtensions.value[*]); a RelationalRule needs two distinct ARM
//     paths, so a same-path "mutually exclusive" cannot be expressed. Their element enums
//     (LanguageExtensionName / LanguageExtensionImageName) are also array-element rules.
//   - optimized_auto_scale minimum <= maximum (kusto_cluster_resource.go:374-376): a
//     value-conditional comparison between two sibling ints, not a presence relation. The
//     RelationalRule kinds only test whether a path is set (both are Required inside the
//     block, so a presence relation is trivially satisfied). Left to a customizer/hook.
//   - virtual_network_configuration.{subnet_id,engine_public_ip_id,data_management_public_ip_id}
//     (properties.virtualNetworkConfiguration.*) are ForceNew but the whole block is
//     deprecated (Azure removed VNet injection for ADX) and 4.x-only; not emitted.
//   - identity is a system/user-assigned block (commonschema) with no cross-property
//     constraints to encode.
//   - uri / data_ingestion_uri are server-assigned read-only endpoints; see ComputedFields.
//   - customer_managed_key (azurerm_kusto_cluster_customer_managed_key,
//     kusto_cluster_customer_managed_key_resource.go:27-93) is a SEPARATE TF resource
//     that mutates THIS cluster's body at properties.keyVaultProperties
//     (keyName/keyVaultUri/keyVersion/userIdentity — clusters model_keyvaultproperties.go:6-11).
//     Its inputs are a composite Key Vault / Managed HSM key ID that ARM splits into
//     keyName + keyVaultUri + keyVersion, plus semantic ID validators
//     (ValidateKeyVaultID, ManagedHSMDataPlane*KeyID, ValidateUserAssignedIdentityID)
//     and an ExactlyOneOf(key_vault_id, managed_hsm_key_id). None of these lower to a
//     single ARM body field with a declarative enum/regex/length constraint (the key
//     values are non-mappable composites), so no rule is emitted; the CMK knowledge is
//     documented here rather than in a separate file.
type KustoCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KustoCluster)(nil)

// NewKustoCluster returns knowledge for the Kusto clusters resource.
func NewKustoCluster() *KustoCluster {
	return &KustoCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kusto/clusters",
			ApiVersions:  []string{"2025-02-14"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.enableDoubleEncryption"},
			},
			RequiredFields: []string{
				// sku is Required; AzureRM derives sku.tier from the sku.name prefix
				// (Dev(No SLA)->Basic, Standard->Standard) but ARM requires it explicitly.
				"sku.name",
				"sku.tier",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// validate.ClusterName: globally unique, must begin with a letter,
					// lowercase alphanumeric + dashes, 4..22 chars.
					Regex:     `^[a-z][a-z0-9\-]+$`,
					MinLength: 4,
					MaxLength: 22,
					Message:   "must be globally unique, begin with a lowercase letter, contain only lowercase alphanumeric characters and dashes, and be 4-22 characters long",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Dev(No SLA)_Standard_D11_v2",
						"Dev(No SLA)_Standard_E2a_v4",
						"Standard_D14_v2",
						"Standard_D11_v2",
						"Standard_D16d_v5",
						"Standard_D13_v2",
						"Standard_D12_v2",
						"Standard_DS14_v2+4TB_PS",
						"Standard_DS14_v2+3TB_PS",
						"Standard_DS13_v2+1TB_PS",
						"Standard_DS13_v2+2TB_PS",
						"Standard_D32d_v5",
						"Standard_D32d_v4",
						"Standard_EC8ads_v5",
						"Standard_EC8as_v5+1TB_PS",
						"Standard_EC8as_v5+2TB_PS",
						"Standard_EC16ads_v5",
						"Standard_EC16as_v5+4TB_PS",
						"Standard_EC16as_v5+3TB_PS",
						"Standard_E80ids_v4",
						"Standard_E8a_v4",
						"Standard_E8ads_v5",
						"Standard_E8as_v5+1TB_PS",
						"Standard_E8as_v5+2TB_PS",
						"Standard_E8as_v4+1TB_PS",
						"Standard_E8as_v4+2TB_PS",
						"Standard_E8d_v5",
						"Standard_E8d_v4",
						"Standard_E8s_v5+1TB_PS",
						"Standard_E8s_v5+2TB_PS",
						"Standard_E8s_v4+1TB_PS",
						"Standard_E8s_v4+2TB_PS",
						"Standard_E4a_v4",
						"Standard_E4ads_v5",
						"Standard_E4d_v5",
						"Standard_E4d_v4",
						"Standard_E16a_v4",
						"Standard_E16ads_v5",
						"Standard_E16as_v5+4TB_PS",
						"Standard_E16as_v5+3TB_PS",
						"Standard_E16as_v4+4TB_PS",
						"Standard_E16as_v4+3TB_PS",
						"Standard_E16d_v5",
						"Standard_E16d_v4",
						"Standard_E16s_v5+4TB_PS",
						"Standard_E16s_v5+3TB_PS",
						"Standard_E16s_v4+4TB_PS",
						"Standard_E16s_v4+3TB_PS",
						"Standard_E64i_v3",
						"Standard_E2a_v4",
						"Standard_E2ads_v5",
						"Standard_E2d_v5",
						"Standard_E2d_v4",
						"Standard_L8as_v3",
						"Standard_L8s",
						"Standard_L8s_v3",
						"Standard_L8s_v2",
						"Standard_L4s",
						"Standard_L16as_v3",
						"Standard_L16s",
						"Standard_L16s_v3",
						"Standard_L16s_v2",
						"Standard_L32as_v3",
						"Standard_L32s_v3",
					},
					Message: "must be a valid Kusto cluster SKU name",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Basic", "Standard"},
					Message:       "must be Basic or Standard",
				},
				{
					PropertyPath:  "properties.publicIPType",
					AllowedValues: []string{"DualStack", "IPv4"},
					Message:       "must be DualStack or IPv4",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.restrictOutboundNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					// Not exposed by AzureRM, but a real ARM enum body field users may set.
					PropertyPath:  "properties.engineType",
					AllowedValues: []string{"V3", "V2"},
					Message:       "must be V2 or V3",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "sku.capacity",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(1000)),
					Message:      "sku capacity must be between 1 and 1000",
				},
				{
					PropertyPath: "properties.optimizedAutoscale.minimum",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(1000)),
					Message:      "optimized autoscale minimum instances must be between 0 and 1000",
				},
				{
					PropertyPath: "properties.optimizedAutoscale.maximum",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(1000)),
					Message:      "optimized autoscale maximum instances must be between 0 and 1000",
				},
			},
			ComputedFields: []string{
				// Server-assigned read-only endpoints (AzureRM Computed-only).
				"properties.uri",
				"properties.dataIngestionUri",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.publicIPType", Value: "IPv4"},
				{PropertyPath: "properties.restrictOutboundNetworkAccess", Value: "Disabled"},
				{PropertyPath: "properties.enableAutoStop", Value: true},
				{PropertyPath: "properties.enableDiskEncryption", Value: false},
				{PropertyPath: "properties.enableStreamingIngest", Value: false},
				{PropertyPath: "properties.enablePurge", Value: false},
				{PropertyPath: "sku.capacity"}, // Optional+Computed: server decides
			},
		},
	}
}

func init() { azwise.Register(NewKustoCluster()) }
