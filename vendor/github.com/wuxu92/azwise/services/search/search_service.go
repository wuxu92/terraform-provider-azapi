package search

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SearchService provides resource knowledge for Microsoft.Search/searchServices.
//
// Mirrors azurerm_search_service. The admin keys (primary_key / secondary_key)
// and query_keys are read-only data-plane outputs fetched via the separate
// adminkeys / querykeys APIs, NOT part of the service body, so they are not
// modelled as sensitive body fields here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/search/search_service_resource.go:32-219
//     (schema), :221-362 (create -> services.SearchService), :39-44 (timeouts).
//   - CustomizeDiff (:51-54, validateSearchServiceSKUUpdate / validateSearchServiceApiAccessControlRbac)
//     enforces SKU-downgrade replacement and RBAC-mode consistency; neither is a
//     declarative single-path ForceNew, so they are documented but not emitted.
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     allowed_ips (validate.IPv4Address / validate.CIDR -> properties.networkRuleSet.ipRules[*].value,
//     an array-element path, skipped).
//   - go-azure-sdk resource-manager/search/2025-05-01/services model_searchservice.go /
//     model_searchserviceproperties.go / model_sku.go / model_networkruleset.go /
//     model_encryptionwithcmk.go / model_dataplaneauthoptions.go /
//     model_dataplaneaadorapikeyauthoption.go / constants.go.
type SearchService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SearchService)(nil)

// NewSearchService returns knowledge for the Azure AI Search service resource.
func NewSearchService() *SearchService {
	return &SearchService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Search/searchServices",
			ApiVersions:  []string{"2025-05-01"},
			ForceNew: []azwise.ForceNewRule{
				// hosting_mode (ForceNew) -> properties.hostingMode.
				{PropertyPath: "properties.hostingMode"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			// sku is Required; location/name are envelope-owned.
			RequiredFields: []string{
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				// replica_count Default 1.
				{PropertyPath: "properties.replicaCount", Value: 1},
				// partition_count Default 1.
				{PropertyPath: "properties.partitionCount", Value: 1},
				// local_authentication_enabled Default true maps to the inverse
				// disableLocalAuth = !true = false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				// hosting_mode Default "Default".
				{PropertyPath: "properties.hostingMode", Value: "Default"},
				// customer_managed_key_enforcement_enabled Default false ->
				// encryptionWithCmk.enforcement "Disabled".
				{PropertyPath: "properties.encryptionWithCmk.enforcement", Value: "Disabled"},
				// public_network_access_enabled Default true -> "Enabled".
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// network_rule_bypass_option Default "None".
				{PropertyPath: "properties.networkRuleSet.bypass", Value: "None"},
			},
			StringRules: []azwise.StringRule{
				// sku: full ARM SkuName enum.
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"free", "basic", "standard", "standard2", "standard3",
						"storage_optimized_l1", "storage_optimized_l2",
					},
					Message: "sku must be a valid Search SkuName",
				},
				// authentication_failure_mode: AadAuthFailureMode enum.
				{
					PropertyPath:  "properties.authOptions.aadOrApiKey.aadAuthFailureMode",
					AllowedValues: []string{"http401WithBearerChallenge", "http403"},
					Message:       "authentication_failure_mode must be http401WithBearerChallenge or http403",
				},
				// hosting_mode: HostingMode enum.
				{
					PropertyPath:  "properties.hostingMode",
					AllowedValues: []string{"Default", "HighDensity"},
					Message:       "hosting_mode must be Default or HighDensity",
				},
				// customer_managed_key_enforcement: SearchEncryptionWithCmk enum.
				{
					PropertyPath:  "properties.encryptionWithCmk.enforcement",
					AllowedValues: []string{"Disabled", "Enabled", "Unspecified"},
					Message:       "encryptionWithCmk enforcement must be Disabled, Enabled or Unspecified",
				},
				// public_network_access: full ARM PublicNetworkAccess enum (AzureRM
				// maps its bool to Enabled/Disabled, ARM also accepts SecuredByPerimeter).
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled", "SecuredByPerimeter"},
					Message:       "publicNetworkAccess must be Enabled, Disabled or SecuredByPerimeter",
				},
				// semantic_search_sku: full ARM SearchSemanticSearch enum (AzureRM
				// restricts to free/standard, ARM also accepts disabled).
				{
					PropertyPath:  "properties.semanticSearch",
					AllowedValues: []string{"disabled", "free", "standard"},
					Message:       "semantic_search_sku must be disabled, free or standard",
				},
				// network_rule_bypass_option: SearchBypass enum.
				{
					PropertyPath:  "properties.networkRuleSet.bypass",
					AllowedValues: []string{"AzureServices", "None"},
					Message:       "network_rule_bypass_option must be AzureServices or None",
				},
			},
			IntRules: []azwise.IntRule{
				// replica_count: IntBetween(1, 12) -> properties.replicaCount.
				{
					PropertyPath: "properties.replicaCount",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](12),
					Message:      "replica_count must be between 1 and 12",
				},
				// partition_count: AzureRM uses IntInSlice(1,2,3,4,6,12), a discrete set
				// (not a contiguous range). The bound rule below rejects <1 and >12; the
				// intermediate invalid values (5, 7-11) cannot be expressed with the
				// available range-only IntRule and are left unvalidated.
				// TODO: represent discrete-value int constraints when azwise supports them.
				{
					PropertyPath: "properties.partitionCount",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](12),
					Message:      "partition_count must be one of 1, 2, 3, 4, 6 or 12",
				},
			},
			// Read-only body properties: present in SearchServiceProperties (GET)
			// but server-computed, not user-settable inputs.
			ComputedFields: []string{
				"properties.endpoint",
				"properties.encryptionWithCmk.encryptionComplianceStatus",
				"properties.provisioningState",
				"properties.status",
				"properties.statusDetails",
				"properties.privateEndpointConnections",
				"properties.sharedPrivateLinkResources",
			},
		},
	}
}

func init() { azwise.Register(NewSearchService()) }
